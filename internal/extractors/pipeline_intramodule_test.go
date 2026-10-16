package extractors

import (
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestMarkIntraModuleLibraries(t *testing.T) {
	files := []TargetFile{
		{Repo: "r", CommitSHA: "abc", Path: "go.mod", Content: []byte("module example.com/acme\n\ngo 1.24\n")},
	}
	nodeMap := map[string]*domain.Node{
		"library:example.com/acme/internal/slices": {
			ID:   "library:example.com/acme/internal/slices",
			Kind: domain.NodeLibrary,
			Metadata: map[string]any{
				"import_path": "example.com/acme/internal/slices",
			},
			Evidence: []domain.EvidenceRecord{{Locator: domain.Locator{Repo: "r"}}},
		},
		"library:github.com/external/dep": {
			ID:   "library:github.com/external/dep",
			Kind: domain.NodeLibrary,
			Metadata: map[string]any{
				"import_path": "github.com/external/dep",
			},
			Evidence: []domain.EvidenceRecord{{Locator: domain.Locator{Repo: "r"}}},
		},
		"service:api": {
			ID:   "service:api",
			Kind: domain.NodeService,
		},
	}
	markIntraModuleLibraries(nodeMap, files)
	if got := nodeMap["library:example.com/acme/internal/slices"].Metadata["origin_detail"]; got != "intra-module" {
		t.Fatalf("own package not marked intra-module: %v", got)
	}
	if _, has := nodeMap["library:github.com/external/dep"].Metadata["origin_detail"]; has {
		t.Fatalf("external dep must stay unmarked")
	}
	if _, has := nodeMap["service:api"].Metadata["origin_detail"]; has {
		t.Fatalf("non-library must stay unmarked")
	}
}
