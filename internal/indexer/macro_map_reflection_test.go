package indexer

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// AC6: Reflection-based test inspecting all exported fields of domain.MacroMap.
// Validates that every exported field is explicitly accounted for, deeply cloned in
// cloneMacroMap, and propagated in CanonicalJSON, failing closed if any future field
// is added without explicit support.
func TestMacroMap_AC6_ReflectionCoverage(t *testing.T) {
	// The exhaustive list of all known, supported exported fields of domain.MacroMap.
	expectedFields := map[string]string{
		"SchemaVersion":         "schema_version",
		"TargetSystem":          "target_system",
		"Nodes":                 "nodes",
		"Relationships":         "relationships",
		"Unsupported":           "unsupported",
		"UnsupportedConstructs": "unsupported_constructs",
		"UnresolvedRefs":        "unresolved_refs",
		"Conflicts":             "conflicts",
		"SourceIdentity":        "source_identity",
	}

	mmType := reflect.TypeOf(domain.MacroMap{})
	if mmType.Kind() != reflect.Struct {
		t.Fatalf("domain.MacroMap must be a struct, got %s", mmType.Kind())
	}

	actualExportedFields := make(map[string]reflect.StructField)
	for i := 0; i < mmType.NumField(); i++ {
		field := mmType.Field(i)
		// Exported fields have empty PkgPath
		if field.PkgPath == "" {
			actualExportedFields[field.Name] = field
		}
	}

	// 1. Verify that no new field was added without updating this test and clone/serialization logic.
	for actualName, field := range actualExportedFields {
		jsonTag, exists := expectedFields[actualName]
		if !exists {
			t.Fatalf("AC6 failure: domain.MacroMap has new exported field %q not registered in expectedFields. "+
				"You MUST update cloneMacroMap, CanonicalJSON, and this test to support the new field without silent data loss.", actualName)
		}

		tag := field.Tag.Get("json")
		cleanTag := strings.Split(tag, ",")[0]
		if cleanTag != jsonTag {
			t.Errorf("field %q json tag mismatch: got %q, want %q", actualName, cleanTag, jsonTag)
		}
	}

	// 2. Verify that no field in expectedFields was removed or renamed
	for expectedName := range expectedFields {
		if _, exists := actualExportedFields[expectedName]; !exists {
			t.Fatalf("AC6 failure: expected field %q missing from domain.MacroMap struct", expectedName)
		}
	}

	// 3. Construct a fully populated MacroMap where EVERY exported field has non-zero, distinct data.
	fixedTime := time.Date(2026, 10, 19, 14, 0, 0, 0, time.UTC)
	fullMap := &domain.MacroMap{
		SchemaVersion: 2,
		TargetSystem:  "reflection-target-system",
		Nodes: []domain.Node{
			{
				ID:   "service:ref-node",
				Kind: domain.NodeService,
				Name: "RefNode",
				Metadata: map[string]any{
					"key": "val",
				},
				Evidence: []domain.EvidenceRecord{
					{
						Locator: domain.Locator{
							Repo:      "repo1",
							CommitSHA: "abc1112223334445556667778889990001112223",
							Path:      "main.go",
						},
						Origin:       domain.OriginExtracted,
						Status:       domain.StatusCurrent,
						DiscoveredAt: fixedTime,
					},
				},
			},
		},
		Relationships: []domain.Relationship{
			{
				ID:       "rel:ref-node-to-ref-node",
				SourceID: "service:ref-node",
				TargetID: "service:ref-node",
				Verb:     domain.VerbProvides,
				Metadata: map[string]any{
					"protocol": "grpc",
				},
				Evidence: []domain.EvidenceRecord{
					{
						RuleID: "test-rule",
					},
				},
			},
		},
		Unsupported: []domain.UnsupportedBoundary{
			{
				Path:         "legacy.unknown",
				Reason:       "unsupported extension",
				UnknownJoins: []string{"ContractA"},
			},
		},
		UnsupportedConstructs: []string{"parsed/file.go"},
		UnresolvedRefs: []domain.UnresolvedReference{
			{
				ID:        "unresolved:1",
				MissingID: "missing:node",
				Evidence: []domain.EvidenceRecord{
					{RuleID: "unresolved-rule"},
				},
			},
		},
		Conflicts: []domain.ConflictRecord{
			{
				ID:          "conflict:1",
				SubjectID:   "service:ref-node",
				Field:       "kind",
				DeclaredVal: "Service",
				PreviousVal: "Library",
				Evidence: []domain.EvidenceRecord{
					{RuleID: "conflict-rule"},
				},
			},
		},
		SourceIdentity: &domain.SourceIdentity{
			SchemaVersion:  2,
			ManifestSHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			Extractors: domain.ExtractorList{
				{ID: "ref-extractor", Version: "1.0.0"},
			},
			Sources: []domain.SourceRepoIdentity{
				{
					RepoID:    "repo1",
					CommitSHA: "abc1112223334445556667778889990001112223",
					SHA:       "abc1112223334445556667778889990001112223",
					URLSHA256: "1111222233334444555566667777888899990000aaaabbbbccccddddeeeeffff",
				},
			},
		},
	}

	// 4. Verify cloneMacroMap deeply propagates all fields
	cloned, err := cloneMacroMap(fullMap)
	if err != nil {
		t.Fatalf("cloneMacroMap failed: %v", err)
	}
	if cloned == nil {
		t.Fatal("cloneMacroMap returned nil")
	}

	valOrig := reflect.ValueOf(fullMap).Elem()
	valCloned := reflect.ValueOf(cloned).Elem()

	for fieldName := range expectedFields {
		fOrig := valOrig.FieldByName(fieldName)
		fCloned := valCloned.FieldByName(fieldName)

		if !fOrig.IsValid() || !fCloned.IsValid() {
			t.Fatalf("field %q invalid in reflect", fieldName)
		}

		// Ensure cloned field is not zero value
		if fCloned.IsZero() {
			t.Fatalf("AC6 violation: cloned field %q is zero-value", fieldName)
		}

		// For pointer or slice fields, ensure memory addresses are distinct (deep copy)
		switch fOrig.Kind() {
		case reflect.Pointer:
			if fOrig.Pointer() == fCloned.Pointer() {
				t.Fatalf("AC6 violation: field %q was shallow-copied (pointer aliasing)", fieldName)
			}
		case reflect.Slice:
			if fOrig.Len() > 0 && fOrig.Pointer() == fCloned.Pointer() {
				t.Fatalf("AC6 violation: field %q slice backing array was aliased", fieldName)
			}
		}
	}

	// 5. Verify CanonicalJSON serializes every field
	canonJSON, err := fullMap.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}

	var jsonMap map[string]any
	if err := json.Unmarshal(canonJSON, &jsonMap); err != nil {
		t.Fatalf("unmarshaling CanonicalJSON to map failed: %v", err)
	}

	for fieldName, jsonTag := range expectedFields {
		val, exists := jsonMap[jsonTag]
		if !exists {
			t.Fatalf("AC6 violation: CanonicalJSON output missing key %q for struct field %q", jsonTag, fieldName)
		}
		if val == nil {
			t.Fatalf("AC6 violation: CanonicalJSON output has nil value for key %q", jsonTag)
		}
	}
}

// Negative test proving that reflection check catches untracked fields
func TestMacroMap_AC6_NegativeDetection(t *testing.T) {
	// Simulate an unhandled hypothetical field
	dummyFields := map[string]string{
		"SchemaVersion": "schema_version",
		// Omit TargetSystem, Nodes, etc.
	}

	mmType := reflect.TypeOf(domain.MacroMap{})
	var unhandled []string
	for i := 0; i < mmType.NumField(); i++ {
		field := mmType.Field(i)
		if field.PkgPath == "" {
			if _, ok := dummyFields[field.Name]; !ok {
				unhandled = append(unhandled, field.Name)
			}
		}
	}

	if len(unhandled) == 0 {
		t.Fatal("negative test failed: expected unhandled fields to be detected")
	}
}
