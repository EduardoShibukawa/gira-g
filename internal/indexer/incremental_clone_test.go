package indexer

import (
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestCloneMacroMapPreservesUnresolvedRefsAndConflicts(t *testing.T) {
	input := &domain.MacroMap{
		SchemaVersion: "1.0.0",
		TargetSystem:  "demo",
		UnresolvedRefs: []domain.UnresolvedReference{
			{
				ID:         "unresolved:missing-contract",
				RelationID: "rel:service-contract",
				SourceID:   "service:api",
				TargetID:   "contract:missing",
				MissingID:  "contract:missing",
				Verb:       domain.VerbConsumes,
				Evidence: []domain.EvidenceRecord{
					{RuleID: "unresolved-rule", Rationale: "target does not exist"},
				},
			},
		},
		Conflicts: []domain.ConflictRecord{
			{
				ID:          "conflict:service-name",
				SubjectID:   "service:api",
				Kind:        "field_mismatch",
				Field:       "name",
				DeclaredVal: "api-v2",
				PreviousVal: "api-v1",
				Evidence: []domain.EvidenceRecord{
					{RuleID: "conflict-rule", Rationale: "declaration differs from extraction"},
				},
			},
		},
	}

	cloned, err := cloneMacroMap(input)
	if err != nil {
		t.Fatalf("cloneMacroMap() error = %v", err)
	}
	if cloned == nil {
		t.Fatal("cloneMacroMap() returned nil")
	}
	if len(cloned.UnresolvedRefs) != 1 {
		t.Fatalf("UnresolvedRefs length = %d, want 1", len(cloned.UnresolvedRefs))
	}
	if got := cloned.UnresolvedRefs[0]; got.ID != input.UnresolvedRefs[0].ID ||
		got.RelationID != input.UnresolvedRefs[0].RelationID ||
		got.SourceID != input.UnresolvedRefs[0].SourceID ||
		got.TargetID != input.UnresolvedRefs[0].TargetID ||
		got.MissingID != input.UnresolvedRefs[0].MissingID ||
		got.Verb != input.UnresolvedRefs[0].Verb ||
		len(got.Evidence) != 1 ||
		got.Evidence[0].RuleID != "unresolved-rule" {
		t.Fatalf("UnresolvedRefs not preserved: got %+v", got)
	}
	if len(cloned.Conflicts) != 1 {
		t.Fatalf("Conflicts length = %d, want 1", len(cloned.Conflicts))
	}
	if got := cloned.Conflicts[0]; got.ID != input.Conflicts[0].ID ||
		got.SubjectID != input.Conflicts[0].SubjectID ||
		got.Kind != input.Conflicts[0].Kind ||
		got.Field != input.Conflicts[0].Field ||
		got.DeclaredVal != input.Conflicts[0].DeclaredVal ||
		got.PreviousVal != input.Conflicts[0].PreviousVal ||
		len(got.Evidence) != 1 ||
		got.Evidence[0].RuleID != "conflict-rule" {
		t.Fatalf("Conflicts not preserved: got %+v", got)
	}

	// Nested evidence slices must not alias the source after cloning.
	cloned.UnresolvedRefs[0].Evidence[0].RuleID = "changed-unresolved"
	cloned.Conflicts[0].Evidence[0].RuleID = "changed-conflict"
	if input.UnresolvedRefs[0].Evidence[0].RuleID != "unresolved-rule" {
		t.Fatal("cloned unresolved-ref evidence aliases source evidence")
	}
	if input.Conflicts[0].Evidence[0].RuleID != "conflict-rule" {
		t.Fatal("cloned conflict evidence aliases source evidence")
	}
}

func TestCloneMacroMapPreservesNilDiagnosticSlices(t *testing.T) {
	input := &domain.MacroMap{SchemaVersion: "1.0.0", TargetSystem: "demo"}
	cloned, err := cloneMacroMap(input)
	if err != nil {
		t.Fatalf("cloneMacroMap() error = %v", err)
	}
	if cloned.UnresolvedRefs != nil {
		t.Fatalf("UnresolvedRefs = %#v, want nil", cloned.UnresolvedRefs)
	}
	if cloned.Conflicts != nil {
		t.Fatalf("Conflicts = %#v, want nil", cloned.Conflicts)
	}
}
