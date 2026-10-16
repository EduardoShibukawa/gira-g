package tools

// Tool represents a tool definition with its properties.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enveloped   bool   `json:"enveloped"`
}

// Registry is the single source of truth for all tool definitions.
// Both MCP server and tokens package must use this registry.
var Registry = []Tool{
	{Name: "get_impact_briefing", Description: "Get impact briefing for a target", Enveloped: true},
	{Name: "get_pr_impact", Description: "Get PR impact analysis", Enveloped: true},
	{Name: "get_tour", Description: "Deterministic guided reading tour", Enveloped: true},
	{Name: "get_doc", Description: "Single-service onboarding doc, same engine as doc --format json", Enveloped: true},
	{Name: "diff_maps", Description: "Architecture diff between two index files", Enveloped: true},
	{Name: "get_context_slice", Description: "Retrieves a bounded context slice", Enveloped: true},
	{Name: "explain_evidence", Description: "Inspects provenance and evidence", Enveloped: true},
	{Name: "locate_boundary", Description: "Finds architectural boundaries", Enveloped: true},
	{Name: "verify_artifact", Description: "Recomputes L1 artifact digest", Enveloped: true},
	{Name: "get_control_total", Description: "Validates control total", Enveloped: true},
	{Name: "get_findings_register", Description: "Returns findings register", Enveloped: true},
	{Name: "get_token_accounting", Description: "Returns token accounting", Enveloped: true},
	{Name: "gate_status", Description: "Returns gate status", Enveloped: true},
}

// Names returns all tool names.
func Names() []string {
	names := make([]string, len(Registry))
	for i, t := range Registry {
		names[i] = t.Name
	}
	return names
}

// ByName returns a tool by name, or nil if not found.
func ByName(name string) *Tool {
	for i := range Registry {
		if Registry[i].Name == name {
			return &Registry[i]
		}
	}
	return nil
}

// EnvelopedNames returns names of all enveloped tools.
func EnvelopedNames() []string {
	var names []string
	for _, t := range Registry {
		if t.Enveloped {
			names = append(names, t.Name)
		}
	}
	return names
}
