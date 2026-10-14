package extractors

import (
	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// resolveGRPCCalls emits service consumes contract:<X> for every method call on
// a field or variable bound to an <X> client, directly or through local aliases,
// anywhere in the service's Go code. Building a client or declaring the field is
// not enough (spec gira-impact-precision, cause 3).
func resolveGRPCCalls(r *resolution, w *wiring) {
	clients := map[string]map[string][]string{} // service → name → contracts
	aliases := map[string]map[string][]string{} // service → name → names it was assigned from
	for _, h := range w.hints {
		switch h.Kind {
		case HintGRPCClient:
			if clients[h.Service] == nil {
				clients[h.Service] = map[string][]string{}
			}
			clients[h.Service][h.Key] = appendNew(clients[h.Service][h.Key], h.Value)
		case HintGoAlias:
			if aliases[h.Service] == nil {
				aliases[h.Service] = map[string][]string{}
			}
			aliases[h.Service][h.Key] = appendNew(aliases[h.Service][h.Key], h.Ref)
		}
	}
	var contractsOf func(service, name string, depth int) []string
	contractsOf = func(service, name string, depth int) []string {
		out := append([]string(nil), clients[service][name]...)
		if depth < 4 {
			for _, ref := range aliases[service][name] {
				for _, c := range contractsOf(service, ref, depth+1) {
					out = appendNew(out, c)
				}
			}
		}
		return out
	}
	for _, h := range w.hints {
		if h.Kind != HintGoMethodCall {
			continue
		}
		for _, contract := range contractsOf(h.Service, h.Key, 0) {
			id := "contract:" + contract
			if _, ok := r.nodes[id]; !ok {
				r.nodes[id] = &domain.Node{ID: id, Kind: domain.NodeContract, Name: contract}
			}
			r.addEdge(h.Service, domain.VerbConsumes, id, evidenceAt(h.Locator, domain.OriginExtracted, "grpc"))
		}
	}
}

func appendNew(xs []string, x string) []string {
	for _, have := range xs {
		if have == x {
			return xs
		}
	}
	return append(xs, x)
}
