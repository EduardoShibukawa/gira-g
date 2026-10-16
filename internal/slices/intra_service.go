package slices

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// IntraContractRef cites one provided or consumed contract: the contract,
// the edge that proves the relation, and every evidence locator.
type IntraContractRef struct {
	ContractID   string           `json:"contract_id"`
	ContractName string           `json:"contract_name"`
	EdgeID       string           `json:"edge_id"`
	Verb         string           `json:"verb"`
	Locators     []domain.Locator `json:"locators"`
}

// IntraServiceDoc is the per-service intra-service view (F2) a later
// `gira doc` renders: entrypoints, symbols, provided/consumed contracts,
// imports split intra-module vs external, the call inventory, and the known
// unknowns. Every item carries a file+line locator; what no extractor
// covers appears under KnownUnknowns, never guessed.
type IntraServiceDoc struct {
	ServiceID          string                       `json:"service_id"`
	ServiceName        string                       `json:"service_name"`
	ServiceLocator     domain.Locator               `json:"service_locator"`
	Entrypoints        []domain.GoEntrypoint        `json:"entrypoints"`
	Packages           []domain.GoPackageSymbol     `json:"packages"`
	Types              []domain.GoTypeSymbol        `json:"types"`
	Funcs              []domain.GoFuncSymbol        `json:"funcs"`
	ProvidedContracts  []IntraContractRef           `json:"provided_contracts"`
	ConsumedContracts  []IntraContractRef           `json:"consumed_contracts"`
	IntraModuleImports []domain.GoImportRecord      `json:"intra_module_imports"`
	ExternalImports    []domain.GoImportRecord      `json:"external_imports"`
	StdlibImports      []domain.GoImportRecord      `json:"stdlib_imports"`
	Calls              []domain.GoCallRecord        `json:"calls"`
	KnownUnknowns      []domain.UnsupportedBoundary `json:"known_unknowns"`
}

// CanonicalJSON serializes the doc deterministically (sorted records,
// sorted keys) so md and json renderers share the same facts.
func (d *IntraServiceDoc) CanonicalJSON() ([]byte, error) {
	c := *d
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decodeIntraDetail reads the resolve-step detail off a Service node's
// metadata, whether it is still the struct (in-memory map) or plain data
// (after a macro-map.json round trip).
func decodeIntraDetail(meta map[string]any) *domain.IntraServiceDetail {
	raw, ok := meta[domain.IntraServiceMetadataKey]
	if !ok {
		return nil
	}
	switch d := raw.(type) {
	case domain.IntraServiceDetail:
		c := d
		return &c
	case *domain.IntraServiceDetail:
		return d
	case map[string]any:
		re, err := json.Marshal(d)
		if err != nil {
			return nil
		}
		var out domain.IntraServiceDetail
		if err := json.Unmarshal(re, &out); err != nil {
			return nil
		}
		if err := out.Validate(); err != nil {
			return nil
		}
		return &out
	}
	return nil
}

// GetIntraServiceDoc returns the intra-service view of one service:
// entrypoints, symbols, provided/consumed contracts (from the graph),
// imports split intra-module vs external vs stdlib, calls, and known
// unknowns. Services without Go detail (other languages, manifests only)
// still get contracts and unknowns; only an unknown service ID errors.
func (e *Engine) GetIntraServiceDoc(serviceID string) (*IntraServiceDoc, error) {
	node, ok := e.nodesByID[serviceID]
	if !ok {
		return nil, fmt.Errorf("service not found: %s", serviceID)
	}

	doc := &IntraServiceDoc{
		ServiceID:   node.ID,
		ServiceName: node.Name,
	}
	if len(node.Evidence) > 0 {
		doc.ServiceLocator = node.Evidence[0].Locator
	}

	if detail := decodeIntraDetail(node.Metadata); detail != nil {
		doc.Entrypoints = append([]domain.GoEntrypoint(nil), detail.Entrypoints...)
		doc.Packages = append([]domain.GoPackageSymbol(nil), detail.Packages...)
		doc.Types = append([]domain.GoTypeSymbol(nil), detail.Types...)
		doc.Funcs = append([]domain.GoFuncSymbol(nil), detail.Funcs...)
		doc.Calls = append([]domain.GoCallRecord(nil), detail.Calls...)
		for _, im := range detail.Imports {
			switch im.Scope {
			case domain.GoImportIntraModule:
				doc.IntraModuleImports = append(doc.IntraModuleImports, im)
			case domain.GoImportStdlib:
				doc.StdlibImports = append(doc.StdlibImports, im)
			default:
				doc.ExternalImports = append(doc.ExternalImports, im)
			}
		}
	}

	doc.ProvidedContracts = e.contractRefs(serviceID, true)
	doc.ConsumedContracts = e.contractRefs(serviceID, false)

	seenUB := map[string]bool{}
	for _, ub := range e.allUnsupported {
		if e.isUnsupportedRelated(ub, node) && !seenUB[ub.Path] {
			seenUB[ub.Path] = true
			doc.KnownUnknowns = append(doc.KnownUnknowns, ub)
		}
	}
	sort.Slice(doc.KnownUnknowns, func(i, j int) bool { return doc.KnownUnknowns[i].Path < doc.KnownUnknowns[j].Path })

	return doc, nil
}

// contractRefs lists the contracts serviceID provides (provided=true) or
// consumes, each with its proving edge and every evidence locator.
func (e *Engine) contractRefs(serviceID string, provided bool) []IntraContractRef {
	var rels []domain.Relationship
	if provided {
		rels = e.providesByService[serviceID]
	} else {
		rels = e.consumesByService[serviceID]
	}
	var out []IntraContractRef
	for _, rel := range rels {
		contractID := rel.TargetID
		if contractID == serviceID {
			contractID = rel.SourceID
		}
		contract, found := e.nodesByID[contractID]
		if !found || contract.Kind != domain.NodeContract {
			continue
		}
		ref := IntraContractRef{
			ContractID:   contract.ID,
			ContractName: contract.Name,
			EdgeID:       rel.ID,
		}
		if provided {
			ref.Verb = string(domain.VerbProvides)
		} else {
			ref.Verb = string(domain.VerbConsumes)
		}
		for _, ev := range rel.Evidence {
			ref.Locators = append(ref.Locators, ev.Locator)
		}
		out = append(out, ref)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ContractID < out[j].ContractID })
	return out
}
