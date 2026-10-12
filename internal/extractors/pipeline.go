package extractors

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"labs/gira-g/internal/domain"
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

// DefaultPipeline initializes a Pipeline with Wave 1 extractors pre-registered.
func DefaultPipeline() *Pipeline {
	return NewPipeline(
		NewDataFormatExtractor(),
		NewComposeExtractor(),
		NewHelmExtractor(),
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

	for _, file := range files {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

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
		}
	}

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
			// Create dummy service node to preserve graph integrity if source not found
			srcNode := domain.Node{
				ID:   r.SourceID,
				Kind: domain.NodeService,
				Name: r.SourceID,
			}
			if err := macroMap.AddNode(srcNode); err == nil {
				nodeMap[r.SourceID] = &srcNode
			}
		}
		if _, ok := nodeMap[r.TargetID]; !ok {
			// Create dummy service node to preserve graph integrity if target not found
			tgtNode := domain.Node{
				ID:   r.TargetID,
				Kind: domain.NodeService,
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
