package extractors

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// ExtractionLayer aliases Layer for support matrix declarations, adding LayerResolution
// for cross-file resolvers.
type ExtractionLayer = Layer

const (
	// LayerResolution represents the cross-file join and resolution pass (ADR 0013).
	LayerResolution Layer = 4
)

// Coverage status values for ComponentCoverage.CoverageStatus.
const (
	CoverageStatusFull         = "full"
	CoverageStatusPartial      = "partial"
	CoverageStatusManifestOnly = "manifest-only"
)

// Dependency posture values explicitly distinguishing "no evidence" from "no dependency".
const (
	DependencyPostureNoEvidence      = "no-evidence"
	DependencyPostureNoDependency    = "no-dependency"
	DependencyPostureHasDependencies = "has-dependencies"
)

// ExtractorSupportEntry documents one extractor or cross-file resolver in Gira's
// canonical support matrix, including executable positive and negative fixtures.
type ExtractorSupportEntry struct {
	ID               string          `json:"id"`
	Version          string          `json:"version"`
	Layer            ExtractionLayer `json:"layer"`
	SupportedFormats []string        `json:"supported_formats"`
	ExtractedFacts   []string        `json:"extracted_facts"`
	KnownLimitations []string        `json:"known_limitations"`
	PositiveFixture  string          `json:"positive_fixture"`
	NegativeFixture  string          `json:"negative_fixture"`
}

// ComponentCoverage reports the extraction coverage taxonomy for one architectural
// component/service in a MacroMap.
type ComponentCoverage struct {
	ComponentID           string   `json:"component_id"`
	DetectedFactsCount    int      `json:"detected_facts_count"`
	CoveredFiles          []string `json:"covered_files"`
	UnsupportedFiles      []string `json:"unsupported_files,omitempty"`
	UnsupportedConstructs []string `json:"unsupported_constructs,omitempty"`
	UnknownJoins          []string `json:"unknown_joins,omitempty"`
	CoverageStatus        string   `json:"coverage_status"`
	DependencyPosture     string   `json:"dependency_posture"`
}

