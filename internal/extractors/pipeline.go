package extractors

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

// Pipeline manages extractors ordered by layer and processes target files.
type Pipeline struct {
	extractors []Extractor
	mu         sync.RWMutex
}

// NewPipeline creates a new Pipeline with the specified extractors.
// Extractors are sorted deterministically by Layer and then by Name.
func NewPipeline(extractors ...Extractor) *Pipeline {
	p := &Pipeline{}
	for _, ext := range extractors {
		p.Register(ext)
	}
	return p
}

// DefaultPipeline initializes a Pipeline with Wave 1, Wave 2, and Layer 3 extractors pre-registered.
func DefaultPipeline() *Pipeline {
	return NewPipeline(
		NewDataFormatExtractor(),
		NewDockerfileExtractor(),
		NewComposeExtractor(),
		NewEnvFileExtractor(),
		NewHelmExtractor(),
		NewGoModExtractor(),
		NewGitHubActionsExtractor(),
		NewOpenAPIExtractor(),
		NewMiseExtractor(),
		NewMarkdownExtractor(),
		NewProtobufExtractor(),
		NewGoASTExtractor(),
	)
}

// Register adds an extractor to the pipeline, keeping the list sorted.
func (p *Pipeline) Register(ext Extractor) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.extractors = append(p.extractors, ext)
	sort.Slice(p.extractors, func(i, j int) bool {
		if p.extractors[i].Layer() != p.extractors[j].Layer() {
			return p.extractors[i].Layer() < p.extractors[j].Layer()
		}
		return p.extractors[i].Name() < p.extractors[j].Name()
	})
}

// Extractors returns a snapshot copy of the registered extractors.
func (p *Pipeline) Extractors() []Extractor {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make([]Extractor, len(p.extractors))
	copy(out, p.extractors)
	return out
}

