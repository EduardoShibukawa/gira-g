package store

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

var (
	// ErrCorruptIndex indicates an index file is malformed JSON, truncated, empty, or has corrupt graph data.
	ErrCorruptIndex = errors.New("index_corrupt: corrupt or invalid macro-map index")

	// ErrUnsupportedSchema indicates an unknown schema version, invalid type, missing mandatory fields,
	// or an attempt to implicitly load legacy schema 1.
	ErrUnsupportedSchema = errors.New("index_schema_unsupported: unsupported index schema")

	// ErrIndexMissing indicates an index file, active head, or version is not found.
	ErrIndexMissing = errors.New("index_missing: index not found")

	// ErrToolMismatch indicates the index was built with an incompatible tool/binary version.
	ErrToolMismatch = errors.New("index_tool_mismatch: index built with incompatible tool")

	// ErrIndexAmbiguous indicates that auto-discovery without a manifest found multiple candidate systems.
	ErrIndexAmbiguous = errors.New("index_ambiguous: multiple candidate systems found in cache")

	// ErrSchemaDowngrade indicates an unauthorized attempt to downgrade a schema 2 index to legacy schema 1.
	ErrSchemaDowngrade = errors.New("refusing to downgrade index schema")
)

// LoadOptions controls schema parsing and backward compatibility rules under Policy R2.
type LoadOptions struct {
	// AllowLegacy authorizes loading Schema 1 ("1.0.0") in read-only mode (exclusive to --index flag).
	AllowLegacy bool
	// WarnWriter receives a visible deprecation warning when a legacy Schema 1 index is loaded.
	WarnWriter io.Writer
}

// IsSchema2 reports whether the given MacroMap conforms to Schema 2 (integer 2).
func IsSchema2(m *domain.MacroMap) bool {
	if m == nil {
		return false
	}
	switch v := m.SchemaVersion.(type) {
	case int:
		return v == 2
	case float64:
		return v == 2.0
	case json.Number:
		n, err := v.Int64()
		return err == nil && n == 2
	default:
		return false
	}
}

// IsSchema1 reports whether the given MacroMap conforms to legacy Schema 1 ("1.0.0").
func IsSchema1(m *domain.MacroMap) bool {
	if m == nil {
		return false
	}
	switch v := m.SchemaVersion.(type) {
	case string:
		return v == "1.0.0"
	default:
		return false
	}
}

