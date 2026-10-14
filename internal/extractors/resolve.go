package extractors

import (
	"context"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// HintKind names a kind of resolution hint (ADR 0013).
type HintKind string

const (
	// HintEnvValue is one KEY=VALUE line of a compose project .env file.
	HintEnvValue HintKind = "env_value"
	// HintComposeBuild is a compose service's build.dockerfile (Value) and
	// build.context (Context), before interpolation.
	HintComposeBuild HintKind = "compose_build"
	// HintDockerfileService is the Service node (Service) a Dockerfile produced.
	HintDockerfileService HintKind = "dockerfile_service"
	// HintComposeEnv is one compose environment entry of Service: Key, and Value
	// when Inline (KEY=VALUE); a bare KEY takes its value from the project .env.
	HintComposeEnv HintKind = "compose_env"
	// HintGoEnvBinding binds a Go field or variable named Key to environment
	// variable Value, through os.Getenv("Value") or f(&Key, "Value").
	HintGoEnvBinding HintKind = "go_env_binding"
	// HintGoEnvVar is a package-level Go variable Key in package directory
	// Context whose value is environment variable Value (may be empty) with the
	// literal fallback Default.
	HintGoEnvVar HintKind = "go_env_var"
	// HintKafkaTopic is the Topic of a Kafka producer message built by Service.
	// Mode says how the topic expression reads: "literal" (Value), "env"
	// (Value), "var" (package variable Key of import path Value, "" for the same
	// package) or "field" (field or variable Key, see HintGoEnvBinding).
	HintKafkaTopic HintKind = "kafka_topic"
	// HintKafkaBroker is the broker list a Kafka producer of Service is built
	// with. Mode "env" (Value), "field" (Key) or "param" (parameter Index of the
	// enclosing function Key, in package directory Context).
	HintKafkaBroker HintKind = "kafka_broker"
	// HintGoCallArg is argument Index of a call to function Key of import path
	// Value, when the argument reads field or variable Ref.
	HintGoCallArg HintKind = "go_call_arg"
)

// Hint is an internal resolution record an Extractor emits next to its facts.
// Hints never reach the Macro Map: the resolution step joins them across files
// into facts, each keeping the locator of the line it rests on (ADR 0013).
type Hint struct {
	Kind    HintKind
	Locator domain.Locator
	// Service is the node ID the hint belongs to, when known.
	Service string
	Key     string
	Value   string
	Context string
	// Default, Mode, Index, Ref and Inline qualify some kinds; see each kind.
	Default string
	Mode    string
	Index   int
	Ref     string
	Inline  bool
}

// HintEmitter is implemented by Extractors that also emit resolution hints.
type HintEmitter interface {
	Hints(ctx context.Context, file TargetFile) ([]Hint, error)
}

// resolution is the state the resolution step works on: the per-file facts,
// keyed by ID, as the pipeline aggregated them.
type resolution struct {
	nodes map[string]*domain.Node
	rels  map[string]*domain.Relationship
	hints []Hint
}

// resolve runs the cross-file resolution step. It is pure: it reads only the
// hints and facts, in sorted order, and never guesses. Anything it cannot
// resolve is left as the per-file extractors emitted it.
func resolve(r *resolution) {
	sort.SliceStable(r.hints, func(i, j int) bool {
		a, b := r.hints[i], r.hints[j]
		if a.Locator.Repo != b.Locator.Repo {
			return a.Locator.Repo < b.Locator.Repo
		}
		if a.Locator.Path != b.Locator.Path {
			return a.Locator.Path < b.Locator.Path
		}
		if a.Locator.StartLine != b.Locator.StartLine {
			return a.Locator.StartLine < b.Locator.StartLine
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Key < b.Key
	})
	env := newEnvScopes(r.hints)
	attributeDockerfiles(r, env)
	resolveKafkaTopics(r, newWiring(r.hints, env))
}

// envScopes holds each compose project's .env, keyed by repo and directory.
type envScopes map[string]map[string]string

func scopeKey(repo, dir string) string { return repo + "\x00" + dir }

func newEnvScopes(hints []Hint) envScopes {
	s := envScopes{}
	for _, h := range hints {
		if h.Kind != HintEnvValue {
			continue
		}
		k := scopeKey(h.Locator.Repo, path.Dir(h.Locator.Path))
		if s[k] == nil {
			s[k] = map[string]string{}
		}
		if _, dup := s[k][h.Key]; !dup {
			s[k][h.Key] = h.Value
		}
	}
	return s
}

var interpolation = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?:(:?-)([^}]*))?\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

