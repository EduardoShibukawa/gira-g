package slices

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// SliceTier defines the progressive disclosure depth for context slices.
type SliceTier string

const (
	TierSummary   SliceTier = "summary"
	TierContracts SliceTier = "contracts"
	TierDeep      SliceTier = "deep"
)

// GraphView defines the architectural projection filter for context slices.
type GraphView string

const (
	ViewFull         GraphView = "full"
	ViewArchitecture GraphView = "architecture"
	ViewGovernance   GraphView = "governance"
)

// ImpactBriefing documents affected services, contracts, and mediation edges
// when an architectural change occurs at a given service or contract.
type ImpactBriefing struct {
	TargetID         string          `json:"target_id"`
	TargetKind       domain.NodeType `json:"target_kind"`
	Depth            int             `json:"depth"`
	UpstreamServices []domain.Node   `json:"upstream_services"`
	// UpstreamLibraries are the Library Components the target imports. They are
	// kept out of upstream_services, which lists only Services and Infrastructure.
	UpstreamLibraries     []domain.Node                `json:"upstream_libraries,omitempty"`
	DownstreamServices    []domain.Node                `json:"downstream_services"`
	AffectedContracts     []domain.Node                `json:"affected_contracts"`
	MediationEdges        []domain.Relationship        `json:"mediation_edges"`
	UnsupportedBoundaries []domain.UnsupportedBoundary `json:"unsupported_boundaries,omitempty"`
	Diff                  *DiffTargets                 `json:"diff,omitempty"`
	Budget                *Budget                      `json:"budget,omitempty"`
}

// ContextSlice encapsulates a bounded architectural subgraph around a focal node.
type ContextSlice struct {
	FocalNode             domain.Node                  `json:"focal_node"`
	Tier                  SliceTier                    `json:"tier"`
	Neighbors             []domain.Node                `json:"neighbors"`
	Relationships         []domain.Relationship        `json:"relationships"`
	UnsupportedBoundaries []domain.UnsupportedBoundary `json:"unsupported_boundaries,omitempty"`
	Budget                *Budget                      `json:"budget,omitempty"`
}

// EvidenceReport details provenance, deployment profiles, and freshness for a node or edge.
type EvidenceReport struct {
	TargetID           string                    `json:"target_id,omitempty"`
	NodeID             string                    `json:"node_id,omitempty"`
	EdgeID             string                    `json:"edge_id,omitempty"`
	Node               *domain.Node              `json:"node,omitempty"`
	Relationship       *domain.Relationship      `json:"relationship,omitempty"`
	Evidence           []domain.EvidenceRecord   `json:"evidence"`
	DeploymentProfiles []string                  `json:"deployment_profiles"`
	Origins            []domain.DerivationOrigin `json:"origins"`
	Statuses           []domain.FreshnessStatus  `json:"statuses"`
}

// BoundaryMatch represents a discovery match across graph components or unhandled files.
type BoundaryMatch struct {
	Kind        string                      `json:"kind"`
	ID          string                      `json:"id"`
	Name        string                      `json:"name,omitempty"`
	MatchedOn   string                      `json:"matched_on"`
	Details     string                      `json:"details,omitempty"`
	Node        *domain.Node                `json:"node,omitempty"`
	Unsupported *domain.UnsupportedBoundary `json:"unsupported,omitempty"`
}

// Engine implements fast context slicing and impact analysis over a domain.MacroMap.
type Engine struct {
	macroMap           *domain.MacroMap
	nodesByID          map[string]domain.Node
	edgesByID          map[string]domain.Relationship
	providesByContract map[string][]domain.Relationship // contractID -> relationships providing it
	consumesByContract map[string][]domain.Relationship // contractID -> relationships consuming it
	providesByService  map[string][]domain.Relationship // serviceID -> relationships where service provides
	consumesByService  map[string][]domain.Relationship // serviceID -> relationships where service consumes
	edgesByNode        map[string][]domain.Relationship // nodeID -> all incident edges
	unsupportedByPath  map[string]domain.UnsupportedBoundary
	allUnsupported     []domain.UnsupportedBoundary
}