// SupportMatrix returns the canonical support matrix for all 12 Wave-1 extractors
// and all 5 cross-file resolvers.
func SupportMatrix() []ExtractorSupportEntry {
	return []ExtractorSupportEntry{
		{
			ID:               "data-format-validator",
			Version:          "1.0.0",
			Layer:            LayerDataFormat,
			SupportedFormats: []string{".json", ".yaml", ".yml"},
			ExtractedFacts:   []string{"syntax-validation"},
			KnownLimitations: []string{
				"Validates JSON/YAML syntax only; does not emit domain nodes or relationships on its own",
				"Does not validate JSON Schema or custom YAML tags",
			},
			PositiveFixture: `{"service": "checkout", "port": 8080}`,
			NegativeFixture: `{"service": "checkout", "port": `,
		},
		{
			ID:               "dockerfile-extractor",
			Version:          "1.0.0",
			Layer:            LayerToolMeaning,
			SupportedFormats: []string{"Dockerfile", "Dockerfile.*", "*.dockerfile"},
			ExtractedFacts:   []string{"Service node", "exposed_ports metadata", "ENTRYPOINT/CMD binary hints"},
			KnownLimitations: []string{
				"Multi-stage build stage aliases and base images are not emitted as standalone architectural nodes",
				"Dynamic ARG-interpolated image names without defaults are not resolved across external build scripts",
			},
			PositiveFixture: "FROM golang:1.24 AS builder\nWORKDIR /app\nCOPY . .\nRUN go build -o /bin/checkout ./cmd/checkout\nFROM alpine:3.20\nCOPY --from=builder /bin/checkout /bin/checkout\nEXPOSE 8080\nENTRYPOINT [\"/bin/checkout\"]\n",
			NegativeFixture: "# Comment-only Dockerfile with no FROM or service instructions\n",
		},
		{
			ID:               "compose-extractor",
			Version:          "1.0.0",
			Layer:            LayerToolMeaning,
			SupportedFormats: []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml", "compose.*.yaml"},
			ExtractedFacts:   []string{"Service and Infrastructure nodes", "published_ports metadata", "corroborated consumes/accesses relationships", "env wiring and dockerfile build hints"},
			KnownLimitations: []string{
				"Ordering-only depends_on entries without corroborating environment or network address wiring do not emit service-to-service consumes edges",
				"External Docker networks without service endpoints are not modeled as nodes",
			},
			PositiveFixture: "services:\n  checkout:\n    build:\n      context: .\n      dockerfile: src/checkout/Dockerfile\n    environment:\n      - CART_ADDR=cart:7070\n    depends_on:\n      - cart\n  cart:\n    image: cart:latest\n    ports:\n      - \"7070:7070\"\n",
			NegativeFixture: "services:\n  checkout:\n    image: checkout:latest\n    depends_on:\n      - quote\n  quote:\n    image: quote:latest\n",
		},
		{
			ID:               "envfile-extractor",
			Version:          "1.0.0",
			Layer:            LayerToolMeaning,
			SupportedFormats: []string{".env"},
			ExtractedFacts:   []string{"HintEnvValue variable default bindings for Compose interpolation"},
			KnownLimitations: []string{
				"Emits resolution hints only; does not create standalone nodes without a referencing Compose or Go service",
				"Command substitution $(...) in .env values is not evaluated",
			},
			PositiveFixture: "CART_HOST=cart\nCART_PORT=7070\nCART_SERVICE_ADDR=${CART_HOST}:${CART_PORT}\n",
			NegativeFixture: "# Empty comment-only .env file\n",
		},
		{
			ID:               "helm-extractor",
			Version:          "1.0.0",
			Layer:            LayerToolMeaning,
			SupportedFormats: []string{"Chart.yaml", "Chart.yml", "values.yaml", "values.yml", "values-*.yaml"},
			ExtractedFacts:   []string{"Service and Infrastructure nodes from chart dependencies and values components", "consumes/accesses relationships"},
			KnownLimitations: []string{
				"Unrendered Go template control flow in templates/*.yaml is not executed without values.yaml definitions",
				"Dynamic Helm lookup() functions against a live cluster are unsupported",
			},
			PositiveFixture: "apiVersion: v2\nname: checkout\nversion: 0.1.0\ndependencies:\n  - name: redis\n    version: \"18.0.0\"\n    repository: \"https://charts.bitnami.com/bitnami\"\n",
			NegativeFixture: "apiVersion: v2\nversion: 0.1.0\n",
		},
		{
			ID:               "gomod-extractor",
			Version:          "1.0.0",
			Layer:            LayerToolMeaning,
			SupportedFormats: []string{"go.mod"},
			ExtractedFacts:   []string{"Service/Library module node", "Library dependency nodes", "imports relationships"},
			KnownLimitations: []string{
				"Indirect dependencies (// indirect) are skipped to prevent transitive dependency noise",
				"replace directives targeting local paths outside the repository scope are not traversed",
			},
			PositiveFixture: "module github.com/acme/checkout\n\ngo 1.24\n\nrequire (\n\tgoogle.golang.org/grpc v1.65.0\n\tgithub.com/IBM/sarama v1.43.0\n)\n",
			NegativeFixture: "module github.com/acme/empty\n\ngo 1.24\n\nrequire (\n\tgolang.org/x/sys v0.20.0 // indirect\n)\n",
		},
		{
			ID:               "github-actions-extractor",
			Version:          "1.0.0",
			Layer:            LayerToolMeaning,
			SupportedFormats: []string{".github/workflows/*.yml", ".github/workflows/*.yaml"},
			ExtractedFacts:   []string{"Task nodes for CI/CD workflows", "Library nodes for external uses actions", "governs and imports relationships"},
			KnownLimitations: []string{
				"Dynamic matrix expressions in uses: fields are not expanded at static analysis time",
				" YAML files outside .github/workflows/ are ignored by this extractor",
			},
			PositiveFixture: "name: CI\non: [push]\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - run: go test ./...\n",
			NegativeFixture: "name: EmptyWorkflow\non: [push]\n",
		},
		{
			ID:               "openapi-extractor",
			Version:          "1.0.0",
			Layer:            LayerToolMeaning,
			SupportedFormats: []string{"openapi*.yaml", "openapi*.yml", "openapi*.json", "swagger*.yaml", "swagger*.json"},
			ExtractedFacts:   []string{"Contract node for REST API specification", "Service provider node and provides relationship", "HTTP route metadata"},
			KnownLimitations: []string{
				"External $ref URLs pointing to remote HTTP servers are not fetched",
				"Non-OpenAPI YAML/JSON files without openapi or swagger root keys are ignored",
			},
			PositiveFixture: "openapi: \"3.0.3\"\ninfo:\n  title: Shipping Service API\n  version: \"1.0.0\"\npaths:\n  /get-quote:\n    post:\n      operationId: getQuote\n",
			NegativeFixture: "name: not-an-openapi-spec\nversion: 1.0.0\n",
		},
		{
			ID:               "mise-extractor",
			Version:          "1.0.0",
			Layer:            LayerToolMeaning,
			SupportedFormats: []string{"mise.toml", ".mise.toml", "mise/config.toml", ".mise/config.toml"},
			ExtractedFacts:   []string{"Task nodes from [tasks.*]", "Infrastructure/toolchain metadata from [tools]", "task dependency relationships"},
			KnownLimitations: []string{
				"Inline shell scripts inside task run blocks are not parsed as ASTs",
				"Arbitrary TOML files not named mise.toml or config.toml under mise/ are ignored",
			},
			PositiveFixture: "[tools]\ngo = \"1.24\"\n\n[tasks.build]\ndescription = \"Build binary\"\nrun = \"go build ./...\"\n",
			NegativeFixture: "# Empty mise.toml with no tools or tasks\n",
		},
		{
			ID:               "markdown-extractor",
			Version:          "1.0.0",
			Layer:            LayerToolMeaning,
			SupportedFormats: []string{"docs/adr/*.md", "**/SKILL.md", "AGENTS.md", "CLAUDE.md", ".scratch/**/*.md"},
			ExtractedFacts:   []string{"Policy nodes from ADRs and agent rules", "Skill nodes from SKILL.md", "Task nodes from issue specs", "governs/implements/verifies relationships"},
			KnownLimitations: []string{
				"Generic README.md files outside ADR/Skill/Agent/Scratch paths are skipped to avoid prose hallucination",
				"Free-form prose without structural headings or frontmatter does not emit service dependency edges",
			},
			PositiveFixture: "---\nschema: dec/1\n---\n# ADR 0005: Service Boundaries\n\nGoverns: checkout, payment\n",
			NegativeFixture: "# Generic Notes\nJust some unstructured text without architectural entities.\n",
		},
		{
			ID:               "protobuf-extractor",
			Version:          "1.0.0",
			Layer:            LayerLanguageAST,
			SupportedFormats: []string{"*.proto"},
			ExtractedFacts:   []string{"Contract nodes for protobuf service and message schema definitions", "RPC method request/response metadata"},
			KnownLimitations: []string{
				"Proto files containing neither service, message, nor package declarations emit no Contract nodes",
				"Dynamic protobuf reflection descriptors loaded at runtime are not statically resolved",
			},
			PositiveFixture: "syntax = \"proto3\";\npackage demo;\nservice PaymentService {\n  rpc Charge(ChargeReq) returns (ChargeResp);\n}\nmessage ChargeReq {}\nmessage ChargeResp {}\n",
			NegativeFixture: "syntax = \"proto3\";\n// Comment-only proto file with no package, service, or message\n",
		},
		{
			ID:               "go-ast-extractor",
			Version:          "1.0.0",
			Layer:            LayerLanguageAST,
			SupportedFormats: []string{"*.go"},
			ExtractedFacts:   []string{"Service and Library nodes", "provides/consumes gRPC Contract edges", "imports relationships", "HTTP/Kafka/env resolution hints", "UnsupportedConstruct boundaries for unresolvable dynamic constructs"},
			KnownLimitations: []string{
				"*_test.go files are skipped during architectural extraction",
				"Reflection-driven calls and dynamic OpAMP/plugin transports emit explicit UnsupportedConstruct boundaries rather than guessed edges",
			},
			PositiveFixture: "package main\n\nimport pb \"example.com/demo/pb\"\n\nfunc main() {\n\tpb.RegisterPaymentServiceServer(nil, nil)\n\tclient := pb.NewCurrencyServiceClient(nil)\n\t_, _ = client.Convert(nil, nil)\n}\n",
			NegativeFixture: "package main\n\nimport pb \"example.com/demo/pb\"\n\nfunc main() {\n\t// Client constructed but never called (SC4)\n\t_ = pb.NewEmailServiceClient(nil)\n}\n",
		},
		{
			ID:               "grpc",
			Version:          "1.0.0",
			Layer:            LayerResolution,
			SupportedFormats: []string{"*.go", "*.proto"},
			ExtractedFacts:   []string{"Cross-file gRPC client call to Contract consumes edges", "stdlib google.golang.org/grpc/health denylist filtering"},
			KnownLimitations: []string{
				"Constructing a New<Service>Client without invoking any RPC method does not emit a consumes edge",
				"Untyped grpc.ClientConn.Invoke calls with dynamic method strings are not resolved",
			},
			PositiveFixture: "package main\n\nimport pb \"example.com/demo/pb\"\n\ntype srv struct{ client pb.CartServiceClient }\n\nfunc (s *srv) run() {\n\t_, _ = s.client.GetCart(nil, nil)\n}\n",
			NegativeFixture: "package main\n\nimport healthpb \"google.golang.org/grpc/health/grpc_health_v1\"\n\nfunc main() {\n\thealthpb.RegisterHealthServer(nil, nil)\n}\n",
		},
		{
			ID:               "http",
			Version:          "1.0.0",
			Layer:            LayerResolution,
			SupportedFormats: []string{"*.go", "compose*.yaml", ".env"},
			ExtractedFacts:   []string{"Cross-file HTTP client call joined with env-bound service address to Service consumes edge with protocol=http and routes metadata"},
			KnownLimitations: []string{
				"Requires an environment variable binding linking the HTTP base URL/host to a target service",
				"Arbitrary external URLs not bound to a known service in the Target System are ignored",
			},
			PositiveFixture: "package main\n\nimport (\n\t\"net/http\"\n\t\"os\"\n)\n\nfunc main() {\n\taddr := os.Getenv(\"SHIPPING_ADDR\")\n\t_, _ = http.Post(addr+\"/ship-order\", \"application/json\", nil)\n}\n",
			NegativeFixture: "package main\n\nimport \"net/http\"\n\nfunc main() {\n\t_, _ = http.Get(\"https://example.com/external\")\n}\n",
		},
		{
			ID:               "kafka",
			Version:          "1.0.0",
			Layer:            LayerResolution,
			SupportedFormats: []string{"*.go", "compose*.yaml", ".env"},
			ExtractedFacts:   []string{"Contract node contract:kafka:<topic>", "Producer Service provides contract:kafka:<topic> (extracted)", "Broker Service provides contract:kafka:<topic> (declared)", "Consumer Service consumes contract:kafka:<topic> (inferred from env wiring)"},
			KnownLimitations: []string{
				"Recognises Sarama producer messages and KAFKA_TOPIC/KAFKA_SERVICE_ADDR environment wiring",
				"Topics computed dynamically from user input at runtime cannot be statically named",
			},
			PositiveFixture: "package main\n\nimport \"github.com/IBM/sarama\"\n\nfunc publish(p sarama.SyncProducer) {\n\t_, _, _ = p.SendMessage(&sarama.ProducerMessage{Topic: \"orders\"})\n}\n",
			NegativeFixture: "package main\n\nfunc publish() {}\n",
		},
		{
			ID:               "env-wiring",
			Version:          "1.0.0",
			Layer:            LayerResolution,
			SupportedFormats: []string{"compose*.yaml", ".env", "*.go"},
			ExtractedFacts:   []string{"Declared Service-to-Service consumes/accesses edges from Compose environment host:port references", "Corroboration of depends_on entries"},
			KnownLimitations: []string{
				"Environment variables whose values do not reference a known Service or Infrastructure host are not turned into edges",
				"Encrypted secrets or external vault references are not decrypted",
			},
			PositiveFixture: "services:\n  agent:\n    image: agent:latest\n    environment:\n      - APPLICATION_ENDPOINT=http://frontend:8080\n  frontend:\n    image: frontend:latest\n",
			NegativeFixture: "services:\n  agent:\n    image: agent:latest\n    environment:\n      - LOG_LEVEL=debug\n",
		},
		{
			ID:               "go-intra",
			Version:          "1.0.0",
			Layer:            LayerResolution,
			SupportedFormats: []string{"go.mod", "*.go"},
			ExtractedFacts:   []string{"Intra-module Library node tagging (origin_detail=intra-module)", "Cross-package function/method call resolution within the same Go module"},
			KnownLimitations: []string{
				"Scoped to packages belonging to the Go module declared in the repository's go.mod",
				"Cross-repository Go calls without go.mod module declarations remain external library imports",
			},
			PositiveFixture: "module example.com/acme\n\ngo 1.24\n",
			NegativeFixture: "module example.com/other\n\ngo 1.24\n",
		},
	}
}

