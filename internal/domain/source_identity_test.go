package domain

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// AC4: source_identity includes all relevant sources, deterministically ordered, without raw URLs.
func TestSourceIdentity_DeterministicOrdering_NoRawURL(t *testing.T) {
	rawManifest := []byte("target_system:\n  name: demo\nrepositories:\n  - id: zebra\n    url: https://user:secret@github.com/org/zebra\n    sha: 1111111111111111111111111111111111111111\n  - id: alpha\n    url: https://github.com/org/alpha.git\n    sha: 2222222222222222222222222222222222222222\n")

	// Notice: repos declared out of alphabetical order, and zebra has embedded credentials
	si := &SourceIdentity{
		SchemaVersion:  2,
		ManifestSHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		Extractors: ExtractorList{
			{ID: "proto-extractor", Version: "1.0.0"},
			{ID: "go-extractor", Version: "1.0.0"},
		},
		Sources: []SourceRepoIdentity{
			{
				RepoID:    "zebra",
				CommitSHA: "1111111111111111111111111111111111111111",
				SHA:       "1111111111111111111111111111111111111111",
				URLSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			{
				RepoID:    "alpha",
				CommitSHA: "2222222222222222222222222222222222222222",
				SHA:       "2222222222222222222222222222222222222222",
				URLSHA256: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			},
		},
	}

	mm := NewMacroMapV2("demo", si)
	if err := mm.Validate(); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	jsonBytes, err := mm.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Ensure no raw URLs leak
	if strings.Contains(jsonStr, "user:secret") || strings.Contains(jsonStr, "https://") {
		t.Fatalf("AC4 violation: raw URL or credentials leaked in JSON: %s", jsonStr)
	}

	// Ensure url_sha256 is present
	if !strings.Contains(jsonStr, `"url_sha256"`) {
		t.Fatalf("AC4 violation: url_sha256 missing from serialized JSON: %s", jsonStr)
	}

	// Unmarshal to inspect sorted order
	var roundtrip MacroMap
	if err := json.Unmarshal(jsonBytes, &roundtrip); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if roundtrip.SourceIdentity == nil {
		t.Fatal("SourceIdentity missing after unmarshal")
	}

	// Sources must be sorted deterministically by repo_id
	if len(roundtrip.SourceIdentity.Sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(roundtrip.SourceIdentity.Sources))
	}
	if roundtrip.SourceIdentity.Sources[0].RepoID != "alpha" || roundtrip.SourceIdentity.Sources[1].RepoID != "zebra" {
		t.Fatalf("AC4 violation: sources not sorted deterministically by repo_id: %+v", roundtrip.SourceIdentity.Sources)
	}

	// Extractors must be sorted deterministically by ID
	if len(roundtrip.SourceIdentity.Extractors) != 2 {
		t.Fatalf("expected 2 extractors, got %d", len(roundtrip.SourceIdentity.Extractors))
	}
	if roundtrip.SourceIdentity.Extractors[0].ID != "go-extractor" || roundtrip.SourceIdentity.Extractors[1].ID != "proto-extractor" {
		t.Fatalf("AC4 violation: extractors not sorted deterministically by ID: %+v", roundtrip.SourceIdentity.Extractors)
	}

	_ = rawManifest
}