// NewEngine initializes an indexing Engine for fast graph traversals.
func NewEngine(m *domain.MacroMap) *Engine {
	if m == nil {
		m = domain.NewMacroMap("")
	}

	e := &Engine{
		macroMap:           m,
		nodesByID:          make(map[string]domain.Node, len(m.Nodes)),
		edgesByID:          make(map[string]domain.Relationship, len(m.Relationships)),
		providesByContract: make(map[string][]domain.Relationship),
		consumesByContract: make(map[string][]domain.Relationship),
		providesByService:  make(map[string][]domain.Relationship),
		consumesByService:  make(map[string][]domain.Relationship),
		edgesByNode:        make(map[string][]domain.Relationship),
		unsupportedByPath:  make(map[string]domain.UnsupportedBoundary, len(m.Unsupported)),
		allUnsupported:     make([]domain.UnsupportedBoundary, len(m.Unsupported)),
	}

	for _, n := range m.Nodes {
		e.nodesByID[n.ID] = n
	}

	for _, rel := range m.Relationships {
		e.edgesByID[rel.ID] = rel
		e.edgesByNode[rel.SourceID] = append(e.edgesByNode[rel.SourceID], rel)
		e.edgesByNode[rel.TargetID] = append(e.edgesByNode[rel.TargetID], rel)

		srcNode, srcFound := e.nodesByID[rel.SourceID]
		tgtNode, tgtFound := e.nodesByID[rel.TargetID]

		switch rel.Verb {
		case domain.VerbProvides:
			if tgtFound && tgtNode.Kind == domain.NodeContract {
				e.providesByContract[rel.TargetID] = append(e.providesByContract[rel.TargetID], rel)
				e.providesByService[rel.SourceID] = append(e.providesByService[rel.SourceID], rel)
			} else if srcFound && srcNode.Kind == domain.NodeContract {
				e.providesByContract[rel.SourceID] = append(e.providesByContract[rel.SourceID], rel)
				e.providesByService[rel.TargetID] = append(e.providesByService[rel.TargetID], rel)
			} else {
				e.providesByService[rel.SourceID] = append(e.providesByService[rel.SourceID], rel)
			}

		case domain.VerbConsumes:
			if tgtFound && tgtNode.Kind == domain.NodeContract {
				e.consumesByContract[rel.TargetID] = append(e.consumesByContract[rel.TargetID], rel)
				e.consumesByService[rel.SourceID] = append(e.consumesByService[rel.SourceID], rel)
			} else if srcFound && srcNode.Kind == domain.NodeContract {
				e.consumesByContract[rel.SourceID] = append(e.consumesByContract[rel.SourceID], rel)
				e.consumesByService[rel.TargetID] = append(e.consumesByService[rel.TargetID], rel)
			} else {
				e.consumesByService[rel.SourceID] = append(e.consumesByService[rel.SourceID], rel)
			}
		}
	}

	for i, ub := range m.Unsupported {
		e.unsupportedByPath[ub.Path] = ub
		e.allUnsupported[i] = ub
	}

	return e
}

