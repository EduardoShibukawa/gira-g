package slices

import (
	"fmt"
	"sort"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/gitdiff"
)

// DiffTargets explains how a diff was mapped to the briefing's targets.
type DiffTargets struct {
	Range        string   `json:"range"`
	ChangedPaths []string `json:"changed_paths"`
	Targets      []string `json:"targets"`
	// UnmatchedPaths are changed files no locator in the index cites at all:
	// gira knows nothing about them.
	UnmatchedPaths []string `json:"unmatched_paths,omitempty"`
	// FileLevelPaths are changed files the index cites, but where no changed range
	// touches a cited range. Their targets are the Components (Services and
	// Infrastructure) whose evidence cites the file: a file-level match, coarser
	// than a range match, and reported as such.
	FileLevelPaths []string `json:"file_level_paths,omitempty"`
}

// GetPRImpact maps changed line ranges to nodes and returns one briefing over all
// of them. A node is a target when one of its evidence locators cites a changed
// range; a relationship whose evidence cites one makes its source a target (the
// side that declares the dependency changed). repoID, when set, restricts locators
// to that repository. A changed file the index cites, where no changed range
// touches a cited range, falls back to the Components citing that file and is
// listed in file_level_paths. A changed file no locator cites is listed as
// unmatched: gira knows nothing about it, which is reported, never guessed.
func (e *Engine) GetPRImpact(revRange, repoID string, changes []gitdiff.FileChange, depth int) (*ImpactBriefing, error) {
	byPath := make(map[string]gitdiff.FileChange, len(changes))
	paths := make([]string, 0, len(changes))
	for _, c := range changes {
		byPath[c.Path] = c
		paths = append(paths, c.Path)
	}
	sort.Strings(paths)

	targets := map[string]bool{}
	matched := map[string]bool{}
	hit := func(evs []domain.EvidenceRecord) bool {
		found := false
		for _, ev := range evs {
			if repoID != "" && ev.Locator.Repo != repoID {
				continue
			}
			c, ok := byPath[ev.Locator.Path]
			if ok && c.Touches(ev.Locator.StartLine, ev.Locator.EndLine) {
				matched[c.Path] = true
				found = true
			}
		}
		return found
	}
	for _, n := range e.macroMap.Nodes {
		if hit(n.Evidence) {
			targets[n.ID] = true
		}
	}
	for _, r := range e.macroMap.Relationships {
		if hit(r.Evidence) {
			if _, ok := e.nodesByID[r.SourceID]; ok {
				targets[r.SourceID] = true
			}
		}
	}

	// File-level fallback for cited files that no range match reached.
	var fileLevel []string
	for _, p := range paths {
		if matched[p] {
			continue
		}
		owners := map[string]bool{}
		for _, n := range e.macroMap.Nodes {
			if n.Kind != domain.NodeService && n.Kind != domain.NodeInfrastructure {
				continue
			}
			for _, ev := range n.Evidence {
				if ev.Locator.Path == p && (repoID == "" || ev.Locator.Repo == repoID) {
					owners[n.ID] = true
				}
			}
		}
		cited := len(owners) > 0
		if !cited {
			cited = e.cites(p, repoID)
		}
		if cited {
			matched[p] = true
			fileLevel = append(fileLevel, p)
			for id := range owners {
				targets[id] = true
			}
		}
	}

	ids := make([]string, 0, len(targets))
	for id := range targets {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var unmatched []string
	for _, p := range paths {
		if !matched[p] {
			unmatched = append(unmatched, p)
		}
	}

	if depth <= 0 {
		depth = 1
	}
	merged := &ImpactBriefing{
		TargetID:           "diff:" + revRange,
		Depth:              depth,
		UpstreamServices:   []domain.Node{},
		DownstreamServices: []domain.Node{},
		AffectedContracts:  []domain.Node{},
		MediationEdges:     []domain.Relationship{},
		Diff:               &DiffTargets{Range: revRange, ChangedPaths: paths, Targets: ids, UnmatchedPaths: unmatched, FileLevelPaths: fileLevel},
	}
	if merged.Diff.ChangedPaths == nil {
		merged.Diff.ChangedPaths = []string{}
	}
	if merged.Diff.Targets == nil {
		merged.Diff.Targets = []string{}
	}

	up, down, contracts := map[string]domain.Node{}, map[string]domain.Node{}, map[string]domain.Node{}
	libs := map[string]domain.Node{}
	edges := map[string]domain.Relationship{}
	ubs := map[string]domain.UnsupportedBoundary{}
	for _, id := range ids {
		b, err := e.GetImpactBriefing(id, depth)
		if err != nil {
			return nil, fmt.Errorf("briefing for diff target %s: %w", id, err)
		}
		for _, n := range b.UpstreamServices {
			up[n.ID] = n
		}
		for _, n := range b.UpstreamLibraries {
			libs[n.ID] = n
		}
		for _, n := range b.DownstreamServices {
			down[n.ID] = n
		}
		for _, n := range b.AffectedContracts {
			contracts[n.ID] = n
		}
		for _, r := range b.MediationEdges {
			edges[r.ID] = r
		}
		for _, u := range b.UnsupportedBoundaries {
			ubs[u.Path] = u
		}
	}
	merged.UpstreamServices = sortedNodes(up)
	if len(libs) > 0 {
		merged.UpstreamLibraries = sortedNodes(libs)
	}
	merged.DownstreamServices = sortedNodes(down)
	merged.AffectedContracts = sortedNodes(contracts)
	for _, r := range edges {
		merged.MediationEdges = append(merged.MediationEdges, r)
	}
	sort.Slice(merged.MediationEdges, func(i, j int) bool { return merged.MediationEdges[i].ID < merged.MediationEdges[j].ID })
	for _, u := range ubs {
		merged.UnsupportedBoundaries = append(merged.UnsupportedBoundaries, u)
	}
	sort.Slice(merged.UnsupportedBoundaries, func(i, j int) bool {
		return merged.UnsupportedBoundaries[i].Path < merged.UnsupportedBoundaries[j].Path
	})
	return merged, nil
}

func sortedNodes(m map[string]domain.Node) []domain.Node {
	out := make([]domain.Node, 0, len(m))
	for _, n := range m {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// cites reports whether any evidence in the index cites path.
func (e *Engine) cites(path, repoID string) bool {
	match := func(evs []domain.EvidenceRecord) bool {
		for _, ev := range evs {
			if ev.Locator.Path == path && (repoID == "" || ev.Locator.Repo == repoID) {
				return true
			}
		}
		return false
	}
	for _, n := range e.macroMap.Nodes {
		if match(n.Evidence) {
			return true
		}
	}
	for _, r := range e.macroMap.Relationships {
		if match(r.Evidence) {
			return true
		}
	}
	return false
}