// ProcessFiles processes the given target files and builds a consolidated domain.MacroMap.
// For any file where no registered extractor returns CanExtract == true, an UnsupportedBoundary
// record is deterministically emitted.
func (p *Pipeline) ProcessFiles(ctx context.Context, targetSystem string, files []TargetFile) (*domain.MacroMap, error) {
	macroMap := domain.NewMacroMap(targetSystem)
	extractors := p.Extractors()

	// Internal deduplication maps
	nodeMap := make(map[string]*domain.Node)
	relMap := make(map[string]*domain.Relationship)
	unsupportedMap := make(map[string]domain.UnsupportedBoundary)
	unsupportedContent := make(map[string][]byte)
	var hints []Hint

	for _, file := range files {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Logical paths are slash-form on every OS: filepath.WalkDir yields
		// backslash paths on Windows, which would otherwise leak into node
		// IDs, locators and golden comparisons. No-op on Linux.
		file.Path = filepath.ToSlash(file.Path)

		var matched []Extractor
		for _, ext := range extractors {
			if ext.CanExtract(file.Path) {
				matched = append(matched, ext)
			}
		}

		if len(matched) == 0 {
			// No extractor registered for file: emit explicit UnsupportedBoundary
			ub := domain.UnsupportedBoundary{
				Path:   file.Path,
				Reason: "no extractor registered for file extension",
				Locator: domain.Locator{
					Repo:      file.Repo,
					CommitSHA: file.CommitSHA,
					Path:      file.Path,
				},
			}
			unsupportedMap[file.Path] = ub
			unsupportedContent[file.Path] = file.Content
			continue
		}

		// Run all matching extractors (ordered by layer)
		for _, ext := range matched {
			nodes, rels, unsupporteds, err := ext.Extract(ctx, file)
			if err != nil {
				return nil, fmt.Errorf("extractor %s failed on file %s: %w", ext.Name(), file.Path, err)
			}

			// Aggregate nodes (merging evidence and metadata on duplicate node ID)
			for _, n := range nodes {
				existing, found := nodeMap[n.ID]
				if !found {
					nCopy := n
					nodeMap[n.ID] = &nCopy
				} else {
					// Merge evidence
					existing.Evidence = append(existing.Evidence, n.Evidence...)
					// Merge metadata
					if existing.Metadata == nil && len(n.Metadata) > 0 {
						existing.Metadata = make(map[string]any)
					}
					for k, v := range n.Metadata {
						if _, has := existing.Metadata[k]; !has {
							existing.Metadata[k] = v
						}
					}
					// Update description if previous was empty
					if existing.Description == "" && n.Description != "" {
						existing.Description = n.Description
					}
				}
			}

			// Aggregate relationships
			for _, r := range rels {
				existing, found := relMap[r.ID]
				if !found {
					rCopy := r
					relMap[r.ID] = &rCopy
				} else {
					existing.Evidence = append(existing.Evidence, r.Evidence...)
				}
			}

			// Aggregate unsupported boundaries
			for _, u := range unsupporteds {
				unsupportedMap[u.Path] = u
			}

			if he, ok := ext.(HintEmitter); ok {
				hs, err := he.Hints(ctx, file)
				if err != nil {
					return nil, fmt.Errorf("extractor %s failed on hints for %s: %w", ext.Name(), file.Path, err)
				}
				hints = append(hints, hs...)
			}
		}
	}

	// Cross-file resolution step (ADR 0013).
	resolve(&resolution{nodes: nodeMap, rels: relMap, hints: hints})
	annotateUnknownJoins(unsupportedMap, unsupportedContent, nodeMap)

	// Transfer nodes to MacroMap
	for _, n := range nodeMap {
		if err := macroMap.AddNode(*n); err != nil {
			return nil, fmt.Errorf("failed adding node %s: %w", n.ID, err)
		}
	}

	// Transfer relationships to MacroMap, ensuring endpoints exist
	for _, r := range relMap {
		// Ensure endpoints exist in nodeMap
		if _, ok := nodeMap[r.SourceID]; !ok {
			// Create dummy node to preserve graph integrity if source not found
			srcKind := deriveKindFromID(r.SourceID)
			srcNode := domain.Node{
				ID:   r.SourceID,
				Kind: srcKind,
				Name: r.SourceID,
			}
			if err := macroMap.AddNode(srcNode); err == nil {
				nodeMap[r.SourceID] = &srcNode
			}
		}
		if _, ok := nodeMap[r.TargetID]; !ok {
			// Create dummy node to preserve graph integrity if target not found
			tgtKind := deriveKindFromID(r.TargetID)
			tgtNode := domain.Node{
				ID:   r.TargetID,
				Kind: tgtKind,
				Name: r.TargetID,
			}
			if err := macroMap.AddNode(tgtNode); err == nil {
				nodeMap[r.TargetID] = &tgtNode
			}
		}

		if err := macroMap.AddRelationship(*r); err != nil {
			return nil, fmt.Errorf("failed adding relationship %s: %w", r.ID, err)
		}
	}

	// Transfer unsupported boundaries
	for _, u := range unsupportedMap {
		macroMap.Unsupported = append(macroMap.Unsupported, u)
	}

	return macroMap, nil
}

// OriginDeclaredOverride identifies provenance from declarative YAML overlays.
const OriginDeclaredOverride domain.DerivationOrigin = "declared:override"