// GetImpactBriefing computes upstream dependencies, downstream consumers,
// mediating contracts, and unsupported boundaries affected by a change.
func (e *Engine) GetImpactBriefing(targetID string, depth int) (*ImpactBriefing, error) {
	targetNode, ok := e.nodesByID[targetID]
	if !ok {
		return nil, fmt.Errorf("target node not found: %s", targetID)
	}

	if depth <= 0 {
		depth = 1
	}

	upstreamServicesMap := make(map[string]domain.Node)
	downstreamServicesMap := make(map[string]domain.Node)
	affectedContractsMap := make(map[string]domain.Node)
	mediationEdgesMap := make(map[string]domain.Relationship)

	if targetNode.Kind == domain.NodeContract {
		affectedContractsMap[targetNode.ID] = targetNode

		// Upstream producers of contract
		for _, rel := range e.providesByContract[targetNode.ID] {
			svcID := rel.SourceID
			if svcID == targetNode.ID {
				svcID = rel.TargetID
			}
			if svc, found := e.nodesByID[svcID]; found {
				upstreamServicesMap[svcID] = svc
			}
			mediationEdgesMap[rel.ID] = rel
		}

		// Downstream consumers of contract
		for _, rel := range e.consumesByContract[targetNode.ID] {
			svcID := rel.SourceID
			if svcID == targetNode.ID {
				svcID = rel.TargetID
			}
			if svc, found := e.nodesByID[svcID]; found {
				downstreamServicesMap[svcID] = svc
			}
			mediationEdgesMap[rel.ID] = rel
		}

		// Multi-hop expansion if depth > 1
		if depth > 1 {
			downstreamFrontier := make(map[string]struct{})
			for id := range downstreamServicesMap {
				downstreamFrontier[id] = struct{}{}
			}
			upstreamFrontier := make(map[string]struct{})
			for id := range upstreamServicesMap {
				upstreamFrontier[id] = struct{}{}
			}

			for d := 2; d <= depth; d++ {
				nextDownstream := make(map[string]struct{})
				for svcID := range downstreamFrontier {
					// Contracts provided by downstream service
					for _, rel := range e.providesByService[svcID] {
						cID := rel.TargetID
						if cID == svcID {
							cID = rel.SourceID
						}
						if cNode, found := e.nodesByID[cID]; found && cNode.Kind == domain.NodeContract {
							affectedContractsMap[cID] = cNode
							mediationEdgesMap[rel.ID] = rel
							for _, cRel := range e.consumesByContract[cID] {
								consumerID := cRel.SourceID
								if consumerID == cID {
									consumerID = cRel.TargetID
								}
								if consumerID != svcID {
									if consumerNode, found := e.nodesByID[consumerID]; found {
										if _, seen := downstreamServicesMap[consumerID]; !seen {
											downstreamServicesMap[consumerID] = consumerNode
											nextDownstream[consumerID] = struct{}{}
										}
									}
									mediationEdgesMap[cRel.ID] = cRel
								}
							}
						}
					}
				}
				downstreamFrontier = nextDownstream

				nextUpstream := make(map[string]struct{})
				for svcID := range upstreamFrontier {
					// Contracts consumed by upstream service
					for _, rel := range e.consumesByService[svcID] {
						cID := rel.TargetID
						if cID == svcID {
							cID = rel.SourceID
						}
						if cNode, found := e.nodesByID[cID]; found && cNode.Kind == domain.NodeContract {
							affectedContractsMap[cID] = cNode
							mediationEdgesMap[rel.ID] = rel
							for _, pRel := range e.providesByContract[cID] {
								providerID := pRel.SourceID
								if providerID == cID {
									providerID = pRel.TargetID
								}
								if providerID != svcID {
									if providerNode, found := e.nodesByID[providerID]; found {
										if _, seen := upstreamServicesMap[providerID]; !seen {
											upstreamServicesMap[providerID] = providerNode
											nextUpstream[providerID] = struct{}{}
										}
									}
									mediationEdgesMap[pRel.ID] = pRel
								}
							}
						}
					}
				}
				upstreamFrontier = nextUpstream
			}
		}

	} else {
		// Target is Service (or other architectural node)
		downstreamFrontier := map[string]struct{}{targetID: {}}
		upstreamFrontier := map[string]struct{}{targetID: {}}

		for d := 1; d <= depth; d++ {
			nextDownstream := make(map[string]struct{})
			for svcID := range downstreamFrontier {
				// 1. Contracts provided by svcID -> who consumes them?
				for _, rel := range e.providesByService[svcID] {
					cID := rel.TargetID
					if cID == svcID {
						cID = rel.SourceID
					}
					if cNode, found := e.nodesByID[cID]; found {
						if cNode.Kind == domain.NodeContract {
							affectedContractsMap[cID] = cNode
						}
						mediationEdgesMap[rel.ID] = rel

						for _, cRel := range e.consumesByContract[cID] {
							consumerID := cRel.SourceID
							if consumerID == cID {
								consumerID = cRel.TargetID
							}
							if consumerID != targetID {
								if consumerNode, found := e.nodesByID[consumerID]; found {
									if _, seen := downstreamServicesMap[consumerID]; !seen {
										downstreamServicesMap[consumerID] = consumerNode
										nextDownstream[consumerID] = struct{}{}
									}
								}
								mediationEdgesMap[cRel.ID] = cRel
							}
						}
					}
				}

				// 2. Direct consumers of svcID
				for _, rel := range e.edgesByNode[svcID] {
					if rel.TargetID == svcID && rel.Verb == domain.VerbConsumes && rel.SourceID != targetID {
						consumerID := rel.SourceID
						if consumerNode, found := e.nodesByID[consumerID]; found {
							if _, seen := downstreamServicesMap[consumerID]; !seen {
								downstreamServicesMap[consumerID] = consumerNode
								nextDownstream[consumerID] = struct{}{}
							}
						}
						mediationEdgesMap[rel.ID] = rel
					}
				}
			}
			downstreamFrontier = nextDownstream

			nextUpstream := make(map[string]struct{})
			for svcID := range upstreamFrontier {
				// 1. Contracts consumed by svcID -> who provides them?
				for _, rel := range e.consumesByService[svcID] {
					cID := rel.TargetID
					if cID == svcID {
						cID = rel.SourceID
					}
					if cNode, found := e.nodesByID[cID]; found {
						if cNode.Kind == domain.NodeContract {
							affectedContractsMap[cID] = cNode
							mediationEdgesMap[rel.ID] = rel

							for _, pRel := range e.providesByContract[cID] {
								providerID := pRel.SourceID
								if providerID == cID {
									providerID = pRel.TargetID
								}
								if providerID != targetID {
									if providerNode, found := e.nodesByID[providerID]; found {
										if _, seen := upstreamServicesMap[providerID]; !seen {
											upstreamServicesMap[providerID] = providerNode
											nextUpstream[providerID] = struct{}{}
										}
									}
									mediationEdgesMap[pRel.ID] = pRel
								}
							}
						} else {
							// Direct service consumption
							if cID != targetID {
								if _, seen := upstreamServicesMap[cID]; !seen {
									upstreamServicesMap[cID] = cNode
									if cNode.Kind == domain.NodeService {
										nextUpstream[cID] = struct{}{}
									}
								}
								mediationEdgesMap[rel.ID] = rel
							}
						}
					}
				}

				// 2. Direct upstream dependencies (accesses infrastructure, imports library)
				for _, rel := range e.edgesByNode[svcID] {
					if rel.SourceID == svcID && (rel.Verb == domain.VerbAccesses || rel.Verb == domain.VerbImports) {
						depID := rel.TargetID
						if depNode, found := e.nodesByID[depID]; found && depID != targetID {
							if _, seen := upstreamServicesMap[depID]; !seen {
								upstreamServicesMap[depID] = depNode
								if depNode.Kind == domain.NodeService {
									nextUpstream[depID] = struct{}{}
								}
							}
							mediationEdgesMap[rel.ID] = rel
						}
					}
				}
			}
			upstreamFrontier = nextUpstream
		}
	}

	upstreamList := make([]domain.Node, 0, len(upstreamServicesMap))
	var librariesList []domain.Node
	for _, n := range upstreamServicesMap {
		if n.Kind == domain.NodeLibrary {
			librariesList = append(librariesList, cloneNode(n))
			continue
		}
		upstreamList = append(upstreamList, cloneNode(n))
	}
	sort.Slice(upstreamList, func(i, j int) bool { return upstreamList[i].ID < upstreamList[j].ID })
	sort.Slice(librariesList, func(i, j int) bool { return librariesList[i].ID < librariesList[j].ID })

	downstreamList := make([]domain.Node, 0, len(downstreamServicesMap))
	for _, n := range downstreamServicesMap {
		downstreamList = append(downstreamList, cloneNode(n))
	}
	sort.Slice(downstreamList, func(i, j int) bool { return downstreamList[i].ID < downstreamList[j].ID })

	contractsList := make([]domain.Node, 0, len(affectedContractsMap))
	for _, n := range affectedContractsMap {
		contractsList = append(contractsList, cloneNode(n))
	}
	sort.Slice(contractsList, func(i, j int) bool { return contractsList[i].ID < contractsList[j].ID })

	edgesList := make([]domain.Relationship, 0, len(mediationEdgesMap))
	for _, r := range mediationEdgesMap {
		edgesList = append(edgesList, cloneRel(r))
	}
	sort.Slice(edgesList, func(i, j int) bool { return edgesList[i].ID < edgesList[j].ID })

	// Correlate relevant unsupported boundaries
	componentNodes := append([]domain.Node{targetNode}, upstreamList...)
	componentNodes = append(componentNodes, downstreamList...)
	componentNodes = append(componentNodes, librariesList...)
	componentNodes = append(componentNodes, contractsList...)

	unsupportedMatches := make([]domain.UnsupportedBoundary, 0)
	seenUB := make(map[string]struct{})
	for _, ub := range e.allUnsupported {
		for _, comp := range componentNodes {
			if e.isUnsupportedRelated(ub, comp) {
				if _, seen := seenUB[ub.Path]; !seen {
					seenUB[ub.Path] = struct{}{}
					unsupportedMatches = append(unsupportedMatches, ub)
				}
				break
			}
		}
	}
	sort.Slice(unsupportedMatches, func(i, j int) bool { return unsupportedMatches[i].Path < unsupportedMatches[j].Path })

	return &ImpactBriefing{
		TargetID:              targetID,
		TargetKind:            targetNode.Kind,
		Depth:                 depth,
		UpstreamServices:      upstreamList,
		UpstreamLibraries:     librariesList,
		DownstreamServices:    downstreamList,
		AffectedContracts:     contractsList,
		MediationEdges:        edgesList,
		UnsupportedBoundaries: unsupportedMatches,
	}, nil
}

