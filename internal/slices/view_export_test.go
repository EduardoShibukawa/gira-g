package slices

import (
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// SC4: the HTML export is a single self-contained document (no external asset
// reference); the Markdown export is a standalone document.
func TestSCViewExport(t *testing.T) {
	mm := &domain.MacroMap{
		TargetSystem: "t",
		Nodes: []domain.Node{
			{ID: "service:x", Kind: domain.NodeService, Name: "x"},
			{ID: "contract:c", Kind: domain.NodeContract, Name: "c"},
		},
		Relationships: []domain.Relationship{
			{ID: "rel:1", SourceID: "service:x", TargetID: "contract:c", Verb: domain.VerbProvides},
		},
	}
	html := ExportHTML(mm)
	if !strings.Contains(html, "<html") {
		t.Fatalf("HTML export looks empty:\n%.200s", html)
	}
	for _, bad := range []string{"<script src=", "<link ", `src="http`, `href="http`, "@import", "url(http"} {
		if strings.Contains(html, bad) {
			t.Fatalf("HTML export is not self-contained: contains %q", bad)
		}
	}
	if md := ExportMarkdown(mm); !strings.Contains(md, "# Target System Architecture") {
		t.Fatalf("markdown export looks wrong:\n%.200s", md)
	}
}