// ApplyOverrides applies declarative manifest overrides to the given MacroMap.
// Relationships added or updated receive origin "declared:override" and a Locator pointing to manifestPath.
// Metadata for existing nodes is updated without opaque mutation (with explicit provenance evidence).
func ApplyOverrides(mm *domain.MacroMap, overrides *manifest.OverridesDeclaration, manifestPath string) error {
	if mm == nil {
		return fmt.Errorf("macro map cannot be nil")
	}
	if overrides == nil {
		return nil
	}
	if strings.TrimSpace(manifestPath) == "" {
		manifestPath = "gira.yaml"
	}

	now := time.Now()

	// 1. Apply metadata overrides to existing nodes
	for nodeID, meta := range overrides.Metadata {
		var targetNode *domain.Node
		for i := range mm.Nodes {
			if mm.Nodes[i].ID == nodeID {
				targetNode = &mm.Nodes[i]
				break
			}
		}

		if targetNode != nil {
			if targetNode.Metadata == nil {
				targetNode.Metadata = make(map[string]any)
			}
			for k, v := range meta {
				targetNode.Metadata[k] = v
			}
			targetNode.Evidence = append(targetNode.Evidence, domain.EvidenceRecord{
				Locator: domain.Locator{
					Path: manifestPath,
				},
				Origin:       OriginDeclaredOverride,
				Status:       domain.StatusCurrent,
				DiscoveredAt: now,
			})
		}
	}

	// 2. Apply relationship overrides
	for _, relOverride := range overrides.Relationships {
		overrideEvidence := domain.EvidenceRecord{
			Locator: domain.Locator{
				Path: manifestPath,
			},
			Origin:       OriginDeclaredOverride,
			Status:       domain.StatusCurrent,
			DiscoveredAt: now,
		}

		// Ensure source and target nodes exist in MacroMap so AddRelationship succeeds
		ensureNodeExists(mm, relOverride.SourceID, manifestPath, now)
		ensureNodeExists(mm, relOverride.TargetID, manifestPath, now)

		// Check if a relationship between source and target already exists
		var matchedRel *domain.Relationship
		for i := range mm.Relationships {
			if mm.Relationships[i].SourceID == relOverride.SourceID && mm.Relationships[i].TargetID == relOverride.TargetID {
				matchedRel = &mm.Relationships[i]
				break
			}
		}

		if matchedRel != nil {
			matchedRel.Verb = domain.RelationVerb(relOverride.Verb)
			matchedRel.Evidence = append(matchedRel.Evidence, overrideEvidence)
		} else {
			relID := fmt.Sprintf("rel:%s-%s-%s", relOverride.SourceID, relOverride.Verb, relOverride.TargetID)
			newRel := domain.Relationship{
				ID:       relID,
				SourceID: relOverride.SourceID,
				TargetID: relOverride.TargetID,
				Verb:     domain.RelationVerb(relOverride.Verb),
				Evidence: []domain.EvidenceRecord{overrideEvidence},
			}
			if err := mm.AddRelationship(newRel); err != nil {
				return fmt.Errorf("failed to add override relationship %s: %w", relID, err)
			}
		}
	}

	return nil
}

func deriveKindFromID(nodeID string) domain.NodeType {
	switch {
	case strings.HasPrefix(nodeID, "contract:"):
		return domain.NodeContract
	case strings.HasPrefix(nodeID, "infra:") || strings.HasPrefix(nodeID, "infrastructure:"):
		return domain.NodeInfrastructure
	case strings.HasPrefix(nodeID, "lib:") || strings.HasPrefix(nodeID, "library:"):
		return domain.NodeLibrary
	case strings.HasPrefix(nodeID, "policy:"):
		return domain.NodePolicy
	case strings.HasPrefix(nodeID, "skill:"):
		return domain.NodeSkill
	case strings.HasPrefix(nodeID, "task:"):
		return domain.NodeTask
	default:
		return domain.NodeService
	}
}

func ensureNodeExists(mm *domain.MacroMap, nodeID, manifestPath string, now time.Time) {
	for _, n := range mm.Nodes {
		if n.ID == nodeID {
			return
		}
	}

	kind := deriveKindFromID(nodeID)

	name := nodeID
	if idx := strings.Index(nodeID, ":"); idx != -1 {
		name = nodeID[idx+1:]
	}

	node := domain.Node{
		ID:   nodeID,
		Kind: kind,
		Name: name,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Path: manifestPath,
				},
				Origin:       OriginDeclaredOverride,
				Status:       domain.StatusCurrent,
				DiscoveredAt: now,
			},
		},
	}
	_ = mm.AddNode(node)
}

// ApplyOverrides applies declarative manifest overrides to the given MacroMap.
func (p *Pipeline) ApplyOverrides(mm *domain.MacroMap, overrides *manifest.OverridesDeclaration, manifestPath string) error {
	return ApplyOverrides(mm, overrides, manifestPath)
}