// AC5: Schema 2 and source_identity survive CanonicalJSON, clone, store roundtrip, and detached snapshot.
func TestMacroMap_AC5_Roundtrip_Clone_CanonicalJSON_Snapshot(t *testing.T) {
	fixedTime := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	si := &SourceIdentity{
		SchemaVersion:  2,
		ManifestSHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		Extractors: ExtractorList{
			{ID: "go-ast-extractor", Version: "1.0.0"},
			{ID: "proto-extractor", Version: "1.0.0"},
		},
		Sources: []SourceRepoIdentity{
			{
				RepoID:    "service-repo",
				CommitSHA: "abc1112223334445556667778889990001112223",
				SHA:       "abc1112223334445556667778889990001112223",
				PathScope: "services/",
				URLSHA256: "1111222233334444555566667777888899990000aaaabbbbccccddddeeeeffff",
			},
		},
	}

	orig := NewMacroMapV2("test-system", si)
	_ = orig.AddNode(Node{
		ID:   "service:api",
		Kind: NodeService,
		Name: "api-service",
		Evidence: []EvidenceRecord{
			{
				Locator: Locator{
					Repo:      "service-repo",
					CommitSHA: "abc1112223334445556667778889990001112223",
					Path:      "services/api/main.go",
				},
				Origin:       OriginExtracted,
				Status:       StatusCurrent,
				DiscoveredAt: fixedTime,
			},
		},
	})
	orig.UnresolvedRefs = []UnresolvedReference{
		{
			ID:        "unresolved:contract",
			MissingID: "contract:missing",
		},
	}
	orig.Conflicts = []ConflictRecord{
		{
			ID:          "conflict:name",
			SubjectID:   "service:api",
			DeclaredVal: "api-v2",
			PreviousVal: "api-v1",
		},
	}

	// 1. CanonicalJSON survival
	data, err := orig.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}

	// Ensure schema_version is serialized as integer 2
	if !bytes.Contains(data, []byte(`"schema_version": 2`)) {
		t.Fatalf("AC5 violation: CanonicalJSON does not serialize schema_version as integer 2:\n%s", string(data))
	}

	var jsonRoundtrip MacroMap
	if err := json.Unmarshal(data, &jsonRoundtrip); err != nil {
		t.Fatalf("unmarshaling CanonicalJSON failed: %v", err)
	}

	if jsonRoundtrip.SchemaVersion != 2 {
		t.Fatalf("AC5 violation: expected schema_version 2 (int), got: %#v (%T)", jsonRoundtrip.SchemaVersion, jsonRoundtrip.SchemaVersion)
	}
	if jsonRoundtrip.SourceIdentity == nil {
		t.Fatal("AC5 violation: source_identity missing after CanonicalJSON roundtrip")
	}
	if jsonRoundtrip.SourceIdentity.SchemaVersion != 2 {
		t.Fatalf("AC5 violation: source_identity schema_version = %d, want 2", jsonRoundtrip.SourceIdentity.SchemaVersion)
	}
	if jsonRoundtrip.SourceIdentity.ManifestSHA256 != si.ManifestSHA256 {
		t.Fatalf("AC5 violation: manifest_sha256 mismatch")
	}
	if len(jsonRoundtrip.UnresolvedRefs) != 1 || jsonRoundtrip.UnresolvedRefs[0].ID != "unresolved:contract" {
		t.Fatalf("AC5 violation: UnresolvedRefs lost in CanonicalJSON: %+v", jsonRoundtrip.UnresolvedRefs)
	}
	if len(jsonRoundtrip.Conflicts) != 1 || jsonRoundtrip.Conflicts[0].ID != "conflict:name" {
		t.Fatalf("AC5 violation: Conflicts lost in CanonicalJSON: %+v", jsonRoundtrip.Conflicts)
	}

	// 2. Detached snapshot file roundtrip
	tmpDir := t.TempDir()
	snapshotPath := filepath.Join(tmpDir, "detached-snapshot.json")
	if err := os.WriteFile(snapshotPath, data, 0644); err != nil {
		t.Fatalf("failed writing detached snapshot: %v", err)
	}

	readBytes, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatalf("failed reading detached snapshot: %v", err)
	}
	var snapshotMap MacroMap
	if err := json.Unmarshal(readBytes, &snapshotMap); err != nil {
		t.Fatalf("failed unmarshaling detached snapshot: %v", err)
	}
	if snapshotMap.SchemaVersion != 2 {
		t.Fatalf("AC5 violation: detached snapshot SchemaVersion = %#v, want 2", snapshotMap.SchemaVersion)
	}
	if snapshotMap.SourceIdentity == nil || snapshotMap.SourceIdentity.ManifestSHA256 != si.ManifestSHA256 {
		t.Fatalf("AC5 violation: detached snapshot source_identity not preserved")
	}
}