// ValidateSchemaData decodes and strictly validates MacroMap JSON bytes according to Policy R2.
func ValidateSchemaData(data []byte, opts LoadOptions) (*domain.MacroMap, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("%w: file is empty", ErrCorruptIndex)
	}

	// 1. JSON syntax check
	if !json.Valid(data) {
		return nil, fmt.Errorf("%w: malformed JSON syntax", ErrCorruptIndex)
	}

	// 2. Decode raw structural header
	var raw struct {
		SchemaVersion  any                    `json:"schema_version"`
		TargetSystem   string                 `json:"target_system"`
		SourceIdentity *domain.SourceIdentity `json:"source_identity"`
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("%w: failed decoding JSON: %v", ErrCorruptIndex, err)
	}

	// 3. Structural validation: schema_version presence and type
	if raw.SchemaVersion == nil {
		return nil, fmt.Errorf("%w: missing required field schema_version", ErrUnsupportedSchema)
	}

	var schemaInt int
	var schemaStr string
	isInteger := false
	isString := false

	switch v := raw.SchemaVersion.(type) {
	case json.Number:
		intVal, err := v.Int64()
		if err != nil {
			return nil, fmt.Errorf("%w: schema_version must be integer, got float/number %s", ErrUnsupportedSchema, v.String())
		}
		schemaInt = int(intVal)
		isInteger = true
	case int:
		schemaInt = v
		isInteger = true
	case float64:
		if v != float64(int(v)) {
			return nil, fmt.Errorf("%w: schema_version must be integer, got float %v", ErrUnsupportedSchema, v)
		}
		schemaInt = int(v)
		isInteger = true
	case string:
		schemaStr = v
		isString = true
	default:
		return nil, fmt.Errorf("%w: schema_version has invalid type %T", ErrUnsupportedSchema, raw.SchemaVersion)
	}

	// 4. Structural validation: target_system
	if strings.TrimSpace(raw.TargetSystem) == "" {
		return nil, fmt.Errorf("%w: missing required field target_system", ErrUnsupportedSchema)
	}

	// 5. Schema allowlist enforcement
	if isInteger {
		if schemaInt != 2 {
			return nil, fmt.Errorf("%w: unsupported integer schema_version %d (supported: 2)", ErrUnsupportedSchema, schemaInt)
		}
		// Strict Schema 2: source_identity is mandatory
		if raw.SourceIdentity == nil {
			return nil, fmt.Errorf("%w: schema 2 requires source_identity metadata", ErrUnsupportedSchema)
		}
		if raw.SourceIdentity.SchemaVersion != 2 {
			return nil, fmt.Errorf("%w: source_identity.schema_version must be 2, got %d", ErrUnsupportedSchema, raw.SourceIdentity.SchemaVersion)
		}
		if strings.TrimSpace(raw.SourceIdentity.ManifestSHA256) == "" {
			return nil, fmt.Errorf("%w: source_identity.manifest_sha256 is required", ErrUnsupportedSchema)
		}
		for _, src := range raw.SourceIdentity.Sources {
			if strings.TrimSpace(src.RepoID) == "" {
				return nil, fmt.Errorf("%w: source_identity.sources entry missing repo_id", ErrUnsupportedSchema)
			}
			if strings.TrimSpace(src.URLSHA256) == "" {
				return nil, fmt.Errorf("%w: source_identity.sources entry %q missing url_sha256", ErrUnsupportedSchema, src.RepoID)
			}
		}
	} else if isString {
		if schemaStr != "1.0.0" {
			return nil, fmt.Errorf("%w: unsupported string schema_version %q (supported legacy: \"1.0.0\")", ErrUnsupportedSchema, schemaStr)
		}
		if !opts.AllowLegacy {
			return nil, fmt.Errorf("%w: legacy schema \"1.0.0\" is only permitted via explicit --index flag", ErrUnsupportedSchema)
		}
		if opts.WarnWriter != nil {
			fmt.Fprintf(opts.WarnWriter, "warning: legacy index format (schema 1.0.0) is deprecated and read-only; re-index with 'gira index' to upgrade to schema 2\n")
		}
	} else {
		return nil, fmt.Errorf("%w: unsupported schema_version type", ErrUnsupportedSchema)
	}

	// 6. Decode full MacroMap
	var mm domain.MacroMap
	if err := json.Unmarshal(data, &mm); err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal macro map: %v", ErrCorruptIndex, err)
	}

	// 7. Graph and consistency validation
	if err := mm.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptIndex, err)
	}

	return &mm, nil
}

// LoadFileWithOptions reads, decodes, and validates a MacroMap from filePath under the given options.
func (s *Store) LoadFileWithOptions(filePath string, opts LoadOptions) (*domain.MacroMap, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: file not found: %s", ErrIndexMissing, filePath)
		}
		return nil, fmt.Errorf("failed to read macro map file %s: %w", filePath, err)
	}
	mm, err := ValidateSchemaData(data, opts)
	if err != nil {
		return nil, fmt.Errorf("%w at %s", err, filePath)
	}
	return mm, nil
}

// LoadFileStrict reads, decodes, and validates a MacroMap from filePath, strictly requiring Schema 2.
func (s *Store) LoadFileStrict(filePath string) (*domain.MacroMap, error) {
	return s.LoadFileWithOptions(filePath, LoadOptions{AllowLegacy: false})
}

// ResolveOptions defines parameters for index resolution across stores and lineages.
type ResolveOptions struct {
	// ExplicitPath is the path passed to --index <path>. If set, resolution operates on this file directly.
	ExplicitPath string
	// ManifestPath is the path to gira.yaml for the current project context.
	ManifestPath string
	// ProjectRoot is the directory containing the project. Defaults to Dir(ManifestPath).
	ProjectRoot string
	// StoreRoots overrides standard store locations (useful in testing).
	StoreRoots []string
	// ExpectedToolID optionally checks that the loaded index was produced by the current binary.
	ExpectedToolID string
	// AllowLegacy permits Schema 1 in explicit path resolution.
	AllowLegacy bool
	// WarnWriter receives deprecation warnings for legacy schemas.
	WarnWriter io.Writer
}

