package extractors

import (
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// resolveHTTPCalls turns a Go HTTP request on an env-bound address into an edge
// to the compose service the address names, extracted at the request line, with
// protocol "http" and the sorted routes in the edge metadata. An address whose
// variable or host does not resolve emits nothing.
func resolveHTTPCalls(r *resolution, w *wiring) {
	for _, h := range w.hints {
		if h.Kind != HintHTTPCall {
			continue
		}
		variable := w.bindings[h.Service][h.Key]
		if variable == "" {
			continue
		}
		e, ok := w.lookup(h.Service, variable)
		if !ok {
			continue
		}
		target := "service:" + host(e.value)
		verb, ok := r.serviceEdgeVerb(target)
		if !ok || target == h.Service {
			continue
		}
		rel := r.addEdge(h.Service, verb, target, evidenceAt(h.Locator, domain.OriginExtracted, "go"))
		if rel.Metadata == nil {
			rel.Metadata = map[string]any{}
		}
		rel.Metadata["protocol"] = "http"
		if h.Value != "" {
			routes, _ := rel.Metadata["routes"].([]string)
			if !containsString(routes, h.Value) {
				routes = append(routes, h.Value)
				sort.Strings(routes)
			}
			rel.Metadata["routes"] = routes
		}
	}
}

// addressSuffixes are the variable-name endings that say a value is an address.
var addressSuffixes = []string{"_ADDR", "_ADDRESS", "_HOST", "_HOSTNAME", "_ENDPOINT", "_URL", "_URI", "_SERVER"}

// looksLikeAddress reports whether a variable's value is meant as an address:
// it has a scheme or a port, or the variable's name says so. This keeps values
// such as OTEL_SERVICE_NAME=agent from reading as wiring.
func looksLikeAddress(variable, value string) bool {
	if strings.Contains(value, "://") || strings.Contains(value, ":") {
		return true
	}
	for _, s := range addressSuffixes {
		if strings.HasSuffix(variable, s) {
			return true
		}
	}
	return false
}

// resolveEnvWiring turns a compose environment entry whose value addresses
// another compose service into a declared edge at the entry's line: accesses
// for Infrastructure, consumes otherwise.
func resolveEnvWiring(r *resolution, w *wiring) {
	for _, h := range w.hints {
		if h.Kind != HintComposeEnv {
			continue
		}
		v, ok := w.scopes.envValue(h)
		if !ok || !looksLikeAddress(h.Key, v) {
			continue
		}
		target := "service:" + host(v)
		verb, ok := r.serviceEdgeVerb(target)
		if !ok || target == h.Service {
			continue
		}
		if _, declared := r.nodes[h.Service]; !declared {
			continue
		}
		r.addEdge(h.Service, verb, target, evidenceAt(h.Locator, domain.OriginDeclared, composeProfile(h.Locator)))
	}
}

func containsString(xs []string, x string) bool {
	for _, have := range xs {
		if have == x {
			return true
		}
	}
	return false
}
