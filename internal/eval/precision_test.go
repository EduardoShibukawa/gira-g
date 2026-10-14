package eval_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/eval"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

// variant is one hermetic variant indexed through the real pipeline, with its
// ground truth (spec gira-impact-precision).
type variant struct {
	name string
	dir  string // the corpus source tree
	gt   *eval.GroundTruth
	mm   *domain.MacroMap
}

func hermeticVariants(t *testing.T) []variant {
	t.Helper()
	gt := loadPublished(t)
	pub, err := publishedCorpus(gt).Index(context.Background(), t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	ps := perturbedSuite(t)
	per, err := ps.Corpus.Index(context.Background(), t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	return []variant{{"published", publishedCorpus(gt).Dir, gt, pub.MacroMap}, {"perturbed", ps.Corpus.Dir, ps.GroundTruth, per.MacroMap}}
}

func isDockerfile(p string) bool {
	b := strings.ToLower(path.Base(p))
	return b == "dockerfile" || strings.HasPrefix(b, "dockerfile.") || strings.HasSuffix(b, ".dockerfile")
}

// SC1: Dockerfile stages and base images are not nodes, briefings list only
// Services and Infrastructure as upstream and downstream, and Libraries reached
// through imports are listed in upstream_libraries.
func TestSC1ImpactSetHoldsOnlyServices(t *testing.T) {
	for _, v := range hermeticVariants(t) {
		for _, n := range v.mm.Nodes {
			if strings.HasPrefix(n.ID, "service:stage:") {
				t.Errorf("%s: Dockerfile stage emitted as node %s", v.name, n.ID)
			}
			if n.Kind == domain.NodeService {
				continue
			}
			for _, ev := range n.Evidence {
				if isDockerfile(ev.Locator.Path) {
					t.Errorf("%s: Dockerfile base image emitted as %s node %s", v.name, n.Kind, n.ID)
					break
				}
			}
		}

		engine := slices.NewEngine(v.mm)
		libraries := 0
		for _, task := range v.gt.Tasks {
			b, err := engine.GetImpactBriefing(task.TargetID, task.Depth)
			if err != nil {
				continue // a target the index does not know is cause 2's (SC3)
			}
			for field, group := range map[string][]domain.Node{"upstream_services": b.UpstreamServices, "downstream_services": b.DownstreamServices} {
				for _, n := range group {
					if n.Kind != domain.NodeService && n.Kind != domain.NodeInfrastructure {
						t.Errorf("%s %s: %s holds %s node %s", v.name, task.ID, field, n.Kind, n.ID)
					}
				}
			}
			for _, n := range b.UpstreamLibraries {
				if n.Kind != domain.NodeLibrary {
					t.Errorf("%s %s: upstream_libraries holds %s node %s", v.name, task.ID, n.Kind, n.ID)
				}
			}
			libraries += len(b.UpstreamLibraries)
		}
		if libraries == 0 {
			t.Errorf("%s: no briefing lists upstream_libraries; the Go imports must still be reported", v.name)
		}
	}
}

// SC2: a Dockerfile a compose service builds is attributed to that compose service.
// In the hermetic corpus every Dockerfile is compose-built, so every Service with
// Dockerfile evidence must also be declared in a compose file, and cart's
// Dockerfile (src/cart/src/) is service:cart, never service:src.
func TestSC2DockerfilesTakeTheirComposeName(t *testing.T) {
	for _, v := range hermeticVariants(t) {
		cart := "service:cart"
		if v.name == "perturbed" {
			cart = "service:basket"
		}
		found := false
		for _, n := range v.mm.Nodes {
			if n.ID == "service:src" {
				t.Errorf("%s: cart's Dockerfile named by its directory: service:src", v.name)
			}
			var dockerfile, compose bool
			for _, ev := range n.Evidence {
				dockerfile = dockerfile || isDockerfile(ev.Locator.Path)
				compose = compose || strings.HasPrefix(path.Base(ev.Locator.Path), "compose")
			}
			if dockerfile && !compose {
				t.Errorf("%s: %s comes from a Dockerfile but no compose service of that name builds it", v.name, n.ID)
			}
			if n.ID == cart && dockerfile {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: %s does not cite its Dockerfile", v.name, cart)
		}
	}
}

// edgeTo returns the relationship source -verb-> target and whether it exists.
func edgeTo(mm *domain.MacroMap, source string, verb domain.RelationVerb, target string) (domain.Relationship, bool) {
	for _, r := range mm.Relationships {
		if r.SourceID == source && r.Verb == verb && r.TargetID == target {
			return r, true
		}
	}
	return domain.Relationship{}, false
}

// hasEvidence reports whether some record has origin and a path with suffix.
func hasEvidence(evs []domain.EvidenceRecord, origin domain.DerivationOrigin, suffix string) bool {
	for _, ev := range evs {
		if ev.Origin == origin && strings.HasSuffix(ev.Locator.Path, suffix) && ev.Locator.StartLine > 0 {
			return true
		}
	}
	return false
}

// SC3: the Kafka topic is a Contract, provided by the Go producer (extracted, at a
// Go line) and by the broker (declared), consumed by the env-wired services
// (inferred, at their KAFKA_TOPIC line), and tasks 05/06 get a briefing.
func TestSC3KafkaTopicContracts(t *testing.T) {
	for _, v := range hermeticVariants(t) {
		topic, producer, consumers := "contract:kafka:orders", "service:checkout", []string{"service:accounting", "service:fraud-detection"}
		if v.name == "perturbed" {
			topic, producer, consumers = "contract:kafka:order-events", "service:ordering", []string{"service:ledger", "service:fraud-detection"}
		}
		found := false
		for _, n := range v.mm.Nodes {
			if n.ID == topic {
				found = n.Kind == domain.NodeContract
			}
		}
		if !found {
			t.Errorf("%s: no Contract %s", v.name, topic)
			continue
		}
		if r, ok := edgeTo(v.mm, producer, domain.VerbProvides, topic); !ok || !hasEvidence(r.Evidence, domain.OriginExtracted, ".go") {
			t.Errorf("%s: %s must provide %s with extracted Go evidence: %+v", v.name, producer, topic, r)
		}
		if r, ok := edgeTo(v.mm, "service:kafka", domain.VerbProvides, topic); !ok || !hasEvidence(r.Evidence, domain.OriginDeclared, ".yaml") {
			t.Errorf("%s: the broker must provide %s with declared evidence: %+v", v.name, topic, r)
		}
		for _, c := range consumers {
			r, ok := edgeTo(v.mm, c, domain.VerbConsumes, topic)
			if !ok || !hasEvidence(r.Evidence, domain.OriginInferred, "compose.full.yaml") {
				t.Errorf("%s: %s must consume %s, inferred, at its KAFKA_TOPIC line: %+v", v.name, c, topic, r)
			}
		}
		engine := slices.NewEngine(v.mm)
		for _, task := range v.gt.Tasks {
			if task.TargetID == topic {
				if _, err := engine.GetImpactBriefing(task.TargetID, task.Depth); err != nil {
					t.Errorf("%s %s: %v", v.name, task.ID, err)
				}
			}
		}
	}
}

// SC4: a gRPC Contract is consumed only when a method is called on a client
// bound to it; checkout builds EmailService and ShippingService clients and never
// calls them.
func TestSC4GRPCConsumptionNeedsACall(t *testing.T) {
	for _, v := range hermeticVariants(t) {
		svc, never, called := "service:checkout", []string{"contract:EmailService", "contract:ShippingService"}, []string{"contract:CartService", "contract:ProductCatalogService", "contract:PaymentService", "contract:CurrencyService"}
		if v.name == "perturbed" {
			// The perturbation removes checkout's CurrencyService client.
			svc, never, called = "service:ordering", []string{"contract:NotifyService", "contract:DispatchService", "contract:ExchangeService"}, []string{"contract:BasketService", "contract:InventoryService", "contract:BillingService"}
		}
		for _, c := range never {
			if r, ok := edgeTo(v.mm, svc, domain.VerbConsumes, c); ok {
				t.Errorf("%s: %s never calls %s but consumes it: %s", v.name, svc, c, origins(r))
			}
		}
		for _, c := range called {
			if r, ok := edgeTo(v.mm, svc, domain.VerbConsumes, c); !ok || !hasEvidence(r.Evidence, domain.OriginExtracted, "main.go") {
				t.Errorf("%s: %s calls %s and must consume it at the call: %+v", v.name, svc, c, r)
			}
		}
	}
}

func origins(r domain.Relationship) string {
	var out []string
	for _, ev := range r.Evidence {
		out = append(out, string(ev.Origin)+"@"+ev.Locator.String())
	}
	return strings.Join(out, " ")
}

// SC5: Go HTTP calls to an env-bound address plus a route are consumes edges
// with protocol and routes, extracted at the request line; compose env wiring
// that names another service is a declared edge at the env line; nothing cites
// unsupported-language source.
func TestSC5HTTPAndEnvWiring(t *testing.T) {
	for _, v := range hermeticVariants(t) {
		checkout, email, shipping, agent, frontend := "service:checkout", "service:email", "service:shipping", "service:agent", "service:frontend"
		if v.name == "perturbed" {
			checkout, email, shipping, frontend = "service:ordering", "service:notify", "service:dispatch", "service:storefront"
		}
		for target, routes := range map[string][]string{email: {"/send_order_confirmation"}, shipping: {"/get-quote", "/ship-order"}} {
			r, ok := edgeTo(v.mm, checkout, domain.VerbConsumes, target)
			if !ok || !hasEvidence(r.Evidence, domain.OriginExtracted, "main.go") {
				t.Errorf("%s: %s must consume %s at the request line: %s", v.name, checkout, target, origins(r))
				continue
			}
			got, _ := r.Metadata["routes"].([]string)
			if r.Metadata["protocol"] != "http" || strings.Join(got, " ") != strings.Join(routes, " ") {
				t.Errorf("%s: %s -> %s metadata %v, want http %v", v.name, checkout, target, r.Metadata, routes)
			}
		}
		if r, ok := edgeTo(v.mm, agent, domain.VerbConsumes, frontend); !ok || !hasEvidence(r.Evidence, domain.OriginDeclared, "compose.agent.yaml") {
			t.Errorf("%s: %s must consume %s through APPLICATION_ENDPOINT, declared: %s", v.name, agent, frontend, origins(r))
		}
		for _, r := range v.mm.Relationships {
			for _, ev := range r.Evidence {
				switch path.Ext(ev.Locator.Path) {
				case ".go", ".proto", ".yaml", ".yml", "", ".mod":
				default:
					t.Errorf("%s: %s cites %s", v.name, r.ID, ev.Locator.Path)
				}
			}
		}
	}
}

// line returns line n of a corpus file.
func line(t *testing.T, dir, p string, n int) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(p)))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	if n < 1 || n > len(lines) {
		t.Fatalf("%s#L%d out of bounds", p, n)
	}
	return lines[n-1]
}

