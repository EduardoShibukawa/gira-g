package extractors

import (
	"path"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// envEntry is a compose environment entry with its resolved value.
type envEntry struct {
	value   string
	locator domain.Locator
}

// wiring indexes the hints the cross-file joins share: each compose service's
// resolved environment, each service's Go env bindings, and the package-level
// env variables by package directory.
type wiring struct {
	env      map[string]map[string]envEntry // service → variable → entry
	bindings map[string]map[string]string   // service → field/var → variable ("" when ambiguous)
	vars     map[string]Hint                // package dir + "\x00" + var → HintGoEnvVar
	hints    []Hint
}

func newWiring(hints []Hint, scopes envScopes) *wiring {
	w := &wiring{
		env:      map[string]map[string]envEntry{},
		bindings: map[string]map[string]string{},
		vars:     map[string]Hint{},
		hints:    hints,
	}
	for _, h := range hints {
		switch h.Kind {
		case HintComposeEnv:
			raw := "${" + h.Key + "}"
			if h.Inline {
				raw = h.Value
			}
			v, ok := scopes.interpolate(h.Locator.Repo, path.Dir(h.Locator.Path), raw)
			if !ok {
				continue
			}
			if w.env[h.Service] == nil {
				w.env[h.Service] = map[string]envEntry{}
			}
			if _, dup := w.env[h.Service][h.Key]; !dup {
				w.env[h.Service][h.Key] = envEntry{value: v, locator: h.Locator}
			}
		case HintGoEnvBinding:
			if w.bindings[h.Service] == nil {
				w.bindings[h.Service] = map[string]string{}
			}
			if prev, seen := w.bindings[h.Service][h.Key]; seen && prev != h.Value {
				w.bindings[h.Service][h.Key] = "" // bound to two variables: ambiguous
			} else {
				w.bindings[h.Service][h.Key] = h.Value
			}
		case HintGoEnvVar:
			k := h.Context + "\x00" + h.Key
			if _, dup := w.vars[k]; !dup {
				w.vars[k] = h
			}
		}
	}
	return w
}

// lookup returns service's resolved environment variable, if compose declares it.
func (w *wiring) lookup(service, variable string) (envEntry, bool) {
	e, ok := w.env[service][variable]
	return e, ok && e.value != ""
}

// packageVar finds the package-level env variable name in the package that
// importPath names (by directory suffix), or in dir when importPath is empty.
func (w *wiring) packageVar(dir, importPath, name string) (Hint, bool) {
	if importPath == "" {
		h, ok := w.vars[dir+"\x00"+name]
		return h, ok
	}
	for _, h := range w.hints {
		if h.Kind == HintGoEnvVar && h.Key == name && (importPath == h.Context || strings.HasSuffix(importPath, "/"+h.Context)) {
			return h, true
		}
	}
	return Hint{}, false
}

// host returns the host part of an address value: scheme, port and path are
// stripped ("http://email:8080/x" → "email").
func host(value string) string {
	if i := strings.Index(value, "://"); i >= 0 {
		value = value[i+3:]
	}
	if i := strings.IndexAny(value, ":/"); i >= 0 {
		value = value[:i]
	}
	return value
}

func evidenceAt(loc domain.Locator, origin domain.DerivationOrigin, profile string) domain.EvidenceRecord {
	return domain.EvidenceRecord{Locator: loc, Origin: origin, DeploymentProfile: profile, Status: domain.StatusCurrent}
}

func composeProfile(loc domain.Locator) string { return "compose:" + path.Base(loc.Path) }

// addEdge adds a relationship, or appends its evidence when it already exists.
func (r *resolution) addEdge(source string, verb domain.RelationVerb, target string, evs ...domain.EvidenceRecord) {
	id := relID(source, verb, target)
	if rel, ok := r.rels[id]; ok {
		rel.Evidence = append(rel.Evidence, evs...)
		return
	}
	r.rels[id] = &domain.Relationship{ID: id, SourceID: source, TargetID: target, Verb: verb, Evidence: evs}
}

// relID follows the extractors' relationship IDs: bare service names, as compose
// writes them ("rel:checkout-consumes-email"), and sanitized contract names, as
// the Go extractor writes them ("rel:checkout-provides-kafka-orders"). An edge
// the resolution step finds again therefore merges with the per-file one.
func relID(source string, verb domain.RelationVerb, target string) string {
	bare := func(id string) string {
		if rest, ok := strings.CutPrefix(id, "contract:"); ok {
			return sanitizeRelID(rest)
		}
		return strings.TrimPrefix(id, "service:")
	}
	return "rel:" + bare(source) + "-" + string(verb) + "-" + bare(target)
}