// interpolate expands ${VAR}, ${VAR:-default}, ${VAR-default} and $VAR against
// the .env of the compose project in dir, the way Compose does, including .env
// values that reference earlier variables. It reports false if any variable is
// undefined and has no default: a partially resolved value is never used.
func (s envScopes) interpolate(repo, dir, value string) (string, bool) {
	vars := s[scopeKey(repo, dir)]
	var expand func(v string, depth int) (string, bool)
	expand = func(v string, depth int) (string, bool) {
		if depth > 8 {
			return "", false
		}
		ok := true
		out := interpolation.ReplaceAllStringFunc(v, func(m string) string {
			sub := interpolation.FindStringSubmatch(m)
			name := sub[1]
			if name == "" {
				name = sub[4]
			}
			val, defined := vars[name]
			if sub[2] == ":-" && val == "" || sub[2] == "-" && !defined {
				return sub[3]
			}
			if !defined {
				ok = false
				return ""
			}
			res, good := expand(val, depth+1)
			ok = ok && good
			return res
		})
		return out, ok
	}
	return expand(value, 0)
}

// attributeDockerfiles renames the Service a Dockerfile produced (named by its
// directory) to the compose service that builds that Dockerfile, merging its
// evidence and metadata into the compose service's node.
func attributeDockerfiles(r *resolution, env envScopes) {
	produced := map[string]string{} // repo + path → node ID
	for _, h := range r.hints {
		if h.Kind == HintDockerfileService {
			produced[scopeKey(h.Locator.Repo, h.Locator.Path)] = h.Service
		}
	}
	for _, h := range r.hints {
		if h.Kind != HintComposeBuild || h.Value == "" {
			continue
		}
		dir := path.Dir(h.Locator.Path)
		dockerfile, ok := env.interpolate(h.Locator.Repo, dir, h.Value)
		if !ok {
			continue
		}
		buildCtx, ok := env.interpolate(h.Locator.Repo, dir, h.Context)
		if !ok {
			continue
		}
		p := path.Clean(path.Join(dir, buildCtx, dockerfile))
		from, found := produced[scopeKey(h.Locator.Repo, p)]
		if !found || from == h.Service {
			continue
		}
		renameNode(r, from, h.Service)
		delete(produced, scopeKey(h.Locator.Repo, p))
	}
}

// renameNode merges node from into node to (creating it if absent) and rewrites
// every relationship endpoint. Evidence keeps both nodes' records.
func renameNode(r *resolution, from, to string) {
	src, ok := r.nodes[from]
	if !ok {
		return
	}
	dst, exists := r.nodes[to]
	if !exists {
		n := *src
		n.ID = to
		n.Name = strings.TrimPrefix(to, "service:")
		r.nodes[to] = &n
	} else {
		dst.Evidence = append(dst.Evidence, src.Evidence...)
		if dst.Metadata == nil && len(src.Metadata) > 0 {
			dst.Metadata = map[string]any{}
		}
		for k, v := range src.Metadata {
			if _, has := dst.Metadata[k]; !has {
				dst.Metadata[k] = v
			}
		}
	}
	delete(r.nodes, from)
	var ids []string
	for id, rel := range r.rels {
		if rel.SourceID == from || rel.TargetID == from {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		rel := r.rels[id]
		moved := *rel
		if moved.SourceID == from {
			moved.SourceID = to
		}
		if moved.TargetID == from {
			moved.TargetID = to
		}
		moved.ID = strings.ReplaceAll(rel.ID, from, to)
		delete(r.rels, id)
		if existing, dup := r.rels[moved.ID]; dup {
			existing.Evidence = append(existing.Evidence, moved.Evidence...)
			continue
		}
		r.rels[moved.ID] = &moved
	}
}
