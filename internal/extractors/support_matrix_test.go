package extractors

import (
	"context"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestSupportMatrixCompletenessAndFixtures(t *testing.T) {
	matrix := SupportMatrix()

	expectedIDs := []string{
		"data-format-validator",
		"dockerfile-extractor",
		"compose-extractor",
		"envfile-extractor",
		"helm-extractor",
		"gomod-extractor",
		"github-actions-extractor",
		"openapi-extractor",
		"mise-extractor",
		"markdown-extractor",
		"protobuf-extractor",
		"go-ast-extractor",
		"grpc",
		"http",
		"kafka",
		"env-wiring",
		"go-intra",
	}

	if len(matrix) != len(expectedIDs) {
		t.Fatalf("expected %d entries in SupportMatrix, got %d", len(expectedIDs), len(matrix))
	}

	byID := make(map[string]ExtractorSupportEntry, len(matrix))
	for _, entry := range matrix {
		if entry.ID == "" || entry.Version == "" || entry.Layer == 0 {
			t.Fatalf("entry missing core metadata: %+v", entry)
		}
		if len(entry.SupportedFormats) == 0 || len(entry.ExtractedFacts) == 0 || len(entry.KnownLimitations) == 0 {
			t.Fatalf("entry %s has empty slices in support matrix", entry.ID)
		}
		if entry.PositiveFixture == "" || entry.NegativeFixture == "" {
			t.Fatalf("entry %s missing positive or negative fixture", entry.ID)
		}
		byID[entry.ID] = entry
	}

	for _, id := range expectedIDs {
		if _, ok := byID[id]; !ok {
			t.Fatalf("missing required support matrix entry %q", id)
		}
	}

	ctx := context.Background()
	pipe := DefaultPipeline()

	// Exercise positive and negative fixtures for each entry
	for _, entry := range matrix {
		t.Run(entry.ID, func(t *testing.T) {
			verifySupportEntryFixtures(t, ctx, pipe, entry)
		})
	}
}

func verifySupportEntryFixtures(t *testing.T, ctx context.Context, pipe *Pipeline, entry ExtractorSupportEntry) {
	t.Helper()
	switch entry.ID {
	case "data-format-validator":
		ext := NewDataFormatExtractor()
		_, _, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "config.json", Content: []byte(entry.PositiveFixture)})
		if err != nil {
			t.Fatalf("positive fixture failed: %v", err)
		}
		_, _, _, err = ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "config.json", Content: []byte(entry.NegativeFixture)})
		if err == nil {
			t.Fatalf("negative fixture must fail on malformed JSON")
		}

	case "dockerfile-extractor":
		ext := NewDockerfileExtractor()
		nodes, _, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "src/checkout/Dockerfile", Content: []byte(entry.PositiveFixture)})
		if err != nil || len(nodes) == 0 {
			t.Fatalf("positive fixture failed: err=%v nodes=%d", err, len(nodes))
		}
		nodesNeg, relsNeg, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "src/checkout/Dockerfile", Content: []byte(entry.NegativeFixture)})
		if err != nil || len(nodesNeg) != 0 || len(relsNeg) != 0 {
			t.Fatalf("negative fixture expected 0 nodes/rels, got nodes=%d rels=%d err=%v", len(nodesNeg), len(relsNeg), err)
		}

	case "compose-extractor":
		posMM, err := pipe.ProcessFiles(ctx, "sys", []TargetFile{
			{Repo: "r", CommitSHA: "c", Path: "compose.yaml", Content: []byte(entry.PositiveFixture)},
		})
		if err != nil || len(posMM.Relationships) == 0 {
			t.Fatalf("positive fixture expected corroborated relationship, got rels=%d err=%v", len(posMM.Relationships), err)
		}
		negMM, err := pipe.ProcessFiles(ctx, "sys", []TargetFile{
			{Repo: "r", CommitSHA: "c", Path: "compose.yaml", Content: []byte(entry.NegativeFixture)},
		})
		if err != nil || len(negMM.Relationships) != 0 {
			t.Fatalf("negative fixture (ordering-only depends_on) must emit 0 edges, got %d (err=%v)", len(negMM.Relationships), err)
		}

	case "envfile-extractor":
		ext := NewEnvFileExtractor()
		hints, err := ext.Hints(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: ".env", Content: []byte(entry.PositiveFixture)})
		if err != nil || len(hints) == 0 {
			t.Fatalf("positive fixture expected env hints, got %d (err=%v)", len(hints), err)
		}
		hintsNeg, err := ext.Hints(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: ".env", Content: []byte(entry.NegativeFixture)})
		if err != nil || len(hintsNeg) != 0 {
			t.Fatalf("negative fixture expected 0 hints, got %d (err=%v)", len(hintsNeg), err)
		}

	case "helm-extractor":
		ext := NewHelmExtractor()
		nodes, rels, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "Chart.yaml", Content: []byte(entry.PositiveFixture)})
		if err != nil || len(nodes) == 0 || len(rels) == 0 {
			t.Fatalf("positive fixture expected helm nodes and rels, got nodes=%d rels=%d err=%v", len(nodes), len(rels), err)
		}
		nodesNeg, relsNeg, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "Chart.yaml", Content: []byte(entry.NegativeFixture)})
		if err != nil || len(nodesNeg) != 0 || len(relsNeg) != 0 {
			t.Fatalf("negative fixture expected 0 nodes/rels, got nodes=%d rels=%d err=%v", len(nodesNeg), len(relsNeg), err)
		}

	case "gomod-extractor":
		ext := NewGoModExtractor()
		nodes, _, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "src/checkout/go.mod", Content: []byte(entry.PositiveFixture)})
		if err != nil || len(nodes) < 2 {
			t.Fatalf("positive fixture expected direct library nodes, got nodes=%d err=%v", len(nodes), err)
		}
		nodesNeg, _, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "src/empty/go.mod", Content: []byte(entry.NegativeFixture)})
		if err != nil || len(nodesNeg) != 0 {
			t.Fatalf("negative fixture (indirect only) expected 0 library nodes, got %d (err=%v)", len(nodesNeg), err)
		}

	case "github-actions-extractor":
		ext := NewGitHubActionsExtractor()
		nodes, rels, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: ".github/workflows/ci.yml", Content: []byte(entry.PositiveFixture)})
		if err != nil || len(nodes) == 0 || len(rels) == 0 {
			t.Fatalf("positive fixture expected workflow nodes/rels, got nodes=%d rels=%d err=%v", len(nodes), len(rels), err)
		}
		nodesNeg, relsNeg, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: ".github/workflows/ci.yml", Content: []byte(entry.NegativeFixture)})
		if err != nil || len(nodesNeg) != 0 || len(relsNeg) != 0 {
			t.Fatalf("negative fixture expected 0 nodes/rels, got nodes=%d rels=%d err=%v", len(nodesNeg), len(relsNeg), err)
		}

	case "openapi-extractor":
		ext := NewOpenAPIExtractor()
		nodes, rels, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "openapi.yaml", Content: []byte(entry.PositiveFixture)})
		if err != nil || len(nodes) == 0 || len(rels) == 0 {
			t.Fatalf("positive fixture expected contract and provides edge, got nodes=%d rels=%d err=%v", len(nodes), len(rels), err)
		}
		nodesNeg, relsNeg, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "openapi.yaml", Content: []byte(entry.NegativeFixture)})
		if err != nil || len(nodesNeg) != 0 || len(relsNeg) != 0 {
			t.Fatalf("negative fixture expected 0 nodes/rels, got nodes=%d rels=%d err=%v", len(nodesNeg), len(relsNeg), err)
		}

	case "mise-extractor":
		ext := NewMiseExtractor()
		nodes, _, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "mise.toml", Content: []byte(entry.PositiveFixture)})
		if err != nil || len(nodes) == 0 {
			t.Fatalf("positive fixture expected task/tool nodes, got nodes=%d err=%v", len(nodes), err)
		}
		nodesNeg, relsNeg, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "mise.toml", Content: []byte(entry.NegativeFixture)})
		if err != nil || len(nodesNeg) != 0 || len(relsNeg) != 0 {
			t.Fatalf("negative fixture expected 0 nodes/rels, got nodes=%d rels=%d err=%v", len(nodesNeg), len(relsNeg), err)
		}

	case "markdown-extractor":
		ext := NewMarkdownExtractor()
		nodes, rels, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "docs/decisions/0005-service-boundaries.md", Content: []byte(entry.PositiveFixture)})
		if err != nil || len(nodes) == 0 || len(rels) == 0 {
			t.Fatalf("positive fixture expected policy node and governs edge, got nodes=%d rels=%d err=%v", len(nodes), len(rels), err)
		}
		if ext.CanExtract("docs/general/notes.txt") {
			t.Fatalf("negative check: non-markdown file must not be extracted by markdown-extractor")
		}
		nodesNeg, relsNeg, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "docs/decisions/empty.md", Content: []byte(entry.NegativeFixture)})
		if err != nil || len(relsNeg) != 0 {
			t.Fatalf("negative fixture expected 0 relationships, got nodes=%d rels=%d err=%v", len(nodesNeg), len(relsNeg), err)
		}

	case "protobuf-extractor":
		ext := NewProtobufExtractor()
		nodes, _, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "pb/demo.proto", Content: []byte(entry.PositiveFixture)})
		if err != nil || len(nodes) == 0 {
			t.Fatalf("positive fixture expected contract node, got nodes=%d err=%v", len(nodes), err)
		}
		nodesNeg, _, _, err := ext.Extract(ctx, TargetFile{Repo: "r", CommitSHA: "c", Path: "pb/messages.proto", Content: []byte(entry.NegativeFixture)})
		if err != nil || len(nodesNeg) != 0 {
			t.Fatalf("negative fixture (messages only) expected 0 contract nodes, got %d (err=%v)", len(nodesNeg), err)
		}

	case "go-ast-extractor":
		posMM, err := pipe.ProcessFiles(ctx, "sys", []TargetFile{
			{Repo: "r", CommitSHA: "c", Path: "src/payment/main.go", Content: []byte(entry.PositiveFixture)},
		})
		if err != nil || len(posMM.Relationships) < 2 {
			t.Fatalf("positive fixture expected provides + consumes edges, got rels=%d err=%v", len(posMM.Relationships), err)
		}
		negMM, err := pipe.ProcessFiles(ctx, "sys", []TargetFile{
			{Repo: "r", CommitSHA: "c", Path: "src/payment/main.go", Content: []byte(entry.NegativeFixture)},
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range negMM.Relationships {
			if r.TargetID == "contract:EmailService" {
				t.Fatalf("negative fixture must not emit consumes edge for uncalled gRPC client")
			}
		}

	case "grpc":
		posMM, err := pipe.ProcessFiles(ctx, "sys", []TargetFile{
			{Repo: "r", CommitSHA: "c", Path: "src/checkout/client.go", Content: []byte("package main\nimport pb \"example.com/demo/pb\"\nfunc newS() *srv { return &srv{client: pb.NewCartServiceClient(nil)} }\n")},
			{Repo: "r", CommitSHA: "c", Path: "src/checkout/handler.go", Content: []byte(entry.PositiveFixture)},
		})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, r := range posMM.Relationships {
			if r.SourceID == "service:checkout" && r.TargetID == "contract:CartService" && r.Verb == domain.VerbConsumes {
				found = true
			}
		}
		if !found {
			t.Fatalf("positive grpc fixture failed to resolve cross-file CartService consumption")
		}
		negMM, err := pipe.ProcessFiles(ctx, "sys", []TargetFile{
			{Repo: "r", CommitSHA: "c", Path: "src/checkout/health.go", Content: []byte(entry.NegativeFixture)},
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range negMM.Nodes {
			if n.ID == "contract:Health" {
				t.Fatalf("negative grpc fixture leaked stdlib contract:Health")
			}
		}

	case "http":
		posMM, err := pipe.ProcessFiles(ctx, "sys", []TargetFile{
			{Repo: "r", CommitSHA: "c", Path: "compose.yaml", Content: []byte("services:\n  checkout:\n    image: checkout\n    environment:\n      - SHIPPING_ADDR=http://shipping:8080\n  shipping:\n    image: shipping\n")},
			{Repo: "r", CommitSHA: "c", Path: "src/checkout/main.go", Content: []byte(entry.PositiveFixture)},
		})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, r := range posMM.Relationships {
			if r.SourceID == "service:checkout" && r.TargetID == "service:shipping" && r.Metadata["protocol"] == "http" {
				found = true
			}
		}
		if !found {
			t.Fatalf("positive http fixture failed to resolve HTTP call to service:shipping")
		}
		negMM, err := pipe.ProcessFiles(ctx, "sys", []TargetFile{
			{Repo: "r", CommitSHA: "c", Path: "src/checkout/main.go", Content: []byte(entry.NegativeFixture)},
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range negMM.Relationships {
			if r.Verb == domain.VerbConsumes {
				t.Fatalf("negative http fixture emitted unexpected consumes edge: %+v", r)
			}
		}

	case "kafka":
		posMM, err := pipe.ProcessFiles(ctx, "sys", []TargetFile{
			{Repo: "r", CommitSHA: "c", Path: "compose.yaml", Content: []byte("services:\n  checkout:\n    image: checkout\n    environment:\n      - KAFKA_ADDR=kafka:9092\n  accounting:\n    image: accounting\n    environment:\n      - KAFKA_ADDR=kafka:9092\n      - KAFKA_TOPIC=orders\n  kafka:\n    image: kafka\n")},
			{Repo: "r", CommitSHA: "c", Path: "src/checkout/main.go", Content: []byte(entry.PositiveFixture)},
		})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, n := range posMM.Nodes {
			if n.ID == "contract:kafka:orders" {
				found = true
			}
		}
		if !found {
			t.Fatalf("positive kafka fixture failed to emit contract:kafka:orders")
		}
		negMM, err := pipe.ProcessFiles(ctx, "sys", []TargetFile{
			{Repo: "r", CommitSHA: "c", Path: "src/checkout/main.go", Content: []byte(entry.NegativeFixture)},
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range negMM.Nodes {
			if n.Kind == domain.NodeContract {
				t.Fatalf("negative kafka fixture emitted unexpected contract: %+v", n)
			}
		}

	case "env-wiring":
		posMM, err := pipe.ProcessFiles(ctx, "sys", []TargetFile{
			{Repo: "r", CommitSHA: "c", Path: "compose.yaml", Content: []byte(entry.PositiveFixture)},
		})
		if err != nil || len(posMM.Relationships) == 0 {
			t.Fatalf("positive env-wiring fixture expected declared consumes edge, got %d (err=%v)", len(posMM.Relationships), err)
		}
		negMM, err := pipe.ProcessFiles(ctx, "sys", []TargetFile{
			{Repo: "r", CommitSHA: "c", Path: "compose.yaml", Content: []byte(entry.NegativeFixture)},
		})
		if err != nil || len(negMM.Relationships) != 0 {
			t.Fatalf("negative env-wiring fixture expected 0 edges, got %d (err=%v)", len(negMM.Relationships), err)
		}

	case "go-intra":
		posMM, err := pipe.ProcessFiles(ctx, "sys", []TargetFile{
			{Repo: "r", CommitSHA: "c", Path: "go.mod", Content: []byte(entry.PositiveFixture)},
			{Repo: "r", CommitSHA: "c", Path: "cmd/app/main.go", Content: []byte("package main\nimport \"example.com/acme/internal/pkg\"\nfunc main() { _ = pkg.Run }\n")},
		})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, n := range posMM.Nodes {
			if n.ID == "library:example.com/acme/internal/pkg" && n.Metadata["origin_detail"] == "intra-module" {
				found = true
			}
		}
		if !found {
			t.Fatalf("positive go-intra fixture failed to mark intra-module library")
		}
		negMM, err := pipe.ProcessFiles(ctx, "sys", []TargetFile{
			{Repo: "r", CommitSHA: "c", Path: "go.mod", Content: []byte(entry.NegativeFixture)},
			{Repo: "r", CommitSHA: "c", Path: "cmd/app/main.go", Content: []byte("package main\nimport \"example.com/acme/internal/pkg\"\nfunc main() { _ = pkg.Run }\n")},
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range negMM.Nodes {
			if n.ID == "library:example.com/acme/internal/pkg" && n.Metadata["origin_detail"] == "intra-module" {
				t.Fatalf("negative go-intra fixture must not mark external module as intra-module")
			}
		}
	}
}

func TestComputeComponentCoverageTaxonomy(t *testing.T) {
	ctx := context.Background()
	pipe := DefaultPipeline()

	files := []TargetFile{
		{
			Repo: "demo", CommitSHA: "sha1", Path: "pb/demo.proto",
			Content: []byte("syntax = \"proto3\";\npackage oteldemo;\nservice CurrencyService {\n  rpc Convert(Req) returns (Resp);\n}\nservice PaymentService {\n  rpc Charge(Req) returns (Resp);\n}\n"),
		},
		{
			Repo: "demo", CommitSHA: "sha1", Path: "compose.yaml",
			Content: []byte("services:\n  checkout:\n    image: checkout\n  standalone:\n    image: standalone\n  catalog:\n    image: catalog\n  currency:\n    image: currency\n"),
		},
		// 1. Go service with full coverage and dependencies
		{
			Repo: "demo", CommitSHA: "sha1", Path: "src/checkout/main.go",
			Content: []byte("package main\nimport pb \"example.com/demo/pb\"\nfunc main() {\n\tclient := pb.NewPaymentServiceClient(nil)\n\t_, _ = client.Charge(nil, nil)\n}\n"),
		},
		// 2. Go service with full coverage and zero outgoing service/contract dependencies ("no-dependency")
		{
			Repo: "demo", CommitSHA: "sha1", Path: "src/standalone/main.go",
			Content: []byte("package main\nfunc main() {}\n"),
		},
		// 3. Go service with an unsupported construct ("partial")
		{
			Repo: "demo", CommitSHA: "sha1", Path: "src/catalog/main.go",
			Content: []byte("package main\nimport \"github.com/open-telemetry/opamp-go/client\"\nfunc startOpAMP() { _ = client.NewHTTP(nil) }\nfunc main() { startOpAMP() }\n"),
		},
		// 4. Polyglot C++ service with no Layer 3 extractor ("manifest-only", "no-evidence", UnsupportedFiles + UnknownJoins)
		{
			Repo: "demo", CommitSHA: "sha1", Path: "src/currency/src/server.cpp",
			Content: []byte("// C++ CurrencyService server implementation\nclass CurrencyServiceImpl final : public CurrencyService::Service {};\n"),
		},
	}

	mm, err := pipe.ProcessFiles(ctx, "demo-system", files)
	if err != nil {
		t.Fatal(err)
	}

	cov := ComputeComponentCoverage(mm)
	byID := make(map[string]ComponentCoverage, len(cov))
	for _, c := range cov {
		byID[c.ComponentID] = c
	}

	// Verify service:checkout -> full, has-dependencies
	checkout, ok := byID["service:checkout"]
	if !ok {
		t.Fatalf("missing coverage entry for service:checkout")
	}
	if checkout.CoverageStatus != CoverageStatusFull || checkout.DependencyPosture != DependencyPostureHasDependencies {
		t.Fatalf("service:checkout want full/has-dependencies, got %+v", checkout)
	}

	// Verify service:standalone -> full, no-dependency (explicitly distinguished from no-evidence!)
	standalone, ok := byID["service:standalone"]
	if !ok {
		t.Fatalf("missing coverage entry for service:standalone")
	}
	if standalone.CoverageStatus != CoverageStatusFull || standalone.DependencyPosture != DependencyPostureNoDependency {
		t.Fatalf("service:standalone want full/no-dependency, got %+v", standalone)
	}

	// Verify service:catalog -> partial, with UnsupportedConstructs
	catalog, ok := byID["service:catalog"]
	if !ok {
		t.Fatalf("missing coverage entry for service:catalog")
	}
	if catalog.CoverageStatus != CoverageStatusPartial || len(catalog.UnsupportedConstructs) == 0 {
		t.Fatalf("service:catalog want partial with UnsupportedConstructs, got %+v", catalog)
	}

	// Verify service:currency -> manifest-only, no-evidence, with UnsupportedFiles and UnknownJoins
	currency, ok := byID["service:currency"]
	if !ok {
		t.Fatalf("missing coverage entry for service:currency")
	}
	if currency.CoverageStatus != CoverageStatusManifestOnly || currency.DependencyPosture != DependencyPostureNoEvidence {
		t.Fatalf("service:currency want manifest-only/no-evidence, got %+v", currency)
	}
	if len(currency.UnsupportedFiles) != 1 || currency.UnsupportedFiles[0] != "src/currency/src/server.cpp" {
		t.Fatalf("service:currency want UnsupportedFiles=[src/currency/src/server.cpp], got %v", currency.UnsupportedFiles)
	}
	if len(currency.UnknownJoins) != 1 || currency.UnknownJoins[0] != "contract:CurrencyService" {
		t.Fatalf("service:currency want UnknownJoins=[contract:CurrencyService], got %v", currency.UnknownJoins)
	}
}
