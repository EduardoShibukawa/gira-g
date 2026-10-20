package store

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

// AC1: Injective encoding tests and fuzzing
func TestInjectiveEncoding_Injectivity(t *testing.T) {
	// ("ab", "c") != ("a", "bc")
	e1 := NewInjectiveEncoder("test")
	e1.WriteString("ab")
	e1.WriteString("c")

	e2 := NewInjectiveEncoder("test")
	e2.WriteString("a")
	e2.WriteString("bc")

	if bytes.Equal(e1.Bytes(), e2.Bytes()) {
		t.Fatalf("InjectiveEncoder collision between ('ab', 'c') and ('a', 'bc'): %x", e1.Bytes())
	}
	if e1.SumHex() == e2.SumHex() {
		t.Fatalf("InjectiveEncoder hash collision: %s == %s", e1.SumHex(), e2.SumHex())
	}

	// Empty string boundaries: ("", "abc") != ("a", "bc") != ("abc", "")
	e3 := NewInjectiveEncoder("test")
	e3.WriteString("")
	e3.WriteString("abc")

	e4 := NewInjectiveEncoder("test")
	e4.WriteString("abc")
	e4.WriteString("")

	if bytes.Equal(e3.Bytes(), e4.Bytes()) {
		t.Fatalf("InjectiveEncoder collision with empty fields: %x == %x", e3.Bytes(), e4.Bytes())
	}
	if bytes.Equal(e1.Bytes(), e3.Bytes()) || bytes.Equal(e2.Bytes(), e3.Bytes()) {
		t.Fatal("InjectiveEncoder collision across empty string variants")
	}

	// Roundtrip through InjectiveDecoder
	dec := NewInjectiveDecoder(e1.Bytes())
	tag, err := dec.ReadString()
	if err != nil || tag != "test" {
		t.Fatalf("failed decoding tag: got %q, err: %v", tag, err)
	}
	s1, err := dec.ReadString()
	if err != nil || s1 != "ab" {
		t.Fatalf("failed decoding s1: got %q, err: %v", s1, err)
	}
	s2, err := dec.ReadString()
	if err != nil || s2 != "c" {
		t.Fatalf("failed decoding s2: got %q, err: %v", s2, err)
	}
	if dec.Remaining() != 0 {
		t.Fatalf("unexpected unread bytes remaining: %d", dec.Remaining())
	}
}

func FuzzInjectiveEncoding(f *testing.F) {
	f.Add([]byte("field1"), []byte("field2"))
	f.Add([]byte(""), []byte("value"))
	f.Add([]byte("prefix"), []byte(""))
	f.Add([]byte("a"), []byte("bc"))
	f.Add([]byte("ab"), []byte("c"))

	f.Fuzz(func(t *testing.T, a, b []byte) {
		enc := NewInjectiveEncoder("fuzz")
		enc.WriteBytes(a)
		enc.WriteBytes(b)

		dec := NewInjectiveDecoder(enc.Bytes())
		tag, err := dec.ReadString()
		if err != nil || tag != "fuzz" {
			t.Fatalf("decoding tag failed: %v", err)
		}
		gotA, err := dec.ReadBytes()
		if err != nil {
			t.Fatalf("decoding a failed: %v", err)
		}
		gotB, err := dec.ReadBytes()
		if err != nil {
			t.Fatalf("decoding b failed: %v", err)
		}
		if !bytes.Equal(a, gotA) || !bytes.Equal(b, gotB) {
			t.Fatalf("roundtrip mismatch: got (%q, %q), want (%q, %q)", gotA, gotB, a, b)
		}
		if dec.Remaining() != 0 {
			t.Fatalf("remaining bytes after full decode: %d", dec.Remaining())
		}
	})
}

