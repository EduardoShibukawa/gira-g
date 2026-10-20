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
	if cloned.SourceIdentity != nil {
		t.Fatalf("SourceIdentity = %#v, want nil", cloned.SourceIdentity)
	}
}

func TestCloneMacroMapPreservesSourceIdentity(t *testing.T) {
	input := &domain.MacroMap{
		SchemaVersion: 2,
		TargetSystem:  "demo",
		SourceIdentity: &domain.SourceIdentity{
			SchemaVersion:  2,
			ManifestSHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			Extractors: domain.ExtractorList{
				{ID: "go-extractor", Version: "1.0.0"},
			},
			Sources: []domain.SourceRepoIdentity{
				{
					RepoID:    "app",
					CommitSHA: "abc",
					SHA:       "abc",
					URLSHA256: "def",
				},
			},
		},
	}

	cloned, err := cloneMacroMap(input)
	if err != nil {
		t.Fatalf("cloneMacroMap() error = %v", err)
	}
	if cloned.SourceIdentity == nil {
		t.Fatal("SourceIdentity was not cloned")
	}
	if cloned.SourceIdentity == input.SourceIdentity {
		t.Fatal("SourceIdentity pointer was aliased, want deep copy")
	}
	if cloned.SourceIdentity.ManifestSHA256 != input.SourceIdentity.ManifestSHA256 {
		t.Fatalf("ManifestSHA256 = %s, want %s", cloned.SourceIdentity.ManifestSHA256, input.SourceIdentity.ManifestSHA256)
	}
	if len(cloned.SourceIdentity.Extractors) != 1 || cloned.SourceIdentity.Extractors[0].ID != "go-extractor" {
		t.Fatalf("Extractors not preserved: %+v", cloned.SourceIdentity.Extractors)
	}
	if len(cloned.SourceIdentity.Sources) != 1 || cloned.SourceIdentity.Sources[0].RepoID != "app" {
		t.Fatalf("Sources not preserved: %+v", cloned.SourceIdentity.Sources)
	}

	// Slices must not alias
	cloned.SourceIdentity.Extractors[0].ID = "mutated-extractor"
	cloned.SourceIdentity.Sources[0].RepoID = "mutated-repo"
	if input.SourceIdentity.Extractors[0].ID != "go-extractor" {
		t.Fatal("cloned extractors slice aliased original")
	}
	if input.SourceIdentity.Sources[0].RepoID != "app" {
		t.Fatal("cloned sources slice aliased original")
	}
}
