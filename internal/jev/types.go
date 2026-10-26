package jev

import (
	"context"
)

// IntentType represents the classified intent of a developer natural language query.
type IntentType string

const (
	IntentPurpose         IntentType = "purpose"
	IntentQuickstart      IntentType = "quickstart"
	IntentImpact          IntentType = "impact"
	IntentComponentDetail IntentType = "component_detail"
	IntentGeneral         IntentType = "general"
)

// DocumentRole represents the architectural role of a non-deterministic markdown or spec file.
type DocumentRole string

const (
	RolePolicy        DocumentRole = "policy"        // Architectural Decision Records (ADRs)
	RoleTask          DocumentRole = "task"          // Task plans, issues, work items
	RoleContract      DocumentRole = "contract"      // Interface or API specifications
	RoleDocumentation DocumentRole = "documentation" // General manuals, guides, overviews
)

// DriftSeverity represents the evaluated impact of a discrepancy between code and intention.
type DriftSeverity string

const (
	SeverityNone     DriftSeverity = "none"     // No divergence
	SeverityCosmetic DriftSeverity = "cosmetic" // Whitespace, comments, or non-breaking renaming
	SeverityWarning  DriftSeverity = "warning"  // Behavioral change without contract violation
	SeverityBreaking DriftSeverity = "breaking" // Direct contract violation or unauthorized cross-boundary dependency
)

// IntentResult carries the structured classification of a user query.
type IntentResult struct {
	Intent      IntentType `json:"intent"`
	Confidence  float64    `json:"confidence"` // 0.0 to 1.0
	TargetID    string     `json:"target_id,omitempty"`
	WantDiagram bool       `json:"want_diagram"`
	Reason      string     `json:"reason"`
	Engine      string     `json:"engine"` // "jev_local" or "jev_runner"
}

// DocumentResult carries the structured classification of a document.
type DocumentResult struct {
	Role       DocumentRole `json:"role"`
	Confidence float64      `json:"confidence"`
	Title      string       `json:"title,omitempty"`
	Tags       []string     `json:"tags,omitempty"`
	Reason     string       `json:"reason"`
	Engine     string       `json:"engine"`
}

// DriftResult carries the evaluated severity of code divergence from a specification.
type DriftResult struct {
	Severity   DriftSeverity `json:"severity"`
	Confidence float64       `json:"confidence"`
	Reason     string        `json:"reason"`
	Engine     string        `json:"engine"`
}

// Classifier is the interface for non-deterministic semantic triage and decision making.
type Classifier interface {
	Name() string
	ClassifyIntent(ctx context.Context, query string) IntentResult
	ClassifyDocument(ctx context.Context, path string, content []byte) DocumentResult
	ClassifyDrift(ctx context.Context, specSummary string, diffSummary string) DriftResult
}
