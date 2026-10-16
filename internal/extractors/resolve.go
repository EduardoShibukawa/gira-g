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
	// HintGRPCClient binds field or variable Key to a client of gRPC Contract
	// Value (a field typed pkg.<Value>Client, or assigned pkg.New<Value>Client).
	HintGRPCClient HintKind = "grpc_client"
	// HintGoAlias records Key = Ref, a variable assigned from a field or variable.
	HintGoAlias HintKind = "go_alias"
	// HintGoMethodCall is a call of exported method Value on field or variable Key.
	HintGoMethodCall HintKind = "go_method_call"
	// HintDependsOn is one compose depends_on entry: Service consumes Value.
	// The Locator is the entry's line (or the service header when the entry
	// line is missing), exactly the line the edge evidence cites, so the
	// resolution step can tell ordering-only lines from corroboration.
	HintDependsOn HintKind = "depends_on"
	// HintHTTPCall is a net/http request whose URL is field or variable Key plus
	// the literal route Value ("" when the URL is the address alone).
	HintHTTPCall HintKind = "http_call"
	// HintGoPackage is one Go package clause: Key is the package name,
	// Context the package directory.
	HintGoPackage HintKind = "go_package"
	// HintGoType is one top-level Go type declaration: Key is the type name,
	// Value the shape (struct, interface, alias, other).
	HintGoType HintKind = "go_type"
	// HintGoFunc is one top-level Go func or method declaration: Key is the
	// func name, Value the dereferenced receiver type ("" for a plain
	// function), Mode "exported" or "unexported".
	HintGoFunc HintKind = "go_func"
	// HintGoImport is one Go import statement: Key the import path, Value
	// the local name (explicit alias, or the derived base name).
	HintGoImport HintKind = "go_import"
	// HintGoCall is one Go call site: Key the enclosing func ("Recv.Name",
	// "Name", or "" at file scope), Value the callee name as written, Ref
	// the import path when the qualifier names an import ("" otherwise),
	// Mode "pkg", "local", "bare" or "bare-dotimport" (a bare call in a
	// file with dot-imports, never resolved across packages).
	HintGoCall HintKind = "go_call"
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
	// modules is the go.mod module paths by repo, for import scope.
	modules map[string][]string
	// renamed maps a service ID to the compose service it was attributed
	// to, so later steps attach facts to the surviving node.
	renamed map[string]string
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
	w := newWiring(r.hints, env)
	resolveKafkaTopics(r, w)
	resolveGRPCCalls(r, w)
	resolveHTTPCalls(r, w)
	resolveEnvWiring(r, w)
	pruneOrderingOnlyEdges(r)
	resolveIntraService(r)
}

// pruneOrderingOnlyEdges drops service-consumes-service edges whose evidence
// is all compose depends_on lines (J2, SC2): start ordering is not
// consumption. Corroboration between the same pair keeps the edge, which
// then cites both the depends_on line and the corroborating line:
//   - other evidence on the edge itself (a call-site line, an env-wiring
//     entry), already merged into its evidence;
//   - an accesses edge between the same pair resting on a non-depends_on
//     line, whose lines are appended to the kept edge (F5);
//   - a contract reference: the source consumes a Contract the target
//     provides; the source's consumption lines are appended (F5).
//
// Infrastructure accesses edges are untouched.
func pruneOrderingOnlyEdges(r *resolution) {
	type lineKey struct {
		repo, path string
		line       int
	}
	ordering := map[string]map[lineKey]bool{} // "source\x00target" → depends_on lines
	for _, h := range r.hints {
		if h.Kind != HintDependsOn {
			continue
		}
		k := h.Service + "\x00service:" + h.Value
		if ordering[k] == nil {
			ordering[k] = map[lineKey]bool{}
		}
		ordering[k][lineKey{h.Locator.Repo, h.Locator.Path, h.Locator.StartLine}] = true
	}
	accesses, consumed, provided := corroborationIndex(r.rels)
	for id, rel := range r.rels {
		if rel.Verb != domain.VerbConsumes {
			continue
		}
		if !strings.HasPrefix(rel.SourceID, "service:") || !strings.HasPrefix(rel.TargetID, "service:") {
			continue
		}
		lines, ok := ordering[rel.SourceID+"\x00"+rel.TargetID]
		if !ok || len(rel.Evidence) == 0 {
			continue
		}
		onlyOrdering := true
		for _, ev := range rel.Evidence {
			if !lines[lineKey{ev.Locator.Repo, ev.Locator.Path, ev.Locator.StartLine}] {
				onlyOrdering = false
				break
			}
		}
		if !onlyOrdering {
			continue
		}
		var extra []domain.EvidenceRecord
		corroborate := func(evs []domain.EvidenceRecord) {
			for _, ev := range evs {
				if !lines[lineKey{ev.Locator.Repo, ev.Locator.Path, ev.Locator.StartLine}] {
					extra = append(extra, ev)
				}
			}
		}
		if acc, ok := accesses[rel.SourceID+"\x00"+rel.TargetID]; ok {
			corroborate(acc.Evidence)
		}
		for _, c := range sortedKeys(consumed[rel.SourceID]) {
			if provided[rel.TargetID+"\x00"+c] {
				corroborate(consumed[rel.SourceID][c].Evidence)
			}
		}
		if len(extra) == 0 {
			delete(r.rels, id)
			continue
		}
		rel.Evidence = appendNewEvidence(rel.Evidence, extra)
	}
}

// corroborationIndex indexes the edges that can corroborate a depends_on
// edge: service-accesses-service edges by "source\x00target", the Contracts
// each service consumes, and the "service\x00contract" pairs provided.
func corroborationIndex(rels map[string]*domain.Relationship) (accesses map[string]*domain.Relationship, consumed map[string]map[string]*domain.Relationship, provided map[string]bool) {
	accesses = map[string]*domain.Relationship{}
	consumed = map[string]map[string]*domain.Relationship{}
	provided = map[string]bool{}
	for _, rel := range rels {
		if !strings.HasPrefix(rel.SourceID, "service:") {
			continue
		}
		switch {
		case rel.Verb == domain.VerbAccesses && strings.HasPrefix(rel.TargetID, "service:"):
			accesses[rel.SourceID+"\x00"+rel.TargetID] = rel
		case rel.Verb == domain.VerbConsumes && strings.HasPrefix(rel.TargetID, "contract:"):
			if consumed[rel.SourceID] == nil {
				consumed[rel.SourceID] = map[string]*domain.Relationship{}
			}
			consumed[rel.SourceID][rel.TargetID] = rel
		case rel.Verb == domain.VerbProvides && strings.HasPrefix(rel.TargetID, "contract:"):
			provided[rel.SourceID+"\x00"+rel.TargetID] = true
		}
	}
	return accesses, consumed, provided
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// appendNewEvidence appends the records of extra whose locator evs does not
// already cite.
func appendNewEvidence(evs, extra []domain.EvidenceRecord) []domain.EvidenceRecord {
	seen := map[string]bool{}
	for _, ev := range evs {
		seen[ev.Locator.String()] = true
	}
	for _, ev := range extra {
		if k := ev.Locator.String(); !seen[k] {
			seen[k] = true
			evs = append(evs, ev)
		}
	}
	return evs
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
		if r.renamed != nil {
			r.renamed[from] = h.Service
		}
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
		moved.ID = relID(moved.SourceID, moved.Verb, moved.TargetID)
		delete(r.rels, id)
		if existing, dup := r.rels[moved.ID]; dup {
			existing.Evidence = append(existing.Evidence, moved.Evidence...)
			continue
		}
		r.rels[moved.ID] = &moved
	}
}