// AC1: Fingerprint algorithm contraexamples table tests
func TestContraexamples_AltersSourceKey(t *testing.T) {
	baseManifest := []byte("target_system:\n  name: demo\nrepositories:\n  - id: core\n    url: https://github.com/example/core\n    sha: 1111111111111111111111111111111111111111\n")
	baseSources := []domain.SourceRepoIdentity{
		{
			RepoID:    "core",
			URLSHA256: gitcache.RepoHash("https://github.com/example/core"),
			CommitSHA: "1111111111111111111111111111111111111111",
		},
	}

	baseKey := ComputeSourceKey(SourceKeyOptions{
		SchemaVersion:   2,
		ManifestLocator: "gira.yaml",
		ManifestBytes:   baseManifest,
		Sources:         baseSources,
	})

	// Contraexample 1: Normalized locator (relative vs different locator path)
	diffLocatorKey := ComputeSourceKey(SourceKeyOptions{
		SchemaVersion:   2,
		ManifestLocator: "sub/gira.yaml",
		ManifestBytes:   baseManifest,
		Sources:         baseSources,
	})
	if diffLocatorKey == baseKey {
		t.Fatalf("contraexample 1 failed: changing manifest locator must alter source_key")
	}

	// Contraexample 2: Comment added to manifest
	commentManifest := []byte("# leading comment\ntarget_system:\n  name: demo\nrepositories:\n  - id: core\n    url: https://github.com/example/core\n    sha: 1111111111111111111111111111111111111111\n")
	commentKey := ComputeSourceKey(SourceKeyOptions{
		SchemaVersion:   2,
		ManifestLocator: "gira.yaml",
		ManifestBytes:   commentManifest,
		Sources:         baseSources,
	})
	if commentKey == baseKey {
		t.Fatalf("contraexample 2 failed: adding comment must alter source_key")
	}

	// Contraexample 3: Same SHA with different path_scope or repo_id
	scopedSources := []domain.SourceRepoIdentity{
		{
			RepoID:    "core",
			URLSHA256: gitcache.RepoHash("https://github.com/example/core"),
			CommitSHA: "1111111111111111111111111111111111111111",
			PathScope: "src/backend",
		},
	}
	scopedKey := ComputeSourceKey(SourceKeyOptions{
		SchemaVersion:   2,
		ManifestLocator: "gira.yaml",
		ManifestBytes:   baseManifest,
		Sources:         scopedSources,
	})
	if scopedKey == baseKey {
		t.Fatalf("contraexample 3 failed: changing path_scope must alter source_key")
	}

	diffRepoSources := []domain.SourceRepoIdentity{
		{
			RepoID:    "other-core",
			URLSHA256: gitcache.RepoHash("https://github.com/example/core"),
			CommitSHA: "1111111111111111111111111111111111111111",
		},
	}
	diffRepoKey := ComputeSourceKey(SourceKeyOptions{
		SchemaVersion:   2,
		ManifestLocator: "gira.yaml",
		ManifestBytes:   baseManifest,
		Sources:         diffRepoSources,
	})
	if diffRepoKey == baseKey {
		t.Fatalf("contraexample 3b failed: changing repo_id must alter source_key")
	}

	// Contraexample 4: Same name in different projects (different clone URL hash)
	diffURLSources := []domain.SourceRepoIdentity{
		{
			RepoID:    "core",
			URLSHA256: gitcache.RepoHash("https://github.com/other-org/core"),
			CommitSHA: "1111111111111111111111111111111111111111",
		},
	}
	diffURLKey := ComputeSourceKey(SourceKeyOptions{
		SchemaVersion:   2,
		ManifestLocator: "gira.yaml",
		ManifestBytes:   baseManifest,
		Sources:         diffURLSources,
	})
	if diffURLKey == baseKey {
		t.Fatalf("contraexample 4 failed: different URL hash must alter source_key")
	}
}

// AC2: Comment, BOM, and EOL alterations change raw manifest hash and source_key
func TestManifestRawBytesHash_Comment_BOM_EOL(t *testing.T) {
	manifestLF := []byte("target_system:\n  name: demo\nrepositories:\n  - id: app\n    url: https://github.com/example/app\n    sha: aabbcc\n")
	manifestCRLF := []byte("target_system:\r\n  name: demo\r\nrepositories:\r\n  - id: app\r\n    url: https://github.com/example/app\r\n    sha: aabbcc\r\n")
	manifestBOM := append([]byte("\xef\xbb\xbf"), manifestLF...)
	manifestComment := append([]byte("# Extra explanatory comment\n"), manifestLF...)

	variants := map[string][]byte{
		"LF":      manifestLF,
		"CRLF":    manifestCRLF,
		"BOM":     manifestBOM,
		"Comment": manifestComment,
	}

	hashes := make(map[string]string)
	sourceKeys := make(map[string]string)

	sources := []domain.SourceRepoIdentity{
		{
			RepoID:    "app",
			URLSHA256: gitcache.RepoHash("https://github.com/example/app"),
			CommitSHA: "aabbcc",
		},
	}

	for name, data := range variants {
		sum := sha256.Sum256(data)
		h := hex.EncodeToString(sum[:])
		sk := ComputeSourceKey(SourceKeyOptions{
			SchemaVersion:   2,
			ManifestLocator: "gira.yaml",
			ManifestBytes:   data,
			Sources:         sources,
		})
		hashes[name] = h
		sourceKeys[name] = sk
	}

	names := []string{"LF", "CRLF", "BOM", "Comment"}
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			n1, n2 := names[i], names[j]
			if hashes[n1] == hashes[n2] {
				t.Errorf("manifest hash collision between %s and %s: %s", n1, n2, hashes[n1])
			}
			if sourceKeys[n1] == sourceKeys[n2] {
				t.Errorf("source_key collision between %s and %s: %s", n1, n2, sourceKeys[n1])
			}
		}
	}
}

