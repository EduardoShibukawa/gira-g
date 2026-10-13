package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/EduardoShibukawa/gira-g/internal/archdiff"
	"github.com/EduardoShibukawa/gira-g/internal/gitdiff"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
	"github.com/EduardoShibukawa/gira-g/internal/store"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request or notification.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error object.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

const (
	ErrParse          = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternal       = -32603
)

// MCP Tool representations
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]PropertyDef `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

type PropertyDef struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type CallToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Server handles MCP protocol communication over stdio or arbitrary streams.
type Server struct {
	Name    string
	Version string
	Engine  *slices.Engine
	mu      sync.Mutex
}

// NewServer creates a new MCP server instance.
func NewServer(name, version string, engine *slices.Engine) *Server {
	return &Server{
		Name:    name,
		Version: version,
		Engine:  engine,
	}
}

// ServeStdio starts serving JSON-RPC requests over os.Stdin and os.Stdout.
func (s *Server) ServeStdio() error {
	return s.Serve(os.Stdin, os.Stdout)
}

// Serve reads JSON-RPC 2.0 messages from in and writes responses to out.
// Supports both newline-delimited JSON and Content-Length framed JSON.
func (s *Server) Serve(in io.Reader, out io.Writer) error {
	reader := bufio.NewReader(in)

	for {
		payload, err := readMessage(reader)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if len(bytes.TrimSpace(payload)) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			resp := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error: &JSONRPCError{
					Code:    ErrParse,
					Message: fmt.Sprintf("Parse error: %v", err),
				},
			}
			if writeErr := s.writeResponse(out, resp); writeErr != nil {
				return writeErr
			}
			continue
		}

		// Handle request
		resp, isNotification := s.handleRequest(&req)
		if !isNotification && resp != nil {
			if err := s.writeResponse(out, resp); err != nil {
				return err
			}
		}
	}
}

func readMessage(r *bufio.Reader) ([]byte, error) {
	for {
		firstByte, err := r.Peek(1)
		if err != nil {
			return nil, err
		}

		// Skip whitespace or empty lines
		if firstByte[0] == '\r' || firstByte[0] == '\n' || firstByte[0] == ' ' || firstByte[0] == '\t' {
			_, _ = r.ReadByte()
			continue
		}

		// Check if it's Content-Length header or direct JSON '{'
		if firstByte[0] == '{' {
			// Newline delimited JSON
			line, err := r.ReadBytes('\n')
			if err != nil && len(line) == 0 {
				return nil, err
			}
			return bytes.TrimSpace(line), nil
		}

		// Otherwise check header framing: "Content-Length: <n>\r\n"
		headerLine, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		headerLine = strings.TrimSpace(headerLine)
		if strings.HasPrefix(strings.ToLower(headerLine), "content-length:") {
			parts := strings.SplitN(headerLine, ":", 2)
			if len(parts) == 2 {
				lengthStr := strings.TrimSpace(parts[1])
				length, convErr := strconv.Atoi(lengthStr)
				if convErr == nil {
					// Read until empty line (\r\n or \n)
					for {
						sepLine, err := r.ReadString('\n')
						if err != nil {
							return nil, err
						}
						if strings.TrimSpace(sepLine) == "" {
							break
						}
					}
					// Read exact length payload
					buf := make([]byte, length)
					_, err := io.ReadFull(r, buf)
					if err != nil {
						return nil, err
					}
					return buf, nil
				}
			}
		}
	}
}

func (s *Server) writeResponse(out io.Writer, resp any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	_, err = out.Write(data)
	return err
}

func (s *Server) handleRequest(req *JSONRPCRequest) (*JSONRPCResponse, bool) {
	switch req.Method {
	case "initialize":
		result := map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    s.Name,
				"version": s.Version,
			},
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
		}
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		}, false

	case "notifications/initialized":
		return nil, true

	case "ping":
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]any{},
		}, false

	case "tools/list":
		tools := s.getToolDefinitions()
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"tools": tools,
			},
		}, false

	case "tools/call":
		res := s.handleToolCall(req.Params)
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  res,
		}, false

	default:
		// Unknown method
		if req.ID == nil {
			return nil, true // Notification with unknown method
		}
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    ErrMethodNotFound,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}, false
	}
}

func (s *Server) getToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "get_impact_briefing",
			Description: "Analyzes upstream services, downstream consumers, mediating contracts, and unsupported boundaries affected by a change to target_id.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"target_id": {
						Type:        "string",
						Description: "ID of the target node (service, contract, or infrastructure) to analyze.",
					},
					"depth": {
						Type:        "integer",
						Description: "Multi-hop graph traversal depth (default 1).",
					},
					"budget": {
						Type:        "integer",
						Description: "Optional hard ceiling on estimated output tokens. The output then carries a budget object {limit, estimated, truncated, dropped, escalate_hint}.",
					},
				},
				Required: []string{"target_id"},
			},
		},
		{
			Name:        "get_pr_impact",
			Description: "Impact briefing for a change set: maps the changed line ranges of a git revision range (or a supplied unified diff) to the nodes whose evidence they touch, and merges their impact briefings. Changed files no locator cites are listed as unmatched.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"range": {
						Type:        "string",
						Description: "Git revision range <base>..<head>, read from repo_path. Labels the result when diff is given.",
					},
					"diff": {
						Type:        "string",
						Description: "Optional unified diff text (for example `git diff -U0 base..head`); used instead of running git.",
					},
					"repo_path": {
						Type:        "string",
						Description: "Repository the range is read from (default: the server's working directory).",
					},
					"repo_id": {
						Type:        "string",
						Description: "Optional repository id; only locators of this repository are matched.",
					},
					"depth": {
						Type:        "integer",
						Description: "Multi-hop graph traversal depth (default 1).",
					},
					"budget": {
						Type:        "integer",
						Description: "Optional hard ceiling on estimated output tokens. The output then carries a budget object {limit, estimated, truncated, dropped, escalate_hint}.",
					},
				},
				Required: []string{"range"},
			},
		},
		{
			Name:        "get_tour",
			Description: "Deterministic guided reading tour: from entry points (services with declared ports) or a focus node, walks consumes/provides/accesses edges breadth-first. Each step has order, title, node_ids, read (locators) and a templated why.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"focus": {
						Type:        "string",
						Description: "Optional node to start the tour at.",
					},
					"max_steps": {
						Type:        "integer",
						Description: "Maximum number of steps (default 20).",
					},
				},
			},
		},
		{
			Name:        "diff_maps",
			Description: "Architecture diff between two index files: renamed, added and removed nodes, added and removed edges, changed contracts, newly unsupported paths, and (informational) moved evidence. Renames are detected from structure.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"index_a": {
						Type:        "string",
						Description: "Path to the earlier macro-map.json.",
					},
					"index_b": {
						Type:        "string",
						Description: "Path to the later macro-map.json.",
					},
				},
				Required: []string{"index_a", "index_b"},
			},
		},
		{
			Name:        "get_context_slice",
			Description: "Retrieves a bounded context slice around node_id across summary, contracts, or deep disclosure tiers.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"node_id": {
						Type:        "string",
						Description: "ID of the focal node to extract context slice for.",
					},
					"tier": {
						Type:        "string",
						Description: "Context slice disclosure tier: 'summary', 'contracts', or 'deep' (default 'summary').",
						Enum:        []string{"summary", "contracts", "deep"},
					},
					"view": {
						Type:        "string",
						Description: "Context slice architectural view: 'full', 'architecture', or 'governance' (default 'full').",
						Enum:        []string{"full", "architecture", "governance"},
					},
					"budget": {
						Type:        "integer",
						Description: "Optional hard ceiling on estimated output tokens. The output then carries a budget object {limit, estimated, truncated, dropped, escalate_hint}.",
					},
				},
				Required: []string{"node_id"},
			},
		},
		{
			Name:        "explain_evidence",
			Description: "Inspects provenance, derivation tier, deployment profile, and freshness status for a node and/or edge.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"node_id": {
						Type:        "string",
						Description: "Optional node ID to explain evidence records for.",
					},
					"edge_id": {
						Type:        "string",
						Description: "Optional edge / relationship ID to explain evidence records for.",
					},
				},
			},
		},
		{
			Name:        "locate_boundary",
			Description: "Finds architectural boundary components, contract RPC methods, or unhandled/unsupported files matching query.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"query": {
						Type:        "string",
						Description: "Substring query to search across node IDs, names, contract methods, and unsupported files.",
					},
				},
				Required: []string{"query"},
			},
		},
	}
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (s *Server) handleToolCall(paramsRaw json.RawMessage) CallToolResult {
	if s.Engine == nil {
		return CallToolResult{
			IsError: true,
			Content: []ToolContent{
				{Type: "text", Text: "Engine not initialized"},
			},
		}
	}

	var params toolCallParams
	if err := json.Unmarshal(paramsRaw, &params); err != nil {
		return CallToolResult{
			IsError: true,
			Content: []ToolContent{
				{Type: "text", Text: fmt.Sprintf("Invalid params: %v", err)},
			},
		}
	}

	switch params.Name {
	case "get_impact_briefing":
		var args struct {
			TargetID string `json:"target_id"`
			Depth    int    `json:"depth"`
			Budget   int    `json:"budget"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Invalid arguments for get_impact_briefing: %v", err)},
				},
			}
		}
		if args.TargetID == "" {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: "Missing required argument 'target_id'"},
				},
			}
		}

		briefing, err := s.Engine.GetImpactBriefing(args.TargetID, args.Depth)
		if err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Failed to get impact briefing: %v", err)},
				},
			}
		}

		outJSON, err := briefing.Emit(slices.FormatJSON, args.Budget)
		if err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Failed to render impact briefing: %v", err)},
				},
			}
		}

		return CallToolResult{
			Content: []ToolContent{
				{Type: "text", Text: string(outJSON)},
			},
		}

	case "get_pr_impact":
		var args struct {
			Range    string `json:"range"`
			Diff     string `json:"diff"`
			RepoPath string `json:"repo_path"`
			RepoID   string `json:"repo_id"`
			Depth    int    `json:"depth"`
			Budget   int    `json:"budget"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return errorResult(fmt.Sprintf("Invalid arguments for get_pr_impact: %v", err))
		}
		if args.Range == "" {
			return errorResult("Missing required argument 'range'")
		}
		diff := args.Diff
		if diff == "" {
			dir := args.RepoPath
			if dir == "" {
				dir = "."
			}
			out, err := gitdiff.Run(context.Background(), dir, args.Range)
			if err != nil {
				return errorResult(fmt.Sprintf("Failed to read diff: %v", err))
			}
			diff = out
		}
		changes, err := gitdiff.Parse(diff)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to parse diff: %v", err))
		}
		briefing, err := s.Engine.GetPRImpact(args.Range, args.RepoID, changes, args.Depth)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get PR impact: %v", err))
		}
		out, err := briefing.Emit(slices.FormatJSON, args.Budget)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to render PR impact: %v", err))
		}
		return CallToolResult{Content: []ToolContent{{Type: "text", Text: string(out)}}}

	case "get_tour":
		var args struct {
			Focus    string `json:"focus"`
			MaxSteps int    `json:"max_steps"`
		}
		if len(params.Arguments) > 0 {
			if err := json.Unmarshal(params.Arguments, &args); err != nil {
				return errorResult(fmt.Sprintf("Invalid arguments for get_tour: %v", err))
			}
		}
		tour, err := s.Engine.GetTour(args.Focus, args.MaxSteps)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to get tour: %v", err))
		}
		out, err := tour.CanonicalJSON()
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to render tour: %v", err))
		}
		return CallToolResult{Content: []ToolContent{{Type: "text", Text: string(out)}}}

	case "diff_maps":
		var args struct {
			IndexA string `json:"index_a"`
			IndexB string `json:"index_b"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return errorResult(fmt.Sprintf("Invalid arguments for diff_maps: %v", err))
		}
		if args.IndexA == "" || args.IndexB == "" {
			return errorResult("Missing required arguments 'index_a' and 'index_b'")
		}
		st := store.NewStore("")
		a, err := st.LoadFile(args.IndexA)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to load index_a: %v", err))
		}
		b, err := st.LoadFile(args.IndexB)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to load index_b: %v", err))
		}
		out, err := archdiff.Compare(a, b).CanonicalJSON()
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to render diff: %v", err))
		}
		return CallToolResult{Content: []ToolContent{{Type: "text", Text: string(out)}}}

	case "get_context_slice":
		var args struct {
			NodeID string `json:"node_id"`
			Tier   string `json:"tier"`
			View   string `json:"view"`
			Budget int    `json:"budget"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Invalid arguments for get_context_slice: %v", err)},
				},
			}
		}
		if args.NodeID == "" {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: "Missing required argument 'node_id'"},
				},
			}
		}

		tier := slices.SliceTier(args.Tier)
		if tier == "" {
			tier = slices.TierSummary
		}

		view := slices.GraphView(args.View)
		if view == "" {
			view = slices.ViewFull
		}

		slice, err := s.Engine.GetContextSliceWithView(args.NodeID, tier, view)
		if err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Failed to get context slice: %v", err)},
				},
			}
		}

		outJSON, err := slice.Emit(slices.FormatJSON, args.Budget)
		if err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Failed to render context slice: %v", err)},
				},
			}
		}

		return CallToolResult{
			Content: []ToolContent{
				{Type: "text", Text: string(outJSON)},
			},
		}

	case "explain_evidence":
		var args struct {
			NodeID string `json:"node_id"`
			EdgeID string `json:"edge_id"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Invalid arguments for explain_evidence: %v", err)},
				},
			}
		}
		if args.NodeID == "" && args.EdgeID == "" {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: "Either node_id or edge_id must be provided"},
				},
			}
		}

		report, err := s.Engine.ExplainEvidence(args.NodeID, args.EdgeID)
		if err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Failed to explain evidence: %v", err)},
				},
			}
		}

		outJSON, err := report.CanonicalJSON()
		if err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Failed to serialize canonical JSON: %v", err)},
				},
			}
		}

		return CallToolResult{
			Content: []ToolContent{
				{Type: "text", Text: string(outJSON)},
			},
		}

	case "locate_boundary":
		var args struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Invalid arguments for locate_boundary: %v", err)},
				},
			}
		}
		if strings.TrimSpace(args.Query) == "" {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: "Missing required argument 'query'"},
				},
			}
		}

		matches, err := s.Engine.LocateBoundary(args.Query)
		if err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Failed to locate boundary: %v", err)},
				},
			}
		}

		data, err := json.MarshalIndent(matches, "", "  ")
		if err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Failed to format matches: %v", err)},
				},
			}
		}

		return CallToolResult{
			Content: []ToolContent{
				{Type: "text", Text: string(data)},
			},
		}

	default:
		return CallToolResult{
			IsError: true,
			Content: []ToolContent{
				{Type: "text", Text: fmt.Sprintf("Unknown tool: %s", params.Name)},
			},
		}
	}
}

func errorResult(msg string) CallToolResult {
	return CallToolResult{IsError: true, Content: []ToolContent{{Type: "text", Text: msg}}}
}
