package indexer_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/corpus"
	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/extractors"
	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/indexer"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

func writeTree(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for p, c := range files {
		full := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func makeManifestAndCache(t *testing.T, srcDir string) (*manifest.Manifest, *gitcache.CacheManager) {
	t.Helper()
	ctx := context.Background()
	repoDir := filepath.Join(t.TempDir(), "repo")
	sha, err := corpus.Materialize(ctx, srcDir, repoDir)
	if err != nil {
		t.Fatal(err)
	}
	cm, err := gitcache.NewCacheManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m := &manifest.Manifest{
		Version:      1,
		TargetSystem: manifest.TargetSystemMeta{Name: "demo-system"},
		Repositories: []manifest.RepositoryDeclaration{
			{ID: "demo", URL: repoDir, SHA: sha, PathScope: "."},
		},
	}
	return m, cm
}

func TestClassifyRefreshCommentOnlyAndRefreshClasses(t *testing.T) {
	pipe := extractors.DefaultPipeline()

	baseFiles := []extractors.TargetFile{
		{
			Repo: "demo", CommitSHA: "sha1", Path: "proto/order.proto",
			Content: []byte("syntax = \"proto3\";\npackage demo;\nservice OrderService {\n  rpc Place(Req) returns (Resp) {}\n}\n"),
		},
		{
			Repo: "demo", CommitSHA: "sha1", Path: "src/order/main.go",
			Content: []byte("package main\n\nimport pb \"example.com/demo/pb\"\n\nfunc main() {\n\tpb.RegisterOrderServiceServer(nil, nil)\n}\n"),
		},
		{
			Repo: "demo", CommitSHA: "sha1", Path: "docker-compose.yml",
			Content: []byte("services:\n  order:\n    image: order:latest\n"),
		},
	}
	prevFP := indexer.ComputeFingerprints(baseFiles, pipe)

	// 1. Comment-only and blank-line edits in Go, Proto, and Compose -> RefreshSkip
	commentEditedFiles := []extractors.TargetFile{
		{
			Repo: "demo", CommitSHA: "sha2", Path: "proto/order.proto",
			Content: []byte("// Top comment\nsyntax = \"proto3\";\n/* block comment */\npackage demo;\nservice OrderService {\n  // rpc doc\n  rpc Place(Req) returns (Resp) {}\n}\n\n"),
		},
		{
			Repo: "demo", CommitSHA: "sha2", Path: "src/order/main.go",
			Content: []byte("// File header comment\npackage main\n\nimport pb \"example.com/demo/pb\"\n\n/* multi-line\n   comment */\nfunc main() {\n\t// inside main\n\tpb.RegisterOrderServiceServer(nil, nil)\n}\n"),
		},
		{
			Repo: "demo", CommitSHA: "sha2", Path: "docker-compose.yml",
			Content: []byte("# Compose header comment\nservices:\n  # order service comment\n  order:\n    image: order:latest # inline comment\n\n"),
		},
	}
	commentFP := indexer.ComputeFingerprints(commentEditedFiles, pipe)

	for path, prev := range prevFP {
		curr := commentFP[path]
		if prev.ContentSHA256 == curr.ContentSHA256 {
			t.Fatalf("expected raw ContentSHA256 to differ after comment edit on %s", path)
		}
		if prev.StructuralHash != curr.StructuralHash {
			t.Fatalf("expected StructuralHash to match after comment-only edit on %s: %s != %s", path, prev.StructuralHash, curr.StructuralHash)
		}
	}

	class, perFile := indexer.ClassifyRefresh(prevFP, commentFP)
	if class != indexer.RefreshSkip {
		t.Fatalf("expected RefreshSkip for comment-only edits, got %s (perFile=%v)", class, perFile)
	}

	// 2. Go file body change -> RefreshPartial
	goBodyEdited := []extractors.TargetFile{
		baseFiles[0],
		{
			Repo: "demo", CommitSHA: "sha3", Path: "src/order/main.go",
			Content: []byte("package main\n\nimport pb \"example.com/demo/pb\"\n\nfunc main() {\n\tpb.RegisterOrderServiceServer(nil, nil)\n\t_ = pb.NewOrderServiceClient(nil)\n}\n"),
		},
		baseFiles[2],
	}
	partialFP := indexer.ComputeFingerprints(goBodyEdited, pipe)
	class, perFile = indexer.ClassifyRefresh(prevFP, partialFP)
	if class != indexer.RefreshPartial {
		t.Fatalf("expected RefreshPartial for Go body change, got %s", class)
	}
	if perFile["src/order/main.go"] != indexer.RefreshPartial {
		t.Fatalf("expected src/order/main.go to be RefreshPartial, got %s", perFile["src/order/main.go"])
	}
	if perFile["proto/order.proto"] != indexer.RefreshSkip {
		t.Fatalf("expected proto/order.proto to be RefreshSkip, got %s", perFile["proto/order.proto"])
	}

	// 3. Compose or Proto change -> RefreshArch
	composeEdited := []extractors.TargetFile{
		baseFiles[0],
		baseFiles[1],
		{
			Repo: "demo", CommitSHA: "sha4", Path: "docker-compose.yml",
			Content: []byte("services:\n  order:\n    image: order:latest\n  payment:\n    image: payment:latest\n"),
		},
	}
	archFP := indexer.ComputeFingerprints(composeEdited, pipe)
	class, perFile = indexer.ClassifyRefresh(prevFP, archFP)
	if class != indexer.RefreshArch {
		t.Fatalf("expected RefreshArch for Compose change, got %s", class)
	}
	if perFile["docker-compose.yml"] != indexer.RefreshArch {
		t.Fatalf("expected docker-compose.yml to be RefreshArch, got %s", perFile["docker-compose.yml"])
	}

	protoEdited := []extractors.TargetFile{
		{
			Repo: "demo", CommitSHA: "sha5", Path: "proto/order.proto",
			Content: []byte("syntax = \"proto3\";\npackage demo;\nservice OrderService {\n  rpc Place(Req) returns (Resp) {}\n}\nservice BillingService {\n  rpc Charge(Req) returns (Resp) {}\n}\n"),
		},
		baseFiles[1],
		baseFiles[2],
	}
	protoArchFP := indexer.ComputeFingerprints(protoEdited, pipe)
	class, _ = indexer.ClassifyRefresh(prevFP, protoArchFP)
	if class != indexer.RefreshArch {
		t.Fatalf("expected RefreshArch for Proto change, got %s", class)
	}

	// 4. Empty previous fingerprints or extractor version change -> RefreshFull
	class, _ = indexer.ClassifyRefresh(nil, prevFP)
	if class != indexer.RefreshFull {
		t.Fatalf("expected RefreshFull when previous fingerprints are nil, got %s", class)
	}
	versionChanged := make(map[string]indexer.FileFingerprint, len(prevFP))
	for k, v := range prevFP {
		v.ExtractorVersion = "2.0.0"
		versionChanged[k] = v
	}
	class, _ = indexer.ClassifyRefresh(prevFP, versionChanged)
	if class != indexer.RefreshFull {
		t.Fatalf("expected RefreshFull when extractor version changes, got %s", class)
	}
}

func TestBuildIncrementalCanonicalJSONEquivalence(t *testing.T) {
	ctx := context.Background()

	// Step 1: initial full build via BuildIncremental (nil prevState -> FULL)
	src1 := t.TempDir()
	writeTree(t, src1, map[string]string{
		"proto/order.proto":  "syntax = \"proto3\";\npackage demo;\nservice OrderService {\n  rpc Place(Req) returns (Resp) {}\n}\n",
		"src/order/main.go":  "package main\n\nimport pb \"example.com/demo/pb\"\n\nfunc main() {\n\tpb.RegisterOrderServiceServer(nil, nil)\n}\n",
		"docker-compose.yml": "services:\n  order:\n    image: order\n",
	})
	m1, cm1 := makeManifestAndCache(t, src1)
	incMap1, state1, err := indexer.BuildIncremental(ctx, nil, nil, m1, cm1, indexer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if state1.LastClass != indexer.RefreshFull {
		t.Fatalf("expected initial incremental build class FULL, got %s", state1.LastClass)
	}
	cleanMap1, err := indexer.Build(ctx, m1, cm1, indexer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	jInc1, _ := incMap1.CanonicalJSON()
	jClean1, _ := cleanMap1.CanonicalJSON()
	if string(jInc1) != string(jClean1) {
		t.Fatalf("initial BuildIncremental != Build canonical JSON")
	}

	// Step 2: comment-only change -> RefreshSkip, prev.CanonicalJSON() == next.CanonicalJSON()
	srcSkip := t.TempDir()
	writeTree(t, srcSkip, map[string]string{
		"proto/order.proto":  "// comment only\nsyntax = \"proto3\";\npackage demo;\nservice OrderService {\n  rpc Place(Req) returns (Resp) {}\n}\n",
		"src/order/main.go":  "// comment\npackage main\n\nimport pb \"example.com/demo/pb\"\n\nfunc main() {\n\t// comment\n\tpb.RegisterOrderServiceServer(nil, nil)\n}\n",
		"docker-compose.yml": "# comment\nservices:\n  order:\n    image: order\n",
	})
	mSkip, cmSkip := makeManifestAndCache(t, srcSkip)
	incMapSkip, stateSkip, err := indexer.BuildIncremental(ctx, incMap1, state1, mSkip, cmSkip, indexer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if stateSkip.LastClass != indexer.RefreshSkip {
		t.Fatalf("expected RefreshSkip on comment-only change, got %s", stateSkip.LastClass)
	}
	jSkip, _ := incMapSkip.CanonicalJSON()
	if string(jInc1) != string(jSkip) {
		t.Fatalf("expected exact CanonicalJSON equivalence on RefreshSkip")
	}

	// Step 3: Go implementation change -> RefreshPartial, BuildIncremental == Build
	srcPartial := t.TempDir()
	writeTree(t, srcPartial, map[string]string{
		"proto/order.proto":  "syntax = \"proto3\";\npackage demo;\nservice OrderService {\n  rpc Place(Req) returns (Resp) {}\n}\n",
		"src/order/main.go":  "package main\n\nimport (\n\tpb \"example.com/demo/pb\"\n\t\"net/http\"\n)\n\nfunc main() {\n\tpb.RegisterOrderServiceServer(nil, nil)\n\t_ = http.ListenAndServe(\":8080\", nil)\n}\n",
		"docker-compose.yml": "services:\n  order:\n    image: order\n",
	})
	mPartial, cmPartial := makeManifestAndCache(t, srcPartial)
	incMapPartial, statePartial, err := indexer.BuildIncremental(ctx, incMap1, state1, mPartial, cmPartial, indexer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if statePartial.LastClass != indexer.RefreshPartial {
		t.Fatalf("expected RefreshPartial, got %s", statePartial.LastClass)
	}
	cleanPartial, err := indexer.Build(ctx, mPartial, cmPartial, indexer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	jIncPartial, _ := incMapPartial.CanonicalJSON()
	jCleanPartial, _ := cleanPartial.CanonicalJSON()
	if string(jIncPartial) != string(jCleanPartial) {
		t.Fatalf("BuildIncremental != Build on PARTIAL:\ngot:  %s\nwant: %s", string(jIncPartial), string(jCleanPartial))
	}

	// Step 4: Compose change -> RefreshArch, BuildIncremental == Build
	srcArch := t.TempDir()
	writeTree(t, srcArch, map[string]string{
		"proto/order.proto":  "syntax = \"proto3\";\npackage demo;\nservice OrderService {\n  rpc Place(Req) returns (Resp) {}\n}\n",
		"src/order/main.go":  "package main\n\nimport (\n\tpb \"example.com/demo/pb\"\n\t\"net/http\"\n)\n\nfunc main() {\n\tpb.RegisterOrderServiceServer(nil, nil)\n\t_ = http.ListenAndServe(\":8080\", nil)\n}\n",
		"docker-compose.yml": "services:\n  order:\n    image: order\n  shipping:\n    image: shipping\n",
	})
	mArch, cmArch := makeManifestAndCache(t, srcArch)
	incMapArch, stateArch, err := indexer.BuildIncremental(ctx, incMapPartial, statePartial, mArch, cmArch, indexer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if stateArch.LastClass != indexer.RefreshArch {
		t.Fatalf("expected RefreshArch, got %s", stateArch.LastClass)
	}
	cleanArch, err := indexer.Build(ctx, mArch, cmArch, indexer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	jIncArch, _ := incMapArch.CanonicalJSON()
	jCleanArch, _ := cleanArch.CanonicalJSON()
	if string(jIncArch) != string(jCleanArch) {
		t.Fatalf("BuildIncremental != Build on ARCH:\ngot:  %s\nwant: %s", string(jIncArch), string(jCleanArch))
	}
}

func TestSymbolLossGate(t *testing.T) {
	prev := &domain.MacroMap{
		SchemaVersion: "1.0.0",
		TargetSystem:  "demo",
		Nodes: []domain.Node{
			{
				ID:   "service:order",
				Kind: domain.NodeService,
				Name: "order",
				Evidence: []domain.EvidenceRecord{
					{Locator: domain.Locator{Repo: "demo", CommitSHA: "sha1", Path: "docker-compose.yml", StartLine: 2}},
				},
			},
			{
				ID:   "service:payment",
				Kind: domain.NodeService,
				Name: "payment",
				Evidence: []domain.EvidenceRecord{
					{Locator: domain.Locator{Repo: "demo", CommitSHA: "sha1", Path: "src/payment/main.go", StartLine: 1}},
				},
			},
			{
				ID:   "contract:OrderService",
				Kind: domain.NodeContract,
				Name: "OrderService",
				Evidence: []domain.EvidenceRecord{
					{Locator: domain.Locator{Repo: "demo", CommitSHA: "sha1", Path: "proto/order.proto", StartLine: 3}},
				},
			},
		},
		Relationships: []domain.Relationship{
			{
				ID:       "rel:order-provides-OrderService",
				SourceID: "service:order",
				TargetID: "contract:OrderService",
				Verb:     domain.VerbProvides,
				Evidence: []domain.EvidenceRecord{
					{Locator: domain.Locator{Repo: "demo", CommitSHA: "sha1", Path: "src/order/main.go", StartLine: 5}},
				},
			},
		},
	}

	// Next map loses service:payment and contract:OrderService
	next := &domain.MacroMap{
		SchemaVersion: "1.0.0",
		TargetSystem:  "demo",
		Nodes: []domain.Node{
			prev.Nodes[0],
		},
	}

	// 1. Unexplained loss (no deleted files, allowShrink=false) -> fails closed with ErrSymbolLossGate
	rep, err := indexer.CheckSymbolLossGate(prev, next, nil, false)
	if !errors.Is(err, indexer.ErrSymbolLossGate) {
		t.Fatalf("expected ErrSymbolLossGate, got %v", err)
	}
	if len(rep.UnexplainedNodes) != 1 || rep.UnexplainedNodes[0] != "service:payment" {
		t.Fatalf("expected unexplained node service:payment, got %v", rep.UnexplainedNodes)
	}
	if len(rep.UnexplainedContracts) != 1 || rep.UnexplainedContracts[0] != "contract:OrderService" {
		t.Fatalf("expected unexplained contract contract:OrderService, got %v", rep.UnexplainedContracts)
	}

	// 2. Source files deleted -> passes gate without error
	deleted := map[string]bool{
		"src/payment/main.go": true,
		"proto/order.proto":   true,
	}
	rep, err = indexer.CheckSymbolLossGate(prev, next, deleted, false)
	if err != nil {
		t.Fatalf("expected gate to pass when defining files were deleted, got %v", err)
	}
	if rep.HasUnexplainedLoss() {
		t.Fatalf("expected no unexplained loss when files were deleted, got %+v", rep)
	}

	// 3. Unexplained loss with allowShrink=true -> passes gate and marks AllowedByShrink
	rep, err = indexer.CheckSymbolLossGate(prev, next, nil, true)
	if err != nil {
		t.Fatalf("expected gate to pass when allowShrink=true, got %v", err)
	}
	if !rep.AllowedByShrink {
		t.Fatalf("expected AllowedByShrink=true")
	}
}