// AC3: main.date alone does not change tool_id; version/commit changes do; untrusted dev build includes executable hash
func TestToolID_Date_Version_Commit_DevBuild(t *testing.T) {
	cleanBuildInfoDate1 := &debug.BuildInfo{
		GoVersion: "go1.24.0",
		Path:      "github.com/EduardoShibukawa/gira-g/cmd/gira",
		Main: debug.Module{
			Path:    "github.com/EduardoShibukawa/gira-g",
			Version: "v1.2.0",
		},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "1234567890abcdef1234567890abcdef12345678"},
			{Key: "vcs.modified", Value: "false"},
			{Key: "-trimpath", Value: "true"},
			{Key: "-ldflags", Value: "-X main.date=2026-01-01T00:00:00Z"},
		},
	}

	cleanBuildInfoDate2 := &debug.BuildInfo{
		GoVersion: "go1.24.0",
		Path:      "github.com/EduardoShibukawa/gira-g/cmd/gira",
		Main: debug.Module{
			Path:    "github.com/EduardoShibukawa/gira-g",
			Version: "v1.2.0",
		},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "1234567890abcdef1234567890abcdef12345678"},
			{Key: "vcs.modified", Value: "false"},
			{Key: "-trimpath", Value: "true"},
			{Key: "-ldflags", Value: "-X main.date=2030-12-31T23:59:59Z"},
		},
	}

	// 1. Release builds differing ONLY in date must have identical tool_id
	toolID1, err := ComputeToolID(ToolIDOptions{
		Version:   "1.2.0",
		Commit:    "1234567890abcdef1234567890abcdef12345678",
		BuildInfo: cleanBuildInfoDate1,
	})
	if err != nil {
		t.Fatalf("ComputeToolID failed: %v", err)
	}

	toolID2, err := ComputeToolID(ToolIDOptions{
		Version:   "1.2.0",
		Commit:    "1234567890abcdef1234567890abcdef12345678",
		BuildInfo: cleanBuildInfoDate2,
	})
	if err != nil {
		t.Fatalf("ComputeToolID failed: %v", err)
	}

	if toolID1 != toolID2 {
		t.Fatalf("AC3 violation: tool_id diverged on date variation in release build: %s != %s", toolID1, toolID2)
	}

	// 2. Version change must alter tool_id
	toolIDVersionDiff, err := ComputeToolID(ToolIDOptions{
		Version:   "1.2.1",
		Commit:    "1234567890abcdef1234567890abcdef12345678",
		BuildInfo: cleanBuildInfoDate1,
	})
	if err != nil {
		t.Fatalf("ComputeToolID failed: %v", err)
	}
	if toolIDVersionDiff == toolID1 {
		t.Fatalf("AC3 violation: version change must alter tool_id")
	}

	// 3. Commit change must alter tool_id
	cleanBuildInfoCommitDiff := &debug.BuildInfo{
		GoVersion: "go1.24.0",
		Path:      "github.com/EduardoShibukawa/gira-g/cmd/gira",
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "fedcba0987654321fedcba0987654321fedcba09"},
			{Key: "vcs.modified", Value: "false"},
		},
	}
	toolIDCommitDiff, err := ComputeToolID(ToolIDOptions{
		Version:   "1.2.0",
		Commit:    "fedcba0987654321fedcba0987654321fedcba09",
		BuildInfo: cleanBuildInfoCommitDiff,
	})
	if err != nil {
		t.Fatalf("ComputeToolID failed: %v", err)
	}
	if toolIDCommitDiff == toolID1 {
		t.Fatalf("AC3 violation: commit change must alter tool_id")
	}

	// 4. Untrusted development builds include executable hash
	devBuildInfo := &debug.BuildInfo{
		GoVersion: "go1.24.0",
		Path:      "github.com/EduardoShibukawa/gira-g/cmd/gira",
		Settings: []debug.BuildSetting{
			{Key: "vcs.modified", Value: "true"},
		},
	}

	devToolID_ExeA, err := ComputeToolID(ToolIDOptions{
		Version:        "0.1.0-dev",
		Commit:         "none",
		BuildInfo:      devBuildInfo,
		ExecutableHash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	if err != nil {
		t.Fatalf("ComputeToolID failed: %v", err)
	}

	devToolID_ExeB, err := ComputeToolID(ToolIDOptions{
		Version:        "0.1.0-dev",
		Commit:         "none",
		BuildInfo:      devBuildInfo,
		ExecutableHash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	})
	if err != nil {
		t.Fatalf("ComputeToolID failed: %v", err)
	}

	if devToolID_ExeA == devToolID_ExeB {
		t.Fatalf("AC3 violation: dev build must include executable hash; differing hashes yielded identical tool_id")
	}
}

// AC7: Same version_id with different canonical bytes returns nondeterministic_build and preserves existing version intact
func TestVersionID_NonDeterministicBuild_PreservesExisting(t *testing.T) {
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "macro-map.json")

	canonicalBytesA := []byte(`{"schema_version":2,"target_system":"sysA","nodes":[],"relationships":[]}` + "\n")
	canonicalBytesB := []byte(`{"schema_version":2,"target_system":"sysA","nodes":[{"id":"node1"}],"relationships":[]}` + "\n")

	// 1. Initial write
	if err := os.WriteFile(indexPath, canonicalBytesA, 0644); err != nil {
		t.Fatalf("failed creating initial file: %v", err)
	}

	// 2. Idempotent check with identical bytes
	if err := VerifyDeterministicBuild(indexPath, canonicalBytesA); err != nil {
		t.Fatalf("expected nil for identical re-indexing, got: %v", err)
	}

	// 3. Nondeterministic build check with divergent bytes
	err := VerifyDeterministicBuild(indexPath, canonicalBytesB)
	if err == nil {
		t.Fatal("expected error for divergent canonical bytes, got nil")
	}
	if !errors.Is(err, ErrNonDeterministicBuild) {
		t.Fatalf("expected ErrNonDeterministicBuild, got: %v", err)
	}
	if !bytes.Contains([]byte(err.Error()), []byte("nondeterministic_build")) {
		t.Fatalf("expected error string to contain 'nondeterministic_build', got: %v", err)
	}

	// 4. Verify existing file is preserved byte-for-byte intact
	preserved, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed reading preserved file: %v", err)
	}
	if !bytes.Equal(preserved, canonicalBytesA) {
		t.Fatalf("existing file was corrupted or overwritten: got %s, want %s", string(preserved), string(canonicalBytesA))
	}
}