var composeHeader = regexp.MustCompile(`^  [A-Za-z0-9_.-]+:\s*$`)

// SC6: compose-derived edges cite their depends_on entry or environment line,
// never the service header; declared ports carry their own locator; tour entry
// steps cite it.
func TestSC6LocatorsCiteTheirLine(t *testing.T) {
	for _, v := range hermeticVariants(t) {
		for _, r := range v.mm.Relationships {
			for _, ev := range r.Evidence {
				if !strings.HasPrefix(path.Base(ev.Locator.Path), "compose") {
					continue
				}
				if l := line(t, v.dir, ev.Locator.Path, ev.Locator.StartLine); composeHeader.MatchString(l) {
					t.Errorf("%s: %s cites the service header %s#L%d %q", v.name, r.ID, ev.Locator.Path, ev.Locator.StartLine, l)
				}
			}
		}
		for _, n := range v.mm.Nodes {
			if _, ok := n.Metadata["ports"]; !ok {
				continue
			}
			for _, key := range []string{"exposed_ports", "published_ports"} {
				entries, _ := n.Metadata[key].([]map[string]any)
				for _, e := range entries {
					loc, _ := e["locator"].(string)
					p, ln, ok := strings.Cut(loc, "#L")
					num := 0
					fmt.Sscanf(ln, "%d", &num)
					if !ok || num == 0 || !strings.Contains(line(t, v.dir, p, num), fmt.Sprint(e["port"])) {
						t.Errorf("%s: %s %s entry %v does not cite its own line", v.name, n.ID, key, e)
					}
				}
			}
			if n.Metadata["exposed_ports"] == nil && n.Metadata["published_ports"] == nil {
				t.Errorf("%s: %s declares ports %v without a locator", v.name, n.ID, n.Metadata["ports"])
			}
		}
		// The CLI and MCP read the index back from JSON: check the tour there.
		data, err := v.mm.CanonicalJSON()
		if err != nil {
			t.Fatal(err)
		}
		var loaded domain.MacroMap
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatal(err)
		}
		tour, err := slices.NewEngine(&loaded).GetTour("", 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range tour.Steps {
			if (s.WhyTemplate == slices.WhyEntry || s.WhyTemplate == slices.WhyEntryDeclared) && s.WhyLocator == "" {
				t.Errorf("%s: entry step %s has no why_locator: %s", v.name, s.NodeIDs[0], s.Why)
			}
		}
	}
}
