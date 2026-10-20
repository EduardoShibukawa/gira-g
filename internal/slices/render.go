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
		Budget:                cs.Budget,
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

	var sortedLibraries []domain.Node
	if len(ib.UpstreamLibraries) > 0 {
		sortedLibraries = make([]domain.Node, len(ib.UpstreamLibraries))
		copy(sortedLibraries, ib.UpstreamLibraries)
		sort.Slice(sortedLibraries, func(i, j int) bool { return sortedLibraries[i].ID < sortedLibraries[j].ID })
	}

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
		UpstreamLibraries:     sortedLibraries,
		DownstreamServices:    sortedDownstream,
		AffectedContracts:     sortedContracts,
		MediationEdges:        sortedMediation,
		UnsupportedBoundaries: sortedUnsupported,
		Diff:                  ib.Diff,
		Budget:                ib.Budget,
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
		Budget:             er.Budget,
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

	renderBudgetMarkdown(&sb, cs.Budget)
	return sb.String()
}

// RenderMarkdown formats an ImpactBriefing as compact Markdown.
func (ib *ImpactBriefing) RenderMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Impact Briefing: %s (%s)\n", ib.TargetID, ib.TargetKind))
	sb.WriteString(fmt.Sprintf("- Traversal Depth: %d\n", ib.Depth))
	if ib.Diff != nil {
		sb.WriteString(fmt.Sprintf("- Diff: `%s`, %d changed path(s)\n", ib.Diff.Range, len(ib.Diff.ChangedPaths)))
		for _, id := range ib.Diff.Targets {
			sb.WriteString(fmt.Sprintf("- Target: `%s`\n", id))
		}
		for _, p := range ib.Diff.FileLevelPaths {
			sb.WriteString(fmt.Sprintf("- Matched at file level (no cited range changed): `%s`\n", p))
		}
		for _, p := range ib.Diff.UnmatchedPaths {
			sb.WriteString(fmt.Sprintf("- Unknown to the index: `%s`\n", p))
		}
	}

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

	// Upstream Libraries (only when the target imports any)
	if len(ib.UpstreamLibraries) > 0 {
		sb.WriteString(fmt.Sprintf("\n## Upstream Libraries (%d)\n", len(ib.UpstreamLibraries)))
		for _, lib := range ib.UpstreamLibraries {
			sb.WriteString(fmt.Sprintf("- `%s`%s\n", lib.ID, primaryLocatorCitation(lib.Evidence)))
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

	renderBudgetMarkdown(&sb, ib.Budget)
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

// renderBudgetMarkdown writes the budget marker. It is never truncated.
func renderBudgetMarkdown(sb *strings.Builder, b *Budget) {
	if b == nil {
		return
	}
	sb.WriteString(fmt.Sprintf("\n## Budget\n- limit: %d, estimated: %d, truncated: %t\n- dropped: nodes %d, relationships %d, evidence %d\n",
		b.Limit, b.Estimated, b.Truncated, b.Dropped.Nodes, b.Dropped.Relationships, b.Dropped.Evidence))
	if b.EscalateHint != "" {
		sb.WriteString(fmt.Sprintf("- escalate: %s\n", b.EscalateHint))
	}
}

// cavemanLoc renders the compact one-fact locator `path:L<start>` for caveman
// output; the full repo@commit citation stays in JSON and Markdown.
func cavemanLoc(ev []domain.EvidenceRecord) string {
	for _, e := range ev {
		if e.Locator.Path != "" {
			if e.Locator.StartLine > 0 {
				return fmt.Sprintf("%s:L%d", e.Locator.Path, e.Locator.StartLine)
			}
			return e.Locator.Path
		}
	}
	return ""
}

func cavemanNodeLine(n domain.Node) string {
	loc := cavemanLoc(n.Evidence)
	if loc == "" {
		return fmt.Sprintf("%s[%s]", n.Kind, n.ID)
	}
	return fmt.Sprintf("%s[%s] @ %s", n.Kind, n.ID, loc)
}

// cavemanBudget writes the never-truncated budget marker for caveman output.
func cavemanBudget(sb *strings.Builder, b *Budget) {
	if b == nil {
		return
	}
	fmt.Fprintf(sb, "Budget: limit=%d estimated=%d truncated=%t\n", b.Limit, b.Estimated, b.Truncated)
	if b.Dropped.Nodes > 0 || b.Dropped.Relationships > 0 || b.Dropped.Evidence > 0 {
		fmt.Fprintf(sb, "Dropped: nodes=%d relationships=%d evidence=%d\n", b.Dropped.Nodes, b.Dropped.Relationships, b.Dropped.Evidence)
	}
	if b.EscalateHint != "" {
		fmt.Fprintf(sb, "Escalate: %s\n", b.EscalateHint)
	}
}

// renderCavemanContractFacts writes the contract metadata of a node, one fact
// per line, as far as the tier kept it: RPC definitions, then plain methods,
// then messages. Nothing is invented when the summary stripped the details (D3).
func renderCavemanContractFacts(sb *strings.Builder, c domain.Node, indent string) {
	if c.Metadata == nil {
		return
	}
	if rDefs, ok := c.Metadata["rpc_definitions"].([]map[string]string); ok {
		for _, r := range rDefs {
			fmt.Fprintf(sb, "%s- %s(%s) -> %s\n", indent, r["name"], r["request"], r["response"])
		}
		return
	}
	written := false
	if mAnyList, ok := c.Metadata["methods"].([]any); ok {
		for _, mAny := range mAnyList {
			if m, ok := mAny.(string); ok {
				fmt.Fprintf(sb, "%s- %s\n", indent, m)
				written = true
			}
		}
	} else if mList, ok := c.Metadata["methods"].([]string); ok {
		for _, m := range mList {
			fmt.Fprintf(sb, "%s- %s\n", indent, m)
			written = true
		}
	}
	if written {
		return
	}
	for _, msgs := range [][]string{stringSliceMeta(c.Metadata["messages"])} {
		for _, m := range msgs {
			fmt.Fprintf(sb, "%s- msg %s\n", indent, m)
		}
	}
}

func stringSliceMeta(v any) []string {
	switch l := v.(type) {
	case []string:
		return l
	case []any:
		var out []string
		for _, e := range l {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// RenderCaveman formats a ContextSlice as one-fact-per-line plain text for LLM
// budgets (INTENT 0001): no markdown, no AST, no warnings, deterministic order.
func (cs *ContextSlice) RenderCaveman() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Component: %s\n", cavemanNodeLine(cs.FocalNode))

	var contracts []domain.Node
	if cs.FocalNode.Kind == domain.NodeContract {
		contracts = append(contracts, cs.FocalNode)
	}
	for _, n := range cs.Neighbors {
		if n.Kind == domain.NodeContract {
			contracts = append(contracts, n)
		}
	}
	if len(contracts) > 0 {
		sb.WriteString("Contracts:\n")
		for _, c := range contracts {
			fmt.Fprintf(&sb, "  - %s\n", cavemanNodeLine(c))
			renderCavemanContractFacts(&sb, c, "    ")
		}
	}

	outDeps, inDeps := 0, 0
	for _, rel := range cs.Relationships {
		if rel.SourceID == cs.FocalNode.ID || rel.TargetID == cs.FocalNode.ID {
			if rel.SourceID == cs.FocalNode.ID {
				outDeps++
			} else {
				inDeps++
			}
		}
	}
	if outDeps > 0 {
		sb.WriteString("Dependencies:\n")
		for _, rel := range cs.Relationships {
			if rel.SourceID != cs.FocalNode.ID {
				continue
			}
			loc := cavemanLoc(rel.Evidence)
			if loc != "" {
				fmt.Fprintf(&sb, "  - %s %s @ %s\n", rel.Verb, rel.TargetID, loc)
			} else {
				fmt.Fprintf(&sb, "  - %s %s\n", rel.Verb, rel.TargetID)
			}
		}
	}
	if inDeps > 0 {
		sb.WriteString("Impact:\n")
		kindByID := map[string]domain.NodeType{}
		for _, n := range cs.Neighbors {
			kindByID[n.ID] = n.Kind
		}
		for _, rel := range cs.Relationships {
			if rel.TargetID != cs.FocalNode.ID {
				continue
			}
			kind := kindByID[rel.SourceID]
			if kind == "" {
				kind = domain.NodeType("Node")
			}
			loc := cavemanLoc(rel.Evidence)
			if loc != "" {
				fmt.Fprintf(&sb, "  - %s by %s[%s] @ %s\n", pastParticiple(rel.Verb), kind, rel.SourceID, loc)
			} else {
				fmt.Fprintf(&sb, "  - %s by %s[%s]\n", pastParticiple(rel.Verb), kind, rel.SourceID)
			}
		}
	}

	if len(cs.UnsupportedBoundaries) > 0 {
		sb.WriteString("Boundaries:\n")
		for _, ub := range cs.UnsupportedBoundaries {
			fmt.Fprintf(&sb, "  - UNSUPPORTED %s: %s\n", ub.Path, ub.Reason)
		}
	}

	cavemanBudget(&sb, cs.Budget)
	return strings.TrimRight(sb.String(), "\n")
}

// pastParticiple renders a relationship verb as the passive agent-facing fact
// ("consumes" → "consumed by"). Verbs outside the closed set fall back to the
// verb itself, never invented prose.
func pastParticiple(v domain.RelationVerb) string {
	switch v {
	case domain.VerbConsumes:
		return "consumed"
	case domain.VerbProvides:
		return "provided"
	case domain.VerbImports:
		return "imported"
	case domain.VerbAccesses:
		return "accessed"
	case domain.VerbImplements:
		return "implemented"
	case domain.VerbVerifies:
		return "verified"
	case domain.VerbGoverns:
		return "governed"
	}
	return string(v)
}

// RenderCaveman formats an ImpactBriefing as one-fact-per-line plain text.
func (ib *ImpactBriefing) RenderCaveman() string {
	var sb strings.Builder
	if ib.Diff != nil {
		fmt.Fprintf(&sb, "Target: diff %s\n", ib.Diff.Range)
	} else {
		fmt.Fprintf(&sb, "Target: %s[%s]\n", ib.TargetKind, ib.TargetID)
	}
	fmt.Fprintf(&sb, "Depth: %d\n", ib.Depth)
	writeGroup := func(title string, group []domain.Node) {
		if len(group) == 0 {
			return
		}
		fmt.Fprintf(&sb, "%s:\n", title)
		for _, n := range group {
			fmt.Fprintf(&sb, "  - %s\n", cavemanNodeLine(n))
		}
	}
	writeGroup("Upstream", ib.UpstreamServices)
	writeGroup("Libraries", ib.UpstreamLibraries)
	writeGroup("Downstream", ib.DownstreamServices)
	writeGroup("Contracts", ib.AffectedContracts)
	if ib.Diff != nil && len(ib.Diff.UnmatchedPaths) > 0 {
		sb.WriteString("Unmatched:\n")
		for _, p := range ib.Diff.UnmatchedPaths {
			fmt.Fprintf(&sb, "  - %s\n", p)
		}
	}
	if len(ib.UnsupportedBoundaries) > 0 {
		sb.WriteString("Boundaries:\n")
		for _, ub := range ib.UnsupportedBoundaries {
			fmt.Fprintf(&sb, "  - UNSUPPORTED %s: %s\n", ub.Path, ub.Reason)
		}
	}
	cavemanBudget(&sb, ib.Budget)
	return strings.TrimRight(sb.String(), "\n")
}

// RenderCaveman formats boundary matches one per line, in the canonical
// BoundaryResult order, unsupported last.
func (br *BoundaryResult) RenderCaveman() string {
	var sb strings.Builder
	for _, m := range br.Matches {
		if m.Unsupported != nil {
			fmt.Fprintf(&sb, "Boundary: %s: %s\n", m.Unsupported.Path, m.Unsupported.Reason)
			continue
		}
		kind := m.Kind
		loc := ""
		if m.Node != nil {
			kind = string(m.Node.Kind)
			loc = cavemanLoc(m.Node.Evidence)
		}
		if loc != "" {
			fmt.Fprintf(&sb, "Match: %s[%s] @ %s\n", kind, m.ID, loc)
		} else {
			fmt.Fprintf(&sb, "Match: %s[%s]\n", kind, m.ID)
		}
	}
	cavemanBudget(&sb, br.Budget)
	return strings.TrimRight(sb.String(), "\n")
}
