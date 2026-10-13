package slices

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// CanonicalJSON serializes a ContextSlice to byte-deterministic JSON with sorted keys.
func (cs *ContextSlice) CanonicalJSON() ([]byte, error) {
	sortedNeighbors := make([]domain.Node, len(cs.Neighbors))
	copy(sortedNeighbors, cs.Neighbors)
	sort.Slice(sortedNeighbors, func(i, j int) bool { return sortedNeighbors[i].ID < sortedNeighbors[j].ID })

	sortedRels := make([]domain.Relationship, len(cs.Relationships))
	copy(sortedRels, cs.Relationships)
	sort.Slice(sortedRels, func(i, j int) bool { return sortedRels[i].ID < sortedRels[j].ID })

	sortedUnsupported := make([]domain.UnsupportedBoundary, len(cs.UnsupportedBoundaries))
	copy(sortedUnsupported, cs.UnsupportedBoundaries)
	sort.Slice(sortedUnsupported, func(i, j int) bool { return sortedUnsupported[i].Path < sortedUnsupported[j].Path })

	normalized := ContextSlice{
		FocalNode:             cs.FocalNode,
		Tier:                  cs.Tier,
		Neighbors:             sortedNeighbors,
		Relationships:         sortedRels,
		UnsupportedBoundaries: sortedUnsupported,
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(normalized); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CanonicalJSON serializes an ImpactBriefing to byte-deterministic JSON with sorted keys.
func (ib *ImpactBriefing) CanonicalJSON() ([]byte, error) {
	sortedUpstream := make([]domain.Node, len(ib.UpstreamServices))
	copy(sortedUpstream, ib.UpstreamServices)
	sort.Slice(sortedUpstream, func(i, j int) bool { return sortedUpstream[i].ID < sortedUpstream[j].ID })

	sortedDownstream := make([]domain.Node, len(ib.DownstreamServices))
	copy(sortedDownstream, ib.DownstreamServices)
	sort.Slice(sortedDownstream, func(i, j int) bool { return sortedDownstream[i].ID < sortedDownstream[j].ID })

	sortedContracts := make([]domain.Node, len(ib.AffectedContracts))
	copy(sortedContracts, ib.AffectedContracts)
	sort.Slice(sortedContracts, func(i, j int) bool { return sortedContracts[i].ID < sortedContracts[j].ID })

	sortedMediation := make([]domain.Relationship, len(ib.MediationEdges))
	copy(sortedMediation, ib.MediationEdges)
	sort.Slice(sortedMediation, func(i, j int) bool { return sortedMediation[i].ID < sortedMediation[j].ID })

	sortedUnsupported := make([]domain.UnsupportedBoundary, len(ib.UnsupportedBoundaries))
	copy(sortedUnsupported, ib.UnsupportedBoundaries)
	sort.Slice(sortedUnsupported, func(i, j int) bool { return sortedUnsupported[i].Path < sortedUnsupported[j].Path })

	normalized := ImpactBriefing{
		TargetID:              ib.TargetID,
		TargetKind:            ib.TargetKind,
		Depth:                 ib.Depth,
		UpstreamServices:      sortedUpstream,
		DownstreamServices:    sortedDownstream,
		AffectedContracts:     sortedContracts,
		MediationEdges:        sortedMediation,
		UnsupportedBoundaries: sortedUnsupported,
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(normalized); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CanonicalJSON serializes an EvidenceReport to byte-deterministic JSON with sorted keys.
func (er *EvidenceReport) CanonicalJSON() ([]byte, error) {
	sortedEvidence := make([]domain.EvidenceRecord, len(er.Evidence))
	copy(sortedEvidence, er.Evidence)
	sort.Slice(sortedEvidence, func(i, j int) bool {
		locI := sortedEvidence[i].Locator.String()
		locJ := sortedEvidence[j].Locator.String()
		if locI != locJ {
			return locI < locJ
		}
		return sortedEvidence[i].Origin < sortedEvidence[j].Origin
	})

	sortedProfiles := make([]string, len(er.DeploymentProfiles))
	copy(sortedProfiles, er.DeploymentProfiles)
	sort.Strings(sortedProfiles)

	sortedOrigins := make([]domain.DerivationOrigin, len(er.Origins))
	copy(sortedOrigins, er.Origins)
	sort.Slice(sortedOrigins, func(i, j int) bool { return sortedOrigins[i] < sortedOrigins[j] })

	sortedStatuses := make([]domain.FreshnessStatus, len(er.Statuses))
	copy(sortedStatuses, er.Statuses)
	sort.Slice(sortedStatuses, func(i, j int) bool { return sortedStatuses[i] < sortedStatuses[j] })

	normalized := EvidenceReport{
		TargetID:           er.TargetID,
		NodeID:             er.NodeID,
		EdgeID:             er.EdgeID,
		Node:               er.Node,
		Relationship:       er.Relationship,
		Evidence:           sortedEvidence,
		DeploymentProfiles: sortedProfiles,
		Origins:            sortedOrigins,
		Statuses:           sortedStatuses,
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(normalized); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderMarkdown formats a ContextSlice as compact Markdown optimized for LLM token budgets.
func (cs *ContextSlice) RenderMarkdown() string {
	var sb strings.Builder

	focalLoc := primaryLocatorCitation(cs.FocalNode.Evidence)
	sb.WriteString(fmt.Sprintf("# Context Slice: %s (%s)\n", cs.FocalNode.Name, cs.FocalNode.Kind))
	sb.WriteString(fmt.Sprintf("- ID: `%s`\n", cs.FocalNode.ID))
	sb.WriteString(fmt.Sprintf("- Tier: `%s`\n", cs.Tier))
	if focalLoc != "" {
		sb.WriteString(fmt.Sprintf("- Locator: %s\n", focalLoc))
	}

	// Stale notice on focal node if present
	if hasStaleEvidence(cs.FocalNode.Evidence) {
		sb.WriteString(fmt.Sprintf("> ⚠️ NOTICE: Stale evidence detected for focal node `%s`\n", cs.FocalNode.ID))
	}

	// Relationships
	sb.WriteString(fmt.Sprintf("\n## Relationships (%d)\n", len(cs.Relationships)))
	if len(cs.Relationships) == 0 {
		sb.WriteString("- (none)\n")
	} else {
		for _, rel := range cs.Relationships {
			sb.WriteString(formatRelationshipEdge(rel))
			sb.WriteString("\n")
			if hasStaleEvidence(rel.Evidence) {
				sb.WriteString(fmt.Sprintf("  > ⚠️ NOTICE: Stale evidence for edge `%s`\n", rel.ID))
			}
		}
	}

	// Neighbors
	sb.WriteString(fmt.Sprintf("\n## Neighbors (%d)\n", len(cs.Neighbors)))
	if len(cs.Neighbors) == 0 {
		sb.WriteString("- (none)\n")
	} else {
		for _, n := range cs.Neighbors {
			nLoc := primaryLocatorCitation(n.Evidence)
			sb.WriteString(fmt.Sprintf("- `%s` (%s)%s\n", n.ID, n.Kind, nLoc))
			if hasStaleEvidence(n.Evidence) {
				sb.WriteString(fmt.Sprintf("  > ⚠️ NOTICE: Stale evidence for `%s`\n", n.ID))
			}
		}
	}

	// Contract definitions (expanded in contracts and deep tiers)
	if cs.Tier == TierContracts || cs.Tier == TierDeep {
		var contractNodes []domain.Node
		if cs.FocalNode.Kind == domain.NodeContract {
			contractNodes = append(contractNodes, cs.FocalNode)
		}
		for _, n := range cs.Neighbors {
			if n.Kind == domain.NodeContract {
				contractNodes = append(contractNodes, n)
			}
		}

		if len(contractNodes) > 0 {
			sb.WriteString(fmt.Sprintf("\n## Contract Definitions (%d)\n", len(contractNodes)))
			for _, c := range contractNodes {
				renderContractMarkdown(&sb, c)
			}
		}
	}

	// Deep tier discrete evidence records
	if cs.Tier == TierDeep {
		allNodes := append([]domain.Node{cs.FocalNode}, cs.Neighbors...)
		var totalEvidence int
		for _, n := range allNodes {
			totalEvidence += len(n.Evidence)
		}
		for _, r := range cs.Relationships {
			totalEvidence += len(r.Evidence)
		}

		sb.WriteString(fmt.Sprintf("\n## Discrete Evidence Records (%d)\n", totalEvidence))
		for _, n := range allNodes {
			for _, ev := range n.Evidence {
				sb.WriteString(fmt.Sprintf("- `%s`: %s (%s) | %s | [%s]\n",
					n.ID, ev.Origin, ev.DeploymentProfile, ev.Status, ev.Locator.String()))
			}
		}
		for _, r := range cs.Relationships {
			for _, ev := range r.Evidence {
				sb.WriteString(fmt.Sprintf("- `%s`: %s (%s) | %s | [%s]\n",
					r.ID, ev.Origin, ev.DeploymentProfile, ev.Status, ev.Locator.String()))
			}
		}
	}

	// Unsupported boundaries
	if len(cs.UnsupportedBoundaries) > 0 {
		sb.WriteString(fmt.Sprintf("\n## Unsupported Boundaries (%d)\n", len(cs.UnsupportedBoundaries)))
		for _, ub := range cs.UnsupportedBoundaries {
			ubLoc := ""
			if ub.Locator.Path != "" {
				ubLoc = fmt.Sprintf(" [%s]", ub.Locator.String())
			}
			sb.WriteString(fmt.Sprintf("> ⚠️ UNSUPPORTED: `%s` - %s%s\n", ub.Path, ub.Reason, ubLoc))
		}
	}

	return sb.String()
}

// RenderMarkdown formats an ImpactBriefing as compact Markdown.
func (ib *ImpactBriefing) RenderMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Impact Briefing: %s (%s)\n", ib.TargetID, ib.TargetKind))
	sb.WriteString(fmt.Sprintf("- Traversal Depth: %d\n", ib.Depth))

	// Upstream Services
	sb.WriteString(fmt.Sprintf("\n## Upstream Services (%d)\n", len(ib.UpstreamServices)))
	if len(ib.UpstreamServices) == 0 {
		sb.WriteString("- (none)\n")
	} else {
		for _, svc := range ib.UpstreamServices {
			loc := primaryLocatorCitation(svc.Evidence)
			sb.WriteString(fmt.Sprintf("- `%s` (%s)%s\n", svc.ID, svc.Kind, loc))
			if hasStaleEvidence(svc.Evidence) {
				sb.WriteString(fmt.Sprintf("  > ⚠️ NOTICE: Stale evidence for `%s`\n", svc.ID))
			}
		}
	}

	// Affected Contracts
	sb.WriteString(fmt.Sprintf("\n## Affected Contracts (%d)\n", len(ib.AffectedContracts)))
	if len(ib.AffectedContracts) == 0 {
		sb.WriteString("- (none)\n")
	} else {
		for _, c := range ib.AffectedContracts {
			loc := primaryLocatorCitation(c.Evidence)
			syntax := ""
			if c.Metadata != nil {
				if s, ok := c.Metadata["syntax"].(string); ok && s != "" {
					syntax = fmt.Sprintf(" [%s]", s)
				}
			}
			sb.WriteString(fmt.Sprintf("- `%s`%s%s\n", c.ID, syntax, loc))
		}
	}

	// Downstream Services
	sb.WriteString(fmt.Sprintf("\n## Downstream Services (%d)\n", len(ib.DownstreamServices)))
	if len(ib.DownstreamServices) == 0 {
		sb.WriteString("- (none)\n")
	} else {
		for _, svc := range ib.DownstreamServices {
			loc := primaryLocatorCitation(svc.Evidence)
			sb.WriteString(fmt.Sprintf("- `%s` (%s)%s\n", svc.ID, svc.Kind, loc))
			if hasStaleEvidence(svc.Evidence) {
				sb.WriteString(fmt.Sprintf("  > ⚠️ NOTICE: Stale evidence for `%s`\n", svc.ID))
			}
		}
	}

	// Mediation Edges
	sb.WriteString(fmt.Sprintf("\n## Mediation Relationships (%d)\n", len(ib.MediationEdges)))
	if len(ib.MediationEdges) == 0 {
		sb.WriteString("- (none)\n")
	} else {
		for _, rel := range ib.MediationEdges {
			sb.WriteString(formatRelationshipEdge(rel))
			sb.WriteString("\n")
		}
	}

	// Unsupported Boundaries
	if len(ib.UnsupportedBoundaries) > 0 {
		sb.WriteString(fmt.Sprintf("\n## Unsupported Boundaries (%d)\n", len(ib.UnsupportedBoundaries)))
		for _, ub := range ib.UnsupportedBoundaries {
			ubLoc := ""
			if ub.Locator.Path != "" {
				ubLoc = fmt.Sprintf(" [%s]", ub.Locator.String())
			}
			sb.WriteString(fmt.Sprintf("> ⚠️ UNSUPPORTED: `%s` - %s%s\n", ub.Path, ub.Reason, ubLoc))
		}
	}

	return sb.String()
}

// RenderMarkdown formats an EvidenceReport as compact Markdown.
func (er *EvidenceReport) RenderMarkdown() string {
	var sb strings.Builder

	targetName := er.TargetID
	if targetName == "" {
		if er.NodeID != "" {
			targetName = er.NodeID
		} else {
			targetName = er.EdgeID
		}
	}

	sb.WriteString(fmt.Sprintf("# Evidence Report: %s\n", targetName))
	if len(er.DeploymentProfiles) > 0 {
		sb.WriteString(fmt.Sprintf("- Profiles: %s\n", strings.Join(er.DeploymentProfiles, ", ")))
	}
	if len(er.Origins) > 0 {
		originStrs := make([]string, len(er.Origins))
		for i, o := range er.Origins {
			originStrs[i] = string(o)
		}
		sb.WriteString(fmt.Sprintf("- Origins: %s\n", strings.Join(originStrs, ", ")))
	}
	if len(er.Statuses) > 0 {
		statusStrs := make([]string, len(er.Statuses))
		for i, s := range er.Statuses {
			statusStrs[i] = string(s)
		}
		sb.WriteString(fmt.Sprintf("- Statuses: %s\n", strings.Join(statusStrs, ", ")))
	}

	sb.WriteString(fmt.Sprintf("\n## Evidence Records (%d)\n", len(er.Evidence)))
	if len(er.Evidence) == 0 {
		sb.WriteString("- (none)\n")
	} else {
		for _, ev := range er.Evidence {
			sb.WriteString(fmt.Sprintf("- [%s] Origin: %s | Profile: %s | Status: %s\n",
				ev.Locator.String(), ev.Origin, ev.DeploymentProfile, ev.Status))
			if ev.Status == domain.StatusStale {
				sb.WriteString(fmt.Sprintf("  > ⚠️ NOTICE: Stale locator citation [%s]\n", ev.Locator.String()))
			}
		}
	}

	return sb.String()
}

// RenderBoundaryMatches formats search matches as compact Markdown.
func RenderBoundaryMatches(matches []BoundaryMatch) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Boundary Matches (%d)\n", len(matches)))
	if len(matches) == 0 {
		sb.WriteString("- No matches found.\n")
		return sb.String()
	}

	for _, m := range matches {
		if m.Kind == "unsupported" {
			sb.WriteString(fmt.Sprintf("> ⚠️ UNSUPPORTED: `%s` (matched on: %s) - %s\n", m.ID, m.MatchedOn, m.Details))
		} else if m.Kind == "contract_method" {
			sb.WriteString(fmt.Sprintf("- `%s`: RPC Method `%s` (matched on: %s)\n", m.ID, m.Details, m.MatchedOn))
		} else {
			sb.WriteString(fmt.Sprintf("- `%s` (%s, matched on: %s)\n", m.ID, m.Kind, m.MatchedOn))
		}
	}
	return sb.String()
}

func formatRelationshipEdge(rel domain.Relationship) string {
	locStr := ""
	if len(rel.Evidence) > 0 && rel.Evidence[0].Locator.Path != "" {
		locStr = fmt.Sprintf(" [%s]", rel.Evidence[0].Locator.String())
	}

	switch rel.Verb {
	case domain.VerbProvides:
		return fmt.Sprintf("- `%s` -provides-> `%s`%s", rel.SourceID, rel.TargetID, locStr)
	case domain.VerbConsumes:
		return fmt.Sprintf("- `%s` <-consumes- `%s`%s", rel.TargetID, rel.SourceID, locStr)
	case domain.VerbImports:
		return fmt.Sprintf("- `%s` -imports-> `%s`%s", rel.SourceID, rel.TargetID, locStr)
	case domain.VerbAccesses:
		return fmt.Sprintf("- `%s` -accesses-> `%s`%s", rel.SourceID, rel.TargetID, locStr)
	default:
		return fmt.Sprintf("- `%s` -%s-> `%s`%s", rel.SourceID, rel.Verb, rel.TargetID, locStr)
	}
}

func primaryLocatorCitation(ev []domain.EvidenceRecord) string {
	if len(ev) > 0 && ev[0].Locator.Path != "" {
		return fmt.Sprintf(" [%s]", ev[0].Locator.String())
	}
	return ""
}

func hasStaleEvidence(ev []domain.EvidenceRecord) bool {
	for _, e := range ev {
		if e.Status == domain.StatusStale {
			return true
		}
	}
	return false
}

func renderContractMarkdown(sb *strings.Builder, c domain.Node) {
	syntax := ""
	if c.Metadata != nil {
		if s, ok := c.Metadata["syntax"].(string); ok && s != "" {
			syntax = fmt.Sprintf(" (%s)", s)
		}
	}
	loc := primaryLocatorCitation(c.Evidence)
	sb.WriteString(fmt.Sprintf("- **%s**%s%s\n", c.Name, syntax, loc))

	if c.Metadata == nil {
		return
	}

	// Render RPC definitions if present
	if rDefs, ok := c.Metadata["rpc_definitions"].([]map[string]string); ok && len(rDefs) > 0 {
		sb.WriteString("  - Methods:\n")
		for _, r := range rDefs {
			sb.WriteString(fmt.Sprintf("    - `%s(%s)` -> `%s`\n", r["name"], r["request"], r["response"]))
		}
	} else if methods, ok := c.Metadata["methods"].([]string); ok && len(methods) > 0 {
		sb.WriteString(fmt.Sprintf("  - Methods: %s\n", strings.Join(methods, ", ")))
	}

	// Render messages if present
	if msgs, ok := c.Metadata["messages"].([]string); ok && len(msgs) > 0 {
		sb.WriteString(fmt.Sprintf("  - Messages: %s\n", strings.Join(msgs, ", ")))
	}
}