// ResolveResult represents the resolved MacroMap and associated provenance.
type ResolveResult struct {
	MacroMap    *domain.MacroMap
	VersionID   string
	LineageID   string
	StoreRoot   string
	IsLegacy    bool
	SourcePath  string
	VersionMeta *VersionMeta
}

// Resolve executes the strict index resolution algorithm (Policy R2).
func Resolve(ctx context.Context, opts ResolveOptions) (*ResolveResult, error) {
	st := NewStore("")

	// Branch 1: Explicit --index flag
	if strings.TrimSpace(opts.ExplicitPath) != "" {
		fi, err := os.Stat(opts.ExplicitPath)
		if err != nil || fi.IsDir() {
			return nil, fmt.Errorf("%w: index file not found: %s", ErrIndexMissing, opts.ExplicitPath)
		}
		mm, err := st.LoadFileWithOptions(opts.ExplicitPath, LoadOptions{
			AllowLegacy: true,
			WarnWriter:  opts.WarnWriter,
		})
		if err != nil {
			return nil, err
		}
		return &ResolveResult{
			MacroMap:   mm,
			IsLegacy:   IsSchema1(mm),
			SourcePath: opts.ExplicitPath,
		}, nil
	}

	// Determine candidate store roots (Hierarchy D2)
	roots := opts.StoreRoots
	if len(roots) == 0 {
		if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
			roots = append(roots, filepath.Join(xdg, "gira", "indexes"))
		}
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			roots = append(roots, filepath.Join(home, ".cache", "gira", "indexes"))
		}
		if opts.ProjectRoot != "" {
			roots = append(roots, filepath.Join(opts.ProjectRoot, ".gira", "indexes"))
		}
		cwdProj := filepath.Join(".gira", "indexes")
		if len(roots) == 0 || roots[len(roots)-1] != cwdProj {
			roots = append(roots, cwdProj)
		}
	}

	// Branch 2: gira.yaml is present (or project context)
	if strings.TrimSpace(opts.ManifestPath) != "" {
		m, err := manifest.ParseFile(opts.ManifestPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse project manifest %s: %w", opts.ManifestPath, err)
		}
		sysName := m.TargetSystem.Name
		if opts.ProjectRoot == "" {
			opts.ProjectRoot = filepath.Dir(opts.ManifestPath)
		}

		lineages := []string{sysName}
		if len(m.Repositories) > 0 {
			sources := make([]domain.SourceRepoIdentity, len(m.Repositories))
			for i, r := range m.Repositories {
				sources[i] = domain.SourceRepoIdentity{
					RepoID:    r.ID,
					URLSHA256: gitcache.RepoHash(r.URL),
				}
			}
			linKey := ComputeLineageKey(sysName, sources)
			if linKey != sysName {
				lineages = append(lineages, linKey)
			}
		}

		for _, root := range roots {
			storeInst := NewStore(root)
			for _, lin := range lineages {
				headPath := filepath.Join(root, lin, ".head")
				if fi, err := os.Stat(headPath); err == nil && !fi.IsDir() {
					activeVer, readErr := storeInst.ReadHead(lin)
					if readErr != nil {
						return nil, fmt.Errorf("%w: failed reading .head at %s: %v", ErrCorruptIndex, headPath, readErr)
					}
					versionDir := filepath.Join(root, lin, "versions", activeVer)
					metaPath := filepath.Join(versionDir, "meta.json")
					mapPath := filepath.Join(versionDir, "macro-map.json")

					if _, err := os.Stat(metaPath); err != nil {
						return nil, fmt.Errorf("%w: version %s missing meta.json: %v", ErrCorruptIndex, activeVer, err)
					}
					if _, err := os.Stat(mapPath); err != nil {
						return nil, fmt.Errorf("%w: version %s missing macro-map.json: %v", ErrCorruptIndex, activeVer, err)
					}

					metaBytes, err := os.ReadFile(metaPath)
					if err != nil {
						return nil, fmt.Errorf("%w: cannot read meta.json: %v", ErrCorruptIndex, err)
					}
					var meta VersionMeta
					if err := json.Unmarshal(metaBytes, &meta); err != nil {
						return nil, fmt.Errorf("%w: corrupted meta.json: %v", ErrCorruptIndex, err)
					}
					if opts.ExpectedToolID != "" && meta.ToolID != "" && meta.ToolID != opts.ExpectedToolID {
						return nil, fmt.Errorf("%w: index tool_id %q does not match current tool_id %q (re-index with 'gira index')", ErrToolMismatch, meta.ToolID, opts.ExpectedToolID)
					}

					mm, err := storeInst.LoadFileWithOptions(mapPath, LoadOptions{AllowLegacy: false})
					if err != nil {
						return nil, err
					}

					return &ResolveResult{
						MacroMap:    mm,
						VersionID:   activeVer,
						LineageID:   lin,
						StoreRoot:   root,
						IsLegacy:    false,
						SourcePath:  mapPath,
						VersionMeta: &meta,
					}, nil
				}
			}
		}

		// When gira.yaml is present, legacy unversioned files are NEVER loaded.
		return nil, fmt.Errorf("%w: no index found for target system %q declared in %s (run 'gira index' first)", ErrIndexMissing, sysName, opts.ManifestPath)
	}

	// Branch 3: No gira.yaml - Discovery mode
	type candidate struct {
		name      string
		storeRoot string
		isVersion bool
	}
	var candidates []candidate
	seen := make(map[string]bool)

	for _, root := range roots {
		storeInst := NewStore(root)
		systems, err := storeInst.List()
		if err != nil {
			continue
		}
		for _, sys := range systems {
			if seen[sys] {
				continue
			}
			headPath := filepath.Join(root, sys, ".head")
			hasHead := false
			if fi, err := os.Stat(headPath); err == nil && !fi.IsDir() {
				hasHead = true
			}
			seen[sys] = true
			candidates = append(candidates, candidate{
				name:      sys,
				storeRoot: root,
				isVersion: hasHead,
			})
		}
	}

	if len(candidates) > 1 {
		names := make([]string, len(candidates))
		for i, c := range candidates {
			names[i] = c.name
		}
		return nil, fmt.Errorf("%w: ambiguous index: multiple systems found in cache (%s); run from a directory with gira.yaml or pass --index <path>", ErrIndexAmbiguous, strings.Join(names, ", "))
	}

	if len(candidates) == 1 {
		c := candidates[0]
		if c.isVersion {
			storeInst := NewStore(c.storeRoot)
			activeVer, readErr := storeInst.ReadHead(c.name)
			if readErr != nil {
				return nil, fmt.Errorf("%w: failed reading .head: %v", ErrCorruptIndex, readErr)
			}
			versionDir := filepath.Join(c.storeRoot, c.name, "versions", activeVer)
			metaPath := filepath.Join(versionDir, "meta.json")
			mapPath := filepath.Join(versionDir, "macro-map.json")

			if metaBytes, err := os.ReadFile(metaPath); err == nil {
				var meta VersionMeta
				if err := json.Unmarshal(metaBytes, &meta); err == nil {
					if opts.ExpectedToolID != "" && meta.ToolID != "" && meta.ToolID != opts.ExpectedToolID {
						return nil, fmt.Errorf("%w: index tool_id %q does not match current tool_id %q", ErrToolMismatch, meta.ToolID, opts.ExpectedToolID)
					}
				}
			}

			mm, err := storeInst.LoadFileWithOptions(mapPath, LoadOptions{AllowLegacy: false})
			if err != nil {
				return nil, err
			}
			return &ResolveResult{
				MacroMap:   mm,
				VersionID:  activeVer,
				LineageID:  c.name,
				StoreRoot:  c.storeRoot,
				SourcePath: mapPath,
			}, nil
		}
		// If candidate is legacy without .head, do not load without explicit --index
		return nil, fmt.Errorf("%w: no index found (run 'gira index' first or pass --index <path>)", ErrIndexMissing)
	}

	return nil, fmt.Errorf("%w: no index found (run 'gira index' first or pass --index <path>)", ErrIndexMissing)
}
