package jev

import (
	"bytes"
	"context"
	"path/filepath"
	"regexp"
	"strings"
)

// LocalClassifier provides a 100% offline, zero-token, sub-millisecond semantic classifier floor.
type LocalClassifier struct{}

func NewLocalClassifier() *LocalClassifier {
	return &LocalClassifier{}
}

func (c *LocalClassifier) Name() string {
	return "jev_local"
}

func (c *LocalClassifier) ClassifyIntent(ctx context.Context, query string) IntentResult {
	qLower := strings.ToLower(strings.TrimSpace(query))
	wantDiagram := strings.Contains(qLower, "diagram") || strings.Contains(qLower, "fluxo") ||
		strings.Contains(qLower, "flow") || strings.Contains(qLower, "desenhe") ||
		strings.Contains(qLower, "grafo") || strings.Contains(qLower, "graph")

	// 1. Purpose / Objective
	purposeRegex := regexp.MustCompile(`(?i)\b(objetivo|prop[oó]sito|purpose|o que faz|what does|what is|sobre|about|vis[aã]o|overview)\b`)
	if purposeRegex.MatchString(qLower) {
		return IntentResult{
			Intent:      IntentPurpose,
			Confidence:  0.92,
			WantDiagram: wantDiagram,
			Reason:      "Matched project purpose/overview keywords",
			Engine:      c.Name(),
		}
	}

	// 2. Setup / Quickstart / How to run
	quickstartRegex := regexp.MustCompile(`(?i)\b(como usar|como uso|utiliz\w*|usar|uso|inicializ\w*|iniciar|in[ií]cio|setup|quickstart|come[çc]\w*|rodar|start|run|execut\w*|instalar|install)\b`)
	if quickstartRegex.MatchString(qLower) {
		return IntentResult{
			Intent:      IntentQuickstart,
			Confidence:  0.90,
			WantDiagram: wantDiagram,
			Reason:      "Matched quickstart/setup/usage keywords",
			Engine:      c.Name(),
		}
	}

	// 3. Impact / Blast Radius
	impactRegex := regexp.MustCompile(`(?i)\b(impacto|impact|o que quebra|quebrar|alterar|mudar|modificar|blast radius|efeito colateral|side effect)\b`)
	if impactRegex.MatchString(qLower) {
		return IntentResult{
			Intent:      IntentImpact,
			Confidence:  0.88,
			WantDiagram: wantDiagram,
			Reason:      "Matched impact/blast radius keywords",
			Engine:      c.Name(),
		}
	}

	// 4. Diagram requested explicitly
	if wantDiagram {
		return IntentResult{
			Intent:      IntentComponentDetail,
			Confidence:  0.80,
			WantDiagram: true,
			Reason:      "Explicit request for visual architecture diagram",
			Engine:      c.Name(),
		}
	}

	// 5. Default General
	return IntentResult{
		Intent:      IntentGeneral,
		Confidence:  0.60,
		WantDiagram: false,
		Reason:      "General architectural query",
		Engine:      c.Name(),
	}
}

func (c *LocalClassifier) ClassifyDocument(ctx context.Context, path string, content []byte) DocumentResult {
	normalized := strings.ToLower(filepath.ToSlash(path))
	cLower := strings.ToLower(string(content))

	// Path heuristics
	if strings.Contains(normalized, "/adr/") || strings.Contains(normalized, "/decisions/") || strings.HasPrefix(normalized, "adr-") {
		return DocumentResult{
			Role:       RolePolicy,
			Confidence: 0.95,
			Reason:     "Path indicates an Architectural Decision Record (ADR)",
			Engine:     c.Name(),
		}
	}

	if strings.Contains(normalized, "/tasks/") || strings.Contains(normalized, "/issues/") || strings.HasPrefix(normalized, "task-") {
		return DocumentResult{
			Role:       RoleTask,
			Confidence: 0.95,
			Reason:     "Path indicates a work task or issue definition",
			Engine:     c.Name(),
		}
	}

	if strings.Contains(normalized, "/contracts/") || strings.Contains(normalized, "/api/") || strings.Contains(normalized, "/specs/") {
		return DocumentResult{
			Role:       RoleContract,
			Confidence: 0.90,
			Reason:     "Path indicates a contract or specification boundary",
			Engine:     c.Name(),
		}
	}

	// Content heuristics
	if strings.Contains(cLower, "status: accepted") || strings.Contains(cLower, "## decision") || strings.Contains(cLower, "## contexto") {
		return DocumentResult{
			Role:       RolePolicy,
			Confidence: 0.85,
			Reason:     "Content follows standard ADR decision structure",
			Engine:     c.Name(),
		}
	}

	if strings.Contains(cLower, "done_when:") || strings.Contains(cLower, "## passos") || bytes.Contains(content, []byte("- [ ]")) {
		return DocumentResult{
			Role:       RoleTask,
			Confidence: 0.85,
			Reason:     "Content follows task plan execution structure",
			Engine:     c.Name(),
		}
	}

	return DocumentResult{
		Role:       RoleDocumentation,
		Confidence: 0.70,
		Reason:     "General documentation markdown file",
		Engine:     c.Name(),
	}
}

func (c *LocalClassifier) ClassifyDrift(ctx context.Context, specSummary string, diffSummary string) DriftResult {
	dLower := strings.ToLower(diffSummary)
	sLower := strings.ToLower(specSummary)

	if len(strings.TrimSpace(diffSummary)) == 0 {
		return DriftResult{
			Severity:   SeverityNone,
			Confidence: 1.0,
			Reason:     "Zero diff detected",
			Engine:     c.Name(),
		}
	}

	// Breaking violations: forbidden imports or broken public interfaces
	if strings.Contains(sLower, "proib") || strings.Contains(sLower, "forbid") || strings.Contains(sLower, "cannot import") {
		if strings.Contains(dLower, "+import") || strings.Contains(dLower, "+ package") {
			return DriftResult{
				Severity:   SeverityBreaking,
				Confidence: 0.95,
				Reason:     "Diff introduces an import into a restricted boundary specified by architecture policy",
				Engine:     c.Name(),
			}
		}
	}

	// Cosmetic diffs: comments, test data or whitespace
	if !strings.Contains(dLower, "func ") && !strings.Contains(dLower, "type ") && !strings.Contains(dLower, "interface ") {
		return DriftResult{
			Severity:   SeverityCosmetic,
			Confidence: 0.85,
			Reason:     "Diff does not modify public types, functions or interfaces",
			Engine:     c.Name(),
		}
	}

	return DriftResult{
		Severity:   SeverityWarning,
		Confidence: 0.75,
		Reason:     "Code signatures modified; requires verification against contracts",
		Engine:     c.Name(),
	}
}
