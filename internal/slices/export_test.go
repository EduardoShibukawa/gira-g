package slices_test

import (
	"strings"
	"testing"

	"labs/gira-g/internal/domain"
	"labs/gira-g/internal/slices"
)

func createSampleMacroMapForExport() *domain.MacroMap {
	mm := domain.NewMacroMap("checkout-system")

	_ = mm.AddNode(domain.Node{
		ID:   "service:checkout",
		Kind: domain.NodeService,
		Name: "checkoutservice",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "checkout-repo",
					CommitSHA: "1234567890abcdef",
					Path:      "src/main.go",
					StartLine: 1,
					EndLine:   40,
				},
				Origin: domain.OriginExtracted,
				Status: domain.StatusCurrent,
			},
		},
	})

	_ = mm.AddNode(domain.Node{
		ID:   "contract:payment-grpc",
		Kind: domain.NodeContract,
		Name: "PaymentService",
		Metadata: map[string]any{
			"methods": []string{"Charge"},
		},
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "checkout-repo",
					CommitSHA: "1234567890abcdef",
					Path:      "proto/payment.proto",
					StartLine: 5,
					EndLine:   15,
				},
				Origin: domain.OriginDeclared,
				Status: domain.StatusCurrent,
			},
		},
	})

	_ = mm.AddNode(domain.Node{
		ID:   "infrastructure:redis",
		Kind: domain.NodeInfrastructure,
		Name: "redis-cache",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "checkout-repo",
					CommitSHA: "1234567890abcdef",
					Path:      "docker-compose.yml",
					StartLine: 20,
					EndLine:   30,
				},
				Origin: domain.OriginDeclared,
				Status: domain.StatusStale,
			},
		},
	})

	_ = mm.AddNode(domain.Node{
		ID:   "library:money",
		Kind: domain.NodeLibrary,
		Name: "pkg/money",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "checkout-repo",
					CommitSHA: "1234567890abcdef",
					Path:      "pkg/money/money.go",
					StartLine: 1,
					EndLine:   100,
				},
				Origin: domain.OriginExtracted,
				Status: domain.StatusCurrent,
			},
		},
	})

	_ = mm.AddRelationship(domain.Relationship{
		ID:       "rel:checkout-consumes-payment",
		SourceID: "service:checkout",
		TargetID: "contract:payment-grpc",
		Verb:     domain.VerbConsumes,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "checkout-repo",
					CommitSHA: "1234567890abcdef",
					Path:      "src/client.go",
					StartLine: 12,
					EndLine:   18,
				},
				Origin: domain.OriginExtracted,
				Status: domain.StatusCurrent,
			},
		},
	})

	_ = mm.AddRelationship(domain.Relationship{
		ID:       "rel:checkout-accesses-redis",
		SourceID: "service:checkout",
		TargetID: "infrastructure:redis",
		Verb:     domain.VerbAccesses,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "checkout-repo",
					CommitSHA: "1234567890abcdef",
					Path:      "docker-compose.yml",
					StartLine: 25,
					EndLine:   26,
				},
				Origin: domain.OriginDeclared,
				Status: domain.StatusStale,
			},
		},
	})

	mm.Unsupported = append(mm.Unsupported, domain.UnsupportedBoundary{
		Path:   "frontend/client.ts",
		Reason: "unsupported typescript parser",
		Locator: domain.Locator{
			Repo:      "checkout-repo",
			CommitSHA: "1234567890abcdef",
			Path:      "frontend/client.ts",
			StartLine: 1,
			EndLine:   50,
		},
	})

	return mm
}

func TestExportMarkdown(t *testing.T) {
	mm := createSampleMacroMapForExport()

	md := slices.ExportMarkdown(mm)

	// Check header
	if !strings.Contains(md, "# Target System Architecture: checkout-system") {
		t.Errorf("expected target system header, got:\n%s", md)
	}

	// Check kind sections
	if !strings.Contains(md, "## Services (1)") {
		t.Errorf("expected Services section in md")
	}
	if !strings.Contains(md, "## Infrastructure (1)") {
		t.Errorf("expected Infrastructure section in md")
	}
	if !strings.Contains(md, "## Contracts (1)") {
		t.Errorf("expected Contracts section in md")
	}
	if !strings.Contains(md, "## Libraries (1)") {
		t.Errorf("expected Libraries section in md")
	}

	// Check relationship table
	if !strings.Contains(md, "## Mediated Relationships (2)") {
		t.Errorf("expected Mediated Relationships table in md")
	}
	if !strings.Contains(md, "| `service:checkout` | **consumes** | `contract:payment-grpc` |") {
		t.Errorf("expected rel:checkout-consumes-payment in table")
	}

	// Check Unsupported Boundaries section
	if !strings.Contains(md, "## Unsupported Boundaries & Known Unknowns (1)") {
		t.Errorf("expected Unsupported Boundaries section in md")
	}
	if !strings.Contains(md, "frontend/client.ts") || !strings.Contains(md, "unsupported typescript parser") {
		t.Errorf("expected unsupported typescript boundary in md")
	}

	// Stale notice
	if !strings.Contains(md, "⚠️ NOTICE: Stale evidence detected for `infrastructure:redis`") {
		t.Errorf("expected stale evidence notice in md")
	}
}

func TestExportHTML(t *testing.T) {
	mm := createSampleMacroMapForExport()

	htmlDoc := slices.ExportHTML(mm)

	// Single-file, standalone HTML checks
	if !strings.HasPrefix(htmlDoc, "<!DOCTYPE html>") {
		t.Errorf("expected DOCTYPE html prefix")
	}
	if !strings.Contains(htmlDoc, "<title>Gira-G Architecture Export: checkout-system</title>") {
		t.Errorf("expected title in html")
	}

	// Zero external asset dependencies: no external scripts or stylesheets (http:// or https://)
	if strings.Contains(htmlDoc, "src=\"http") || strings.Contains(htmlDoc, "href=\"http") {
		t.Errorf("HTML export contains external network references")
	}

	// Discrete derivation tags
	if !strings.Contains(htmlDoc, `badge-declared`) || !strings.Contains(htmlDoc, `badge-extracted`) {
		t.Errorf("expected discrete origin badges in HTML")
	}

	// Embedded SVG topological visualization
	if !strings.Contains(htmlDoc, `<svg viewBox=`) || !strings.Contains(htmlDoc, `marker-end="url(#arrow)"`) {
		t.Errorf("expected embedded topological SVG in HTML")
	}

	// Unsupported boundaries section
	if !strings.Contains(htmlDoc, "Unsupported Boundaries & Known Unknowns") {
		t.Errorf("expected Unsupported Boundaries section in HTML")
	}
	if !strings.Contains(htmlDoc, "frontend/client.ts") {
		t.Errorf("expected unsupported TS file in HTML")
	}
}

func TestExport_EmptyMap(t *testing.T) {
	mm := domain.NewMacroMap("empty-sys")
	mm.SchemaVersion = "1.0.0"

	md := slices.ExportMarkdown(mm)
	if !strings.Contains(md, "# Target System Architecture: empty-sys") {
		t.Errorf("expected header for empty map in md")
	}
	if !strings.Contains(md, "(no mediated relationships found)") {
		t.Errorf("expected no relationships note in md")
	}

	htmlDoc := slices.ExportHTML(mm)
	if !strings.Contains(htmlDoc, "Target System: empty-sys") {
		t.Errorf("expected header in empty HTML")
	}
}
