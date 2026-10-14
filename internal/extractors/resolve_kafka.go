package extractors

import (
	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// topicSource is where a producer's topic name came from.
type topicSource struct {
	topic    string
	variable string // the environment variable the topic expression reads, if any
	evidence []domain.EvidenceRecord
}

// resolveKafkaTopics emits a Contract per Kafka topic a Go producer writes to,
// with the producer and its broker as providers and the services wired to the
// same topic variable as consumers (ADR 0013). A topic whose name does not
// resolve emits nothing.
func resolveKafkaTopics(r *resolution, w *wiring) {
	produces := map[string]bool{} // service + topic
	for _, h := range w.hints {
		if src, ok := w.topic(h); h.Kind == HintKafkaTopic && ok {
			produces[h.Service+"\x00"+src.topic] = true
		}
	}
	for _, h := range w.hints {
		if h.Kind != HintKafkaTopic {
			continue
		}
		src, ok := w.topic(h)
		if !ok {
			continue
		}
		topicID := "contract:kafka:" + src.topic
		if _, exists := r.nodes[topicID]; !exists {
			r.nodes[topicID] = &domain.Node{
				ID:       topicID,
				Kind:     domain.NodeContract,
				Name:     src.topic,
				Metadata: map[string]any{"protocol": "kafka", "topic": src.topic},
			}
		}
		topic := r.nodes[topicID]
		topic.Evidence = append(topic.Evidence, src.evidence[0])
		r.addEdge(h.Service, domain.VerbProvides, topicID, src.evidence...)

		// The broker the producer is built with hosts the topic.
		for _, b := range w.hints {
			if b.Kind != HintKafkaBroker || b.Service != h.Service {
				continue
			}
			for _, variable := range w.brokerVariables(b) {
				e, ok := w.lookup(h.Service, variable)
				if !ok {
					continue
				}
				broker := "service:" + host(e.value)
				if _, declared := r.nodes[broker]; broker == h.Service || !declared {
					continue
				}
				r.addEdge(broker, domain.VerbProvides, topicID, evidenceAt(e.locator, domain.OriginDeclared, composeProfile(e.locator)))
			}
		}

		// Every other service wired to the same topic variable consumes it. Their
		// code is not read, so the direction is inferred.
		if src.variable == "" {
			continue
		}
		for _, c := range w.hints {
			if c.Kind != HintComposeEnv || c.Key != src.variable || produces[c.Service+"\x00"+src.topic] {
				continue
			}
			if e, ok := w.lookup(c.Service, src.variable); ok && e.locator == c.Locator && e.value == src.topic {
				r.addEdge(c.Service, domain.VerbConsumes, topicID, evidenceAt(c.Locator, domain.OriginInferred, composeProfile(c.Locator)))
			}
		}
	}
}

// topic resolves a producer's topic: a literal, or the environment variable the
// expression reads, from the producer's compose wiring (declared) or else from
// the code's literal fallback (extracted).
func (w *wiring) topic(h Hint) (topicSource, bool) {
	code := evidenceAt(h.Locator, domain.OriginExtracted, "go")
	var variable, fallback string
	var fallbackAt domain.Locator
	switch h.Mode {
	case "literal":
		return topicSource{topic: h.Value, evidence: []domain.EvidenceRecord{code}}, h.Value != ""
	case "env":
		variable = h.Value
	case "field":
		variable = w.bindings[h.Service][h.Key]
	case "var":
		v, ok := w.packageVar(h.Context, h.Value, h.Key)
		switch {
		case ok:
			variable, fallback, fallbackAt = v.Value, v.Default, v.Locator
		case h.Value == "":
			variable = w.bindings[h.Service][h.Key] // a local variable bound to the environment
		default:
			return topicSource{}, false
		}
	}
	if variable != "" {
		if e, ok := w.lookup(h.Service, variable); ok {
			return topicSource{topic: e.value, variable: variable, evidence: []domain.EvidenceRecord{code, evidenceAt(e.locator, domain.OriginDeclared, composeProfile(e.locator))}}, true
		}
	}
	if fallback != "" {
		return topicSource{topic: fallback, variable: variable, evidence: []domain.EvidenceRecord{code, evidenceAt(fallbackAt, domain.OriginExtracted, "go")}}, true
	}
	return topicSource{}, false
}

// brokerVariables returns the environment variables a broker hint reads,
// following one call into the producer's constructor when the broker list is a
// parameter of the function that builds it.
func (w *wiring) brokerVariables(b Hint) []string {
	switch b.Mode {
	case "env":
		return []string{b.Value}
	case "field":
		if v := w.bindings[b.Service][b.Key]; v != "" {
			return []string{v}
		}
	case "param":
		var out []string
		for _, c := range w.hints {
			if c.Kind != HintGoCallArg || c.Service != b.Service || c.Key != b.Key || c.Index != b.Index {
				continue
			}
			if c.Value != b.Context && !hasPathSuffix(c.Value, b.Context) {
				continue
			}
			if v := w.bindings[c.Service][c.Ref]; v != "" {
				out = append(out, v)
			}
		}
		return out
	}
	return nil
}

func hasPathSuffix(importPath, dir string) bool {
	return len(importPath) > len(dir) && importPath[len(importPath)-len(dir)-1:] == "/"+dir
}
