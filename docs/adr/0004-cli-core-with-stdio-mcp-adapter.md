# CLI and Go library core with stdio MCP adapter

The harness core is implemented as a pure Go library and CLI tool; MCP is an external adapter exposing a subset of question-shaped operations over stdio. All operations within the harness remain 100% deterministic with zero model inference (reinforcing ADR 0001). Coding agents integrate either through the stdio MCP adapter (`harness serve mcp`) or directly via CLI invocations (`harness impact`, `harness slice`) guided by portable `AGENTS.md` instructions and agent skills. The MCP implementation uses the official `modelcontextprotocol/go-sdk` behind an internal interface.

## Considered Options

- **MCP-first architecture where the harness only exists as an MCP server.** Rejected: makes CLI scripting, CI evaluation benchmarks, and integration with shell-only agents or test runners unnecessarily complex and dependent on an MCP client harness.
- **`mark3labs/mcp-go` community SDK.** Rejected in favor of official `modelcontextprotocol/go-sdk`: the official SDK is Tier 1, directly tracks protocol revisions, and avoids third-party divergence. The SDK is wrapped behind an internal seam so that changing the transport or library remains an isolated internal detail.
- **In-harness LLM summarization or inference via MCP sampling.** Rejected: ADR 0001 establishes that the MVP ships no model client. Furthermore, the MCP protocol deprecates sampling in recent revisions; any model reasoning belongs strictly to the consuming agent.
- **Host-specific hooks or automatic instruction injection (e.g. Graphify/Understand Anything style).** Rejected: hooks are vendor-specific and fragile across different agents (Ticket 19); portable `AGENTS.md` and `SKILL.md` files provide universal discovery without polluting developer environments.

Source: [Agent integration surface](../../../../.scratch/system-context-harness/issues/14-agent-integration-surface.md)