// Version ID derivation
func TestComputeVersionID(t *testing.T) {
	sourceKey1 := "1111111111111111111111111111111111111111111111111111111111111111"
	toolID1 := "2222222222222222222222222222222222222222222222222222222222222222"

	v1 := ComputeVersionID(sourceKey1, toolID1)
	if len(v1) != 64 {
		t.Fatalf("expected 64-char hex version_id, got %d chars: %s", len(v1), v1)
	}

	// Deterministic: same inputs give same version_id
	v2 := ComputeVersionID(sourceKey1, toolID1)
	if v1 != v2 {
		t.Fatalf("ComputeVersionID not deterministic: %s != %s", v1, v2)
	}

	// Different source_key gives different version_id
	sourceKey2 := "3333333333333333333333333333333333333333333333333333333333333333"
	v3 := ComputeVersionID(sourceKey2, toolID1)
	if v3 == v1 {
		t.Fatalf("different source_key must yield different version_id")
	}

	// Different tool_id gives different version_id
	toolID2 := "4444444444444444444444444444444444444444444444444444444444444444"
	v4 := ComputeVersionID(sourceKey1, toolID2)
	if v4 == v1 {
		t.Fatalf("different tool_id must yield different version_id")
	}
}

// BuildSourceIdentityFromManifest verification
func TestBuildSourceIdentityFromManifest(t *testing.T) {
	manifestBytes := []byte("dummy manifest content")
	repos := []manifest.RepositoryDeclaration{
		{
			ID:        "zebra-repo",
			URL:       "https://github.com/example/zebra",
			SHA:       "1111111111111111111111111111111111111111",
			PathScope: "pkg/z",
		},
		{
			ID:        "alpha-repo",
			URL:       "https://github.com/example/alpha",
			SHA:       "2222222222222222222222222222222222222222",
		},
	}
	extractors := []domain.ExtractorIdentity{
		{ID: "proto-extractor", Version: "1.0.0"},
		{ID: "go-extractor", Version: "1.0.0"},
	}

	si := BuildSourceIdentityFromManifest(manifestBytes, repos, extractors)
	if si == nil {
		t.Fatal("expected non-nil SourceIdentity")
	}
	if si.SchemaVersion != 2 {
		t.Fatalf("expected SchemaVersion 2, got %d", si.SchemaVersion)
	}

	sum := sha256.Sum256(manifestBytes)
	wantManifestSHA := hex.EncodeToString(sum[:])
	if si.ManifestSHA256 != wantManifestSHA {
		t.Fatalf("manifest_sha256 mismatch: got %s, want %s", si.ManifestSHA256, wantManifestSHA)
	}

	// Verify sources sorted by repo_id
	if len(si.Sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(si.Sources))
	}
	if si.Sources[0].RepoID != "alpha-repo" || si.Sources[1].RepoID != "zebra-repo" {
		t.Fatalf("sources not sorted by repo_id: %+v", si.Sources)
	}

	// Verify extractors sorted by ID
	if len(si.Extractors) != 2 {
		t.Fatalf("expected 2 extractors, got %d", len(si.Extractors))
	}
	if si.Extractors[0].ID != "go-extractor" || si.Extractors[1].ID != "proto-extractor" {
		t.Fatalf("extractors not sorted by ID: %+v", si.Extractors)
	}
}

