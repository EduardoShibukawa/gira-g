package tools

// Tool represents a tool definition with its properties.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enveloped   bool                   `json:"enveloped"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Registry is the single source of truth for all tool definitions.
// Both MCP server and tokens package must use this registry.
var Registry = []Tool{
	{
		Name:        "gira_ask",
		Description: "Query codebase architecture, system purpose, components, entrypoints, and blast radius in natural language. Powered by deterministic AST graph queries with local JEV routing. Supports format: caveman (ultra-compact for agent token budgets), human, mermaid, or json.",
		Enveloped:   false,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Natural language question or architectural intent (e.g. 'Qual o objetivo do projeto?', 'O que quebra se alterar billing?').",
				},
				"diagram": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to include a Mermaid architecture diagram in the response.",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Output format: 'caveman' (dense facts for AI token budget), 'human' (plain text), 'mermaid' (diagram only), or 'json'. Default is 'caveman'.",
					"enum":        []string{"caveman", "human", "mermaid", "json"},
				},
			},
			"required": []string{"query"},
		},
	},
	{
		Name:        "get_impact_briefing",
		Description: "Analyzes upstream services, downstream consumers, mediating contracts, and unsupported boundaries affected by a change to target_id.",
		Enveloped:   true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"target_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the target node (service, contract, or infrastructure) to analyze.",
				},
				"depth": map[string]interface{}{
					"type":        "integer",
					"description": "Multi-hop graph traversal depth (default 1).",
				},
				"budget": map[string]interface{}{
					"type":        "integer",
					"description": "Optional hard ceiling on estimated output tokens. The output then carries a budget object {limit, estimated, truncated, dropped, escalate_hint}.",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Output format: 'json' (default), 'caveman' (dense one-fact-per-line for AI agents, saving 85% tokens), or 'md'.",
					"enum":        []string{"json", "caveman", "md"},
				},
			},
			"required": []string{"target_id"},
		},
	},
	{
		Name:        "get_pr_impact",
		Description: "Impact briefing for a change set: maps the changed line ranges of a git revision range (or a supplied unified diff) to the nodes whose evidence they touch, and merges their impact briefings. Changed files no locator cites are listed as unmatched.",
		Enveloped:   true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"range": map[string]interface{}{
					"type":        "string",
					"description": "Git revision range <base>..<head>, read from repo_path. Labels the result when diff is given.",
				},
				"diff": map[string]interface{}{
					"type":        "string",
					"description": "Optional unified diff text (for example `git diff -U0 base..head`); used instead of running git.",
				},
				"repo_path": map[string]interface{}{
					"type":        "string",
					"description": "Repository the range is read from (default: the server's working directory).",
				},
				"repo_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional repository id; only locators of this repository are matched.",
				},
				"depth": map[string]interface{}{
					"type":        "integer",
					"description": "Multi-hop graph traversal depth (default 1).",
				},
				"budget": map[string]interface{}{
					"type":        "integer",
					"description": "Optional hard ceiling on estimated output tokens. The output then carries a budget object {limit, estimated, truncated, dropped, escalate_hint}.",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Output format: 'json' (default), 'caveman' (dense one-fact-per-line for AI agents, saving 85% tokens), or 'md'.",
					"enum":        []string{"json", "caveman", "md"},
				},
			},
			"required": []string{"range"},
		},
	},
	{
		Name:        "get_tour",
		Description: "Deterministic guided reading tour: from entry points (services with declared ports) or a focus node, walks consumes/provides/accesses edges breadth-first. Each step has order, title, node_ids, read (locators) and a templated why.",
		Enveloped:   true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"focus": map[string]interface{}{
					"type":        "string",
					"description": "Optional node to start the tour at.",
				},
				"max_steps": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of steps (default 20).",
				},
				"budget": map[string]interface{}{
					"type":        "integer",
					"description": "Optional hard ceiling on estimated output tokens. The output then carries a budget object {limit, estimated, truncated, dropped, escalate_hint}.",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Output format: 'json' (default), 'caveman' (dense one-fact-per-line for AI agents, saving 85% tokens), or 'md'.",
					"enum":        []string{"json", "caveman", "md"},
				},
			},
		},
	},
	{
		Name:        "get_doc",
		Description: "Single-service onboarding doc: entrypoints, symbols, provided/consumed contracts, imports, calls, run/test manifests, change dirs, tour, flows and known unknowns for a service. Same engine as `gira doc --format json`.",
		Enveloped:   true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the service to document (for example service:acme).",
				},
				"budget": map[string]interface{}{
					"type":        "integer",
					"description": "Optional hard ceiling on estimated output tokens. The output then carries a budget object {limit, estimated, truncated, dropped, escalate_hint}.",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Output format: 'json' (default), 'caveman' (dense one-fact-per-line for AI agents, saving 85% tokens), or 'md'.",
					"enum":        []string{"json", "caveman", "md"},
				},
			},
			"required": []string{"service_id"},
		},
	},
	{
		Name:        "diff_maps",
		Description: "Architecture diff between two index files: renamed, added and removed nodes, added and removed edges, changed contracts, newly unsupported paths, and (informational) moved evidence. Renames are detected from structure.",
		Enveloped:   true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"index_a": map[string]interface{}{
					"type":        "string",
					"description": "Path to the earlier macro-map.json.",
				},
				"index_b": map[string]interface{}{
					"type":        "string",
					"description": "Path to the later macro-map.json.",
				},
				"budget": map[string]interface{}{
					"type":        "integer",
					"description": "Optional hard ceiling on estimated output tokens. The output then carries a budget object {limit, estimated, truncated, dropped, escalate_hint}.",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Output format: 'json' (default), 'caveman' (dense one-fact-per-line for AI agents, saving 85% tokens), or 'md'.",
					"enum":        []string{"json", "caveman", "md"},
				},
			},
			"required": []string{"index_a", "index_b"},
		},
	},
	{
		Name:        "get_context_slice",
		Description: "Retrieves a bounded context slice around node_id across summary, contracts, or deep disclosure tiers.",
		Enveloped:   true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"node_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the focal node to extract context slice for.",
				},
				"tier": map[string]interface{}{
					"type":        "string",
					"description": "Context slice disclosure tier: 'summary', 'contracts', or 'deep' (default 'summary').",
					"enum":        []string{"summary", "contracts", "deep"},
				},
				"view": map[string]interface{}{
					"type":        "string",
					"description": "Context slice architectural view: 'full', 'architecture', or 'governance' (default 'full').",
					"enum":        []string{"full", "architecture", "governance"},
				},
				"budget": map[string]interface{}{
					"type":        "integer",
					"description": "Optional hard ceiling on estimated output tokens. The output then carries a budget object {limit, estimated, truncated, dropped, escalate_hint}.",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Output format: 'json' (default), 'caveman' (dense one-fact-per-line for AI agents, saving 85% tokens), or 'md'.",
					"enum":        []string{"json", "caveman", "md"},
				},
			},
			"required": []string{"node_id"},
		},
	},
	{
		Name:        "explain_evidence",
		Description: "Inspects provenance, derivation tier, deployment profile, and freshness status for a node and/or edge.",
		Enveloped:   true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"node_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional node ID to explain evidence records for.",
				},
				"edge_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional edge / relationship ID to explain evidence records for.",
				},
				"budget": map[string]interface{}{
					"type":        "integer",
					"description": "Optional hard ceiling on estimated output tokens. The output then carries a budget object {limit, estimated, truncated, dropped, escalate_hint}.",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Output format: 'json' (default), 'caveman' (dense one-fact-per-line for AI agents, saving 85% tokens), or 'md'.",
					"enum":        []string{"json", "caveman", "md"},
				},
			},
		},
	},
	{
		Name:        "locate_boundary",
		Description: "Finds architectural boundary components, contract RPC methods, or unhandled/unsupported files matching query.",
		Enveloped:   true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Substring query to search across node IDs, names, contract methods, and unsupported files.",
				},
				"budget": map[string]interface{}{
					"type":        "integer",
					"description": "Optional hard ceiling on estimated output tokens. The output then carries a budget object {limit, estimated, truncated, dropped, escalate_hint}.",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Output format: 'json' (default), 'caveman' (dense one-fact-per-line for AI agents, saving 85% tokens), or 'md'.",
					"enum":        []string{"json", "caveman", "md"},
				},
			},
			"required": []string{"query"},
		},
	},
	{
		Name:        "verify_artifact",
		Description: "F1: recomputes the L1 artifact digest for CHECK/EXPECT/CWD+exit+fingerprint and returns {digest, closing_line}.",
		Enveloped:   true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"check": map[string]interface{}{
					"type":        "string",
					"description": "CHECK: command label recorded in the digest.",
				},
				"expect": map[string]interface{}{
					"type":        "string",
					"description": "EXPECT: success marker recorded in the digest.",
				},
				"cwd": map[string]interface{}{
					"type":        "string",
					"description": "CWD recorded in the digest.",
				},
				"exit_code": map[string]interface{}{
					"type":        "integer",
					"description": "Exit code recorded in the digest.",
				},
				"fingerprint": map[string]interface{}{
					"type":        "string",
					"description": "Frozen input fingerprint hex.",
				},
				"taxonomy": map[string]interface{}{
					"type":        "string",
					"description": "L3 claim taxonomy recorded in the digest (measured|checked|inferred|assumed, default inferred).",
				},
			},
		},
	},
	{
		Name:        "get_control_total",
		Description: "F2: validates control:{total,matched,mismatched,sampled_n,seed} against the seeded max(20,5%) rule. Format validation only: it rejects inconsistent lines but does not derive a measurement (derivation lives in gira eval --control-total).",
		Enveloped:   true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"total": map[string]interface{}{
					"type":        "integer",
					"description": "Control denominator (task count).",
				},
				"matched": map[string]interface{}{
					"type":        "integer",
					"description": "Blind re-derivation matches on the sample.",
				},
				"mismatched": map[string]interface{}{
					"type":        "integer",
					"description": "Blind re-derivation mismatches on the sample.",
				},
				"seed": map[string]interface{}{
					"type":        "integer",
					"description": "Sampler seed.",
				},
			},
		},
	},
	{
		Name:        "get_findings_register",
		Description: "F3: returns the L3 findings register entries (pointer only, outside slices).",
		Enveloped:   true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"register": map[string]interface{}{
					"type":        "string",
					"description": "Path to the findings register file.",
				},
			},
		},
	},
	{
		Name:        "get_token_accounting",
		Description: "F4: returns honest token accounting {gross,expansions,net,measured_bool} (measurement only).",
		Enveloped:   true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text": map[string]interface{}{
					"type":        "string",
					"description": "Text to measure (lifted before measuring).",
				},
				"expansions": map[string]interface{}{
					"type":        "integer",
					"description": "Expansion debit.",
				},
			},
		},
	},
	{
		Name:        "gate_status",
		Description: "F5: non-executing gate status read for a ledger file.",
		Enveloped:   true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"state": map[string]interface{}{
					"type":        "string",
					"description": "Gate ledger file path.",
				},
			},
		},
	},
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