// GetContextSlice extracts a bounded subgraph around nodeID for a given SliceTier and default ViewFull.
func (e *Engine) GetContextSlice(nodeID string, tier SliceTier) (*ContextSlice, error) {
	return e.GetContextSliceWithView(nodeID, tier, ViewFull)
}

// isNodeAllowedInView checks if a node kind is eligible for the specified GraphView.
func isNodeAllowedInView(kind domain.NodeType, view GraphView) bool {
	switch view {
	case ViewArchitecture:
		return kind == domain.NodeService || kind == domain.NodeContract || kind == domain.NodeInfrastructure || kind == domain.NodeLibrary
	case ViewGovernance:
		return kind == domain.NodePolicy || kind == domain.NodeTask || kind == domain.NodeSkill
	default:
		return true
	}
}

// GetContextSliceWithView extracts a bounded subgraph around nodeID for a given SliceTier and GraphView.
func (e *Engine) GetContextSliceWithView(nodeID string, tier SliceTier, view GraphView) (*ContextSlice, error) {
	node, ok := e.nodesByID[nodeID]
	if !ok {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	if view == "" {
		view = ViewFull
	}

	neighborsMap := make(map[string]domain.Node)
	relsMap := make(map[string]domain.Relationship)

	for _, rel := range e.edgesByNode[nodeID] {
		neighborID := rel.TargetID
		if neighborID == nodeID {
			neighborID = rel.SourceID
		}
		neighbor, found := e.nodesByID[neighborID]
		if !found {
			continue
		}

		if !isNodeAllowedInView(neighbor.Kind, view) {
			continue
		}

		neighborsMap[neighborID] = neighbor
		relsMap[rel.ID] = rel
	}

	focalCopy := cloneNode(node)
	neighbors := make([]domain.Node, 0, len(neighborsMap))
	for _, n := range neighborsMap {
		neighbors = append(neighbors, cloneNode(n))
	}

	relationships := make([]domain.Relationship, 0, len(relsMap))
	for _, r := range relsMap {
		relationships = append(relationships, cloneRel(r))
	}

	var relevantUnsupported []domain.UnsupportedBoundary

	switch tier {
	case TierSummary:
		// Strip contract method/message details, omit full evidence lists
		focalCopy.Metadata = stripSummaryDetails(focalCopy.Kind, focalCopy.Metadata)
		if len(focalCopy.Evidence) > 1 {
			focalCopy.Evidence = focalCopy.Evidence[:1]
		}

		for i := range neighbors {
			neighbors[i].Metadata = stripSummaryDetails(neighbors[i].Kind, neighbors[i].Metadata)
			if len(neighbors[i].Evidence) > 1 {
				neighbors[i].Evidence = neighbors[i].Evidence[:1]
			}
		}

		for i := range relationships {
			if len(relationships[i].Evidence) > 1 {
				relationships[i].Evidence = relationships[i].Evidence[:1]
			}
		}

	case TierContracts:
		// Keep contract RPC definitions and signatures in metadata, trim evidence lists
		if len(focalCopy.Evidence) > 1 {
			focalCopy.Evidence = focalCopy.Evidence[:1]
		}

		for i := range neighbors {
			if len(neighbors[i].Evidence) > 1 {
				neighbors[i].Evidence = neighbors[i].Evidence[:1]
			}
		}

		for i := range relationships {
			if len(relationships[i].Evidence) > 1 {
				relationships[i].Evidence = relationships[i].Evidence[:1]
			}
		}

	case TierDeep:
		// Full discrete Evidence records retained, correlate unsupported boundaries
		allNodes := append([]domain.Node{focalCopy}, neighbors...)
		seenUB := make(map[string]struct{})
		for _, ub := range e.allUnsupported {
			for _, n := range allNodes {
				if e.isUnsupportedRelated(ub, n) {
					if _, seen := seenUB[ub.Path]; !seen {
						seenUB[ub.Path] = struct{}{}
						relevantUnsupported = append(relevantUnsupported, ub)
					}
					break
				}
			}
		}
		sort.Slice(relevantUnsupported, func(i, j int) bool { return relevantUnsupported[i].Path < relevantUnsupported[j].Path })
	}

	sort.Slice(neighbors, func(i, j int) bool { return neighbors[i].ID < neighbors[j].ID })
	sort.Slice(relationships, func(i, j int) bool { return relationships[i].ID < relationships[j].ID })

	return &ContextSlice{
		FocalNode:             focalCopy,
		Tier:                  tier,
		Neighbors:             neighbors,
		Relationships:         relationships,
		UnsupportedBoundaries: relevantUnsupported,
	}, nil
}

// ExplainEvidence inspects and documents discrete evidence records for a node and/or edge.
func (e *Engine) ExplainEvidence(nodeID string, edgeID string) (*EvidenceReport, error) {
	if nodeID == "" && edgeID == "" {
		return nil, fmt.Errorf("either nodeID or edgeID must be provided")
	}

	report := &EvidenceReport{
		NodeID:             nodeID,
		EdgeID:             edgeID,
		Evidence:           make([]domain.EvidenceRecord, 0),
		DeploymentProfiles: make([]string, 0),
		Origins:            make([]domain.DerivationOrigin, 0),
		Statuses:           make([]domain.FreshnessStatus, 0),
	}

	var targetID string

	if nodeID != "" {
		node, ok := e.nodesByID[nodeID]
		if !ok {
			return nil, fmt.Errorf("node not found: %s", nodeID)
		}
		nodeCopy := cloneNode(node)
		report.Node = &nodeCopy
		targetID = nodeID
		report.Evidence = append(report.Evidence, node.Evidence...)
	}

	if edgeID != "" {
		edge, ok := e.edgesByID[edgeID]
		if !ok {
			return nil, fmt.Errorf("relationship not found: %s", edgeID)
		}
		edgeCopy := cloneRel(edge)
		report.Relationship = &edgeCopy
		if targetID == "" {
			targetID = edgeID
		} else {
			targetID = targetID + "+" + edgeID
		}
		report.Evidence = append(report.Evidence, edge.Evidence...)
	}

	report.TargetID = targetID

	profileSet := make(map[string]struct{})
	originSet := make(map[domain.DerivationOrigin]struct{})
	statusSet := make(map[domain.FreshnessStatus]struct{})

	for _, ev := range report.Evidence {
		if ev.DeploymentProfile != "" {
			profileSet[ev.DeploymentProfile] = struct{}{}
		}
		if ev.Origin != "" {
			originSet[ev.Origin] = struct{}{}
		}
		if ev.Status != "" {
			statusSet[ev.Status] = struct{}{}
		}
	}

	for p := range profileSet {
		report.DeploymentProfiles = append(report.DeploymentProfiles, p)
	}
	sort.Strings(report.DeploymentProfiles)

	for o := range originSet {
		report.Origins = append(report.Origins, o)
	}
	sort.Slice(report.Origins, func(i, j int) bool { return report.Origins[i] < report.Origins[j] })

	for s := range statusSet {
		report.Statuses = append(report.Statuses, s)
	}
	sort.Slice(report.Statuses, func(i, j int) bool { return report.Statuses[i] < report.Statuses[j] })

	return report, nil
}

// LocateBoundary performs substring and exact case-insensitive search across nodes, methods, and unhandled files.
func (e *Engine) LocateBoundary(query string) ([]BoundaryMatch, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	lowerQ := strings.ToLower(q)
	matches := make([]BoundaryMatch, 0)

	sortedNodes := make([]domain.Node, 0, len(e.nodesByID))
	for _, n := range e.nodesByID {
		sortedNodes = append(sortedNodes, n)
	}
	sort.Slice(sortedNodes, func(i, j int) bool { return sortedNodes[i].ID < sortedNodes[j].ID })

	for _, n := range sortedNodes {
		nodeCopy := cloneNode(n)

		if strings.Contains(strings.ToLower(n.ID), lowerQ) {
			matches = append(matches, BoundaryMatch{
				Kind:      string(n.Kind),
				ID:        n.ID,
				Name:      n.Name,
				MatchedOn: "id",
				Node:      &nodeCopy,
			})
		} else if strings.Contains(strings.ToLower(n.Name), lowerQ) {
			matches = append(matches, BoundaryMatch{
				Kind:      string(n.Kind),
				ID:        n.ID,
				Name:      n.Name,
				MatchedOn: "name",
				Node:      &nodeCopy,
			})
		} else if strings.EqualFold(string(n.Kind), q) {
			matches = append(matches, BoundaryMatch{
				Kind:      string(n.Kind),
				ID:        n.ID,
				Name:      n.Name,
				MatchedOn: "kind",
				Node:      &nodeCopy,
			})
		}

		if n.Metadata != nil {
			if mList, ok := n.Metadata["methods"].([]string); ok {
				for _, m := range mList {
					if strings.Contains(strings.ToLower(m), lowerQ) {
						matches = append(matches, BoundaryMatch{
							Kind:      "contract_method",
							ID:        n.ID,
							Name:      n.Name,
							MatchedOn: "method",
							Details:   m,
							Node:      &nodeCopy,
						})
					}
				}
			} else if mAnyList, ok := n.Metadata["methods"].([]any); ok {
				for _, mAny := range mAnyList {
					if mStr, ok := mAny.(string); ok && strings.Contains(strings.ToLower(mStr), lowerQ) {
						matches = append(matches, BoundaryMatch{
							Kind:      "contract_method",
							ID:        n.ID,
							Name:      n.Name,
							MatchedOn: "method",
							Details:   mStr,
							Node:      &nodeCopy,
						})
					}
				}
			}

			if rDefs, ok := n.Metadata["rpc_definitions"].([]map[string]string); ok {
				for _, rDef := range rDefs {
					mName := rDef["name"]
					if strings.Contains(strings.ToLower(mName), lowerQ) {
						alreadyAdded := false
						for _, ex := range matches {
							if ex.ID == n.ID && ex.Kind == "contract_method" && ex.Details == mName {
								alreadyAdded = true
								break
							}
						}
						if !alreadyAdded {
							details := fmt.Sprintf("%s(%s) returns (%s)", mName, rDef["request"], rDef["response"])
							matches = append(matches, BoundaryMatch{
								Kind:      "contract_method",
								ID:        n.ID,
								Name:      n.Name,
								MatchedOn: "method",
								Details:   details,
								Node:      &nodeCopy,
							})
						}
					}
				}
			}
		}
	}

	sortedUnsupported := make([]domain.UnsupportedBoundary, len(e.allUnsupported))
	copy(sortedUnsupported, e.allUnsupported)
	sort.Slice(sortedUnsupported, func(i, j int) bool { return sortedUnsupported[i].Path < sortedUnsupported[j].Path })

	for _, ub := range sortedUnsupported {
		ubCopy := ub
		if strings.Contains(strings.ToLower(ub.Path), lowerQ) {
			matches = append(matches, BoundaryMatch{
				Kind:        "unsupported",
				ID:          ub.Path,
				Name:        filepath.Base(ub.Path),
				MatchedOn:   "path",
				Details:     ub.Reason,
				Unsupported: &ubCopy,
			})
		} else if strings.Contains(strings.ToLower(ub.Reason), lowerQ) {
			matches = append(matches, BoundaryMatch{
				Kind:        "unsupported",
				ID:          ub.Path,
				Name:        filepath.Base(ub.Path),
				MatchedOn:   "reason",
				Details:     ub.Reason,
				Unsupported: &ubCopy,
			})
		}
	}

	return matches, nil
}

func (e *Engine) isUnsupportedRelated(ub domain.UnsupportedBoundary, node domain.Node) bool {
	lowerPath := strings.ToLower(ub.Path)
	lowerReason := strings.ToLower(ub.Reason)
	lowerName := strings.ToLower(node.Name)

	if lowerName != "" && (strings.Contains(lowerPath, lowerName) || strings.Contains(lowerReason, lowerName)) {
		return true
	}

	nodePrefix := strings.TrimPrefix(node.ID, string(node.Kind)+":")
	nodePrefix = strings.ToLower(nodePrefix)
	if nodePrefix != "" && (strings.Contains(lowerPath, nodePrefix) || strings.Contains(lowerReason, nodePrefix)) {
		return true
	}

	for _, ev := range node.Evidence {
		if ev.Locator.Path != "" {
			dir1 := filepath.Dir(ev.Locator.Path)
			dir2 := filepath.Dir(ub.Path)
			if dir1 != "." && dir1 == dir2 {
				return true
			}
		}
	}

	return false
}

func cloneNode(n domain.Node) domain.Node {
	c := n
	if n.Metadata != nil {
		c.Metadata = make(map[string]any, len(n.Metadata))
		for k, v := range n.Metadata {
			c.Metadata[k] = v
		}
	}
	if n.Evidence != nil {
		c.Evidence = make([]domain.EvidenceRecord, len(n.Evidence))
		copy(c.Evidence, n.Evidence)
	}
	return c
}

func cloneRel(r domain.Relationship) domain.Relationship {
	c := r
	if r.Evidence != nil {
		c.Evidence = make([]domain.EvidenceRecord, len(r.Evidence))
		copy(c.Evidence, r.Evidence)
	}
	return c
}

// stripSummaryDetails drops what the summary tier leaves to deeper tiers: a
// Contract's method and message definitions, and how a Service is built (its
// Dockerfile stages and base images).
func stripSummaryDetails(kind domain.NodeType, meta map[string]any) map[string]any {
	switch kind {
	case domain.NodeContract:
		return stripContractDetails(meta)
	case domain.NodeService:
		return stripKeys(meta, "stages", "base_images")
	}
	return meta
}

func stripKeys(meta map[string]any, keys ...string) map[string]any {
	if meta == nil {
		return nil
	}
	clean := make(map[string]any, len(meta))
	for k, v := range meta {
		clean[k] = v
	}
	for _, k := range keys {
		delete(clean, k)
	}
	return clean
}

func stripContractDetails(meta map[string]any) map[string]any {
	if meta == nil {
		return nil
	}
	clean := make(map[string]any)
	for k, v := range meta {
		if k == "methods" || k == "rpc_definitions" || k == "messages" {
			continue
		}
		clean[k] = v
	}
	return clean
}