// ComputeComponentCoverage computes the per-component extraction coverage taxonomy
// for every Service node in mm, classifying each component as "full", "partial",
// or "manifest-only", and explicitly distinguishing "no-evidence" from "no-dependency".
func ComputeComponentCoverage(mm *domain.MacroMap) []ComponentCoverage {
	if mm == nil {
		return nil
	}

	var services []domain.Node
	for _, n := range mm.Nodes {
		if n.Kind == domain.NodeService {
			services = append(services, n)
		}
	}
	sort.Slice(services, func(i, j int) bool {
		return services[i].ID < services[j].ID
	})

	constructSet := make(map[string]bool, len(mm.UnsupportedConstructs))
	for _, p := range mm.UnsupportedConstructs {
		constructSet[filepath.ToSlash(p)] = true
	}

	out := make([]ComponentCoverage, 0, len(services))
	for _, svc := range services {
		svcName := svc.Name
		if svcName == "" {
			svcName = strings.TrimPrefix(svc.ID, "service:")
		}

		coveredSet := make(map[string]bool)
		factsCount := 0
		hasOutgoingDep := false

		for _, ev := range svc.Evidence {
			if ev.Locator.Path != "" {
				coveredSet[filepath.ToSlash(ev.Locator.Path)] = true
				factsCount++
			}
		}

		for _, r := range mm.Relationships {
			if r.SourceID != svc.ID && r.TargetID != svc.ID {
				continue
			}
			factsCount++
			if r.SourceID == svc.ID && (r.Verb == domain.VerbConsumes || r.Verb == domain.VerbAccesses || r.Verb == domain.VerbProvides) {
				hasOutgoingDep = true
			}
			for _, ev := range r.Evidence {
				if ev.Locator.Path != "" {
					p := filepath.ToSlash(ev.Locator.Path)
					if belongsToComponent(p, svcName) || !isSubdirOfAnotherService(p, svcName, services) {
						coveredSet[p] = true
					}
				}
			}
		}

		// Also check if any other nodes (e.g. intra-module libraries) have evidence under this component's directory
		for _, n := range mm.Nodes {
			for _, ev := range n.Evidence {
				p := filepath.ToSlash(ev.Locator.Path)
				if p != "" && belongsToComponent(p, svcName) {
					coveredSet[p] = true
				}
			}
		}

		unsupportedFilesSet := make(map[string]bool)
		unsupportedConstructsSet := make(map[string]bool)
		unknownJoinsSet := make(map[string]bool)

		for _, ub := range mm.Unsupported {
			p := filepath.ToSlash(ub.Path)
			if !belongsToComponent(p, svcName) {
				continue
			}
			if constructSet[p] {
				unsupportedConstructsSet[p] = true
			} else {
				unsupportedFilesSet[p] = true
			}
			for _, uj := range ub.UnknownJoins {
				if uj != "" {
					unknownJoinsSet[uj] = true
				}
			}
		}

		coveredFiles := sortedKeys(coveredSet)
		unsupportedFiles := sortedKeys(unsupportedFilesSet)
		unsupportedConstructs := sortedKeys(unsupportedConstructsSet)
		unknownJoins := sortedKeys(unknownJoinsSet)

		hasLayer3Code := false
		for _, f := range coveredFiles {
			ext := strings.ToLower(filepath.Ext(f))
			if ext == ".go" || ext == ".proto" {
				hasLayer3Code = true
				break
			}
		}

		var status string
		switch {
		case !hasLayer3Code:
			status = CoverageStatusManifestOnly
		case len(unsupportedFiles) > 0 || len(unsupportedConstructs) > 0:
			status = CoverageStatusPartial
		default:
			status = CoverageStatusFull
		}

		var posture string
		switch {
		case hasOutgoingDep:
			posture = DependencyPostureHasDependencies
		case status == CoverageStatusFull:
			posture = DependencyPostureNoDependency
		default:
			posture = DependencyPostureNoEvidence
		}

		out = append(out, ComponentCoverage{
			ComponentID:           svc.ID,
			DetectedFactsCount:    factsCount,
			CoveredFiles:          coveredFiles,
			UnsupportedFiles:      unsupportedFiles,
			UnsupportedConstructs: unsupportedConstructs,
			UnknownJoins:          unknownJoins,
			CoverageStatus:        status,
			DependencyPosture:     posture,
		})
	}

	return out
}

func belongsToComponent(path, svcName string) bool {
	if svcName == "" {
		return false
	}
	norm := strings.TrimPrefix(filepath.ToSlash(path), "./")
	prefixes := []string{
		"src/" + svcName + "/",
		"services/" + svcName + "/",
		"cmd/" + svcName + "/",
		svcName + "/",
	}
	for _, pref := range prefixes {
		if strings.HasPrefix(norm, pref) {
			return true
		}
	}
	return false
}

func isSubdirOfAnotherService(path, currentSvc string, services []domain.Node) bool {
	for _, other := range services {
		otherName := other.Name
		if otherName == "" {
			otherName = strings.TrimPrefix(other.ID, "service:")
		}
		if otherName == currentSvc {
			continue
		}
		if belongsToComponent(path, otherName) {
			return true
		}
	}
	return false
}