// AC5: Store roundtrip with Schema 2 and SourceIdentity
func TestStore_Schema2_SourceIdentity_Roundtrip(t *testing.T) {
	tempDir := t.TempDir()
	st := NewStore(tempDir)

	si := &domain.SourceIdentity{
		SchemaVersion:  2,
		ManifestSHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		Extractors: domain.ExtractorList{
			{ID: "go-extractor", Version: "1.0.0"},
		},
		Sources: []domain.SourceRepoIdentity{
			{
				RepoID:    "app",
				CommitSHA: "1111111111111111111111111111111111111111",
				URLSHA256: "2222222222222222222222222222222222222222222222222222222222222222",
			},
		},
	}

	mm := domain.NewMacroMapV2("demo-system", si)
	_ = mm.AddNode(domain.Node{
		ID:   "service:app",
		Kind: domain.NodeService,
		Name: "app",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "app",
					CommitSHA: "1111111111111111111111111111111111111111",
					Path:      "main.go",
				},
				Origin: domain.OriginExtracted,
				Status: domain.StatusCurrent,
			},
		},
	})

	filePath := filepath.Join(tempDir, "demo-system", "macro-map.json")
	if err := st.SaveFile(filePath, mm); err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}

	loaded, err := st.LoadFile(filePath)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	if loaded.SchemaVersion != 2 {
		t.Fatalf("AC5 violation: loaded SchemaVersion = %#v (%T), want 2", loaded.SchemaVersion, loaded.SchemaVersion)
	}
	if loaded.SourceIdentity == nil {
		t.Fatal("AC5 violation: loaded SourceIdentity is nil")
	}
	if loaded.SourceIdentity.SchemaVersion != 2 {
		t.Fatalf("AC5 violation: SourceIdentity.SchemaVersion = %d, want 2", loaded.SourceIdentity.SchemaVersion)
	}
	if loaded.SourceIdentity.ManifestSHA256 != si.ManifestSHA256 {
		t.Fatalf("AC5 violation: ManifestSHA256 mismatch: got %s, want %s", loaded.SourceIdentity.ManifestSHA256, si.ManifestSHA256)
	}
	if len(loaded.SourceIdentity.Sources) != 1 || loaded.SourceIdentity.Sources[0].RepoID != "app" {
		t.Fatalf("AC5 violation: SourceIdentity.Sources mismatch: %+v", loaded.SourceIdentity.Sources)
	}
}

