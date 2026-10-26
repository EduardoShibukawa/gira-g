package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/EduardoShibukawa/gira-g/internal/archdiff"
	"github.com/EduardoShibukawa/gira-g/internal/ask"
	"github.com/EduardoShibukawa/gira-g/internal/eval"
	gatepkg "github.com/EduardoShibukawa/gira-g/internal/gate"
	"github.com/EduardoShibukawa/gira-g/internal/gitdiff"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
	"github.com/EduardoShibukawa/gira-g/internal/store"
	"github.com/EduardoShibukawa/gira-g/internal/tokens"
	"github.com/EduardoShibukawa/gira-g/internal/tools"
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
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Enveloped   bool                   `json:"enveloped"`
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
	Type string         `json:"type"`
	Text string         `json:"text"`
	Data map[string]any `json:"data,omitempty"`
}

// StructuredError represents a structured error response for MCP tools.
type StructuredError struct {
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	Retryable   bool     `json:"retryable"`
	Suggestions []string `json:"suggestions,omitempty"`
}

func newStructuredError(code, message string, retryable bool, suggestions ...string) CallToolResult {
	err := StructuredError{
		Code:        code,
		Message:     message,
		Retryable:   retryable,
		Suggestions: suggestions,
	}
	data, _ := json.Marshal(err)
	return CallToolResult{IsError: true, Content: []ToolContent{{Type: "text", Text: string(data)}}}
}

func errorResult(msg string) CallToolResult {
	return CallToolResult{IsError: true, Content: []ToolContent{{Type: "text", Text: msg}}}
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
	defs := make([]ToolDefinition, 0, len(tools.Registry))
	for _, t := range tools.Registry {
		defs = append(defs, ToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
			Enveloped:   t.Enveloped,
		})
	}
	return defs
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (s *Server) handleToolCall(paramsRaw json.RawMessage) CallToolResult {
	var params toolCallParams
	if err := json.Unmarshal(paramsRaw, &params); err != nil {
		return CallToolResult{
			IsError: true,
			Content: []ToolContent{
				{Type: "text", Text: fmt.Sprintf("Invalid params: %v", err)},
			},
		}
	}

	// F1: verify_artifact needs no engine (pure digest recomputation).
	if params.Name == "verify_artifact" {
		var args struct {
			Check       string `json:"check"`
			Expect      string `json:"expect"`
			CWD         string `json:"cwd"`
			ExitCode    int    `json:"exit_code"`
			Fingerprint string `json:"fingerprint"`
			Taxonomy    string `json:"taxonomy"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Invalid arguments for verify_artifact: %v", err)},
				},
			}
		}
		if args.Taxonomy == "" {
			args.Taxonomy = "inferred"
		}
		if !eval.ValidTaxonomy(args.Taxonomy) {
			return CallToolResult{IsError: true, Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("unknown taxonomy %q", args.Taxonomy)}}}
		}
		digest := eval.ArtifactDigest(args.Check, args.Expect, args.CWD, args.ExitCode, args.Fingerprint, args.Taxonomy)
		out, _ := json.Marshal(map[string]string{"digest": digest, "closing_line": eval.FormatClosingLine(digest)})
		return CallToolResult{Content: []ToolContent{{Type: "text", Text: string(out)}}}
	}

	// F2: get_control_total needs no engine (pure control math).
	if params.Name == "get_control_total" {
		var args struct {
			Total      int    `json:"total"`
			Matched    int    `json:"matched"`
			Mismatched int    `json:"mismatched"`
			Seed       uint64 `json:"seed"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Invalid arguments for get_control_total: %v", err)},
				},
			}
		}
		rep := eval.ControlReport(args.Total, args.Matched, args.Mismatched, args.Seed)
		if err := rep.Validate(); err != nil {
			return CallToolResult{IsError: true, Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("inconsistent control line: %v", err)}}}
		}
		out, _ := json.Marshal(rep)
		return CallToolResult{Content: []ToolContent{{Type: "text", Text: string(out)}}}
	}

	// F3: get_findings_register reads the register file (no engine).
	if params.Name == "get_findings_register" {
		var args struct {
			Register string `json:"register"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Invalid arguments for get_findings_register: %v", err)},
				},
			}
		}
		path := args.Register
		if path == "" {
			path = "reviews/findings.md"
		}
		reg, err := eval.LoadRegister(path)
		if err != nil {
			return CallToolResult{IsError: true, Content: []ToolContent{{Type: "text", Text: err.Error()}}}
		}
		out, _ := json.Marshal(reg.Findings)
		return CallToolResult{Content: []ToolContent{{Type: "text", Text: string(out)}}}
	}

	// F4: get_token_accounting is measurement only (no engine).
	if params.Name == "get_token_accounting" {
		var args struct {
			Text       string `json:"text"`
			Expansions int    `json:"expansions"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Invalid arguments for get_token_accounting: %v", err)},
				},
			}
		}
		m := tokens.Measure([]byte(args.Text))
		// Caller-supplied expansions are unmeasured input: a mixed report
		// must not be labeled measured (quarantine). Pure measurements stay
		// measured.
		rep := tokens.NewReport(m.Gross, args.Expansions, args.Expansions == 0)
		out, _ := json.Marshal(rep)
		return CallToolResult{Content: []ToolContent{{Type: "text", Text: string(out)}}}
	}

	// F5: gate_status is a non-executing read (no engine).
	if params.Name == "gate_status" {
		var args struct {
			State string `json:"state"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return CallToolResult{
				IsError: true,
				Content: []ToolContent{
					{Type: "text", Text: fmt.Sprintf("Invalid arguments for gate_status: %v", err)},
				},
			}
		}
		path := args.State
		if path == "" {
			path = ".gira/gate.json"
		}
		s, err := gatepkg.Status(path)
		if err != nil {
			return CallToolResult{IsError: true, Content: []ToolContent{{Type: "text", Text: err.Error()}}}
		}
		out, _ := json.Marshal(s)
		return CallToolResult{Content: []ToolContent{{Type: "text", Text: string(out)}}}
	}

	if s.Engine == nil {
		return CallToolResult{
			IsError: true,
			Content: []ToolContent{
				{Type: "text", Text: "Engine not initialized"},
			},
		}
	}

	switch params.Name {
	case "gira_ask":
		var args struct {
			Query   string `json:"query"`
			Diagram bool   `json:"diagram"`
			Format  string `json:"format"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return newStructuredError("invalid_query", fmt.Sprintf("Invalid arguments for gira_ask: %v", err), false, "Check arguments format")
		}
		if strings.TrimSpace(args.Query) == "" {
			return newStructuredError("invalid_query", "Missing required argument 'query'", false, "Provide natural language query")
		}

		resp := ask.ProcessQuery(args.Query, s.Engine.MacroMap(), args.Diagram)

		fmtChoice := strings.ToLower(strings.TrimSpace(args.Format))
		var outputText string
		switch fmtChoice {
		case "human", "md", "markdown":
			outputText = resp.RenderMarkdown()
		case "mermaid":
			if resp.Diagram != "" {
				outputText = resp.Diagram
			} else {
				outputText = "graph TD\n  None[No diagram generated]"
			}
		case "json":
			data, _ := json.MarshalIndent(resp, "", "  ")
			outputText = string(data)
		default: // "caveman" is the default for agent context
			outputText = resp.RenderCaveman()
		}

		return CallToolResult{
			Content: []ToolContent{
				{Type: "text", Text: outputText},
			},
		}

	case "get_impact_briefing":
		var args struct {
			TargetID string `json:"target_id"`
			Depth    int    `json:"depth"`
			Budget   int    `json:"budget"`
			Format   string `json:"format"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return newStructuredError("invalid_budget", fmt.Sprintf("Invalid arguments for get_impact_briefing: %v", err), false, "Check arguments format")
		}
		if args.TargetID == "" {
			return newStructuredError("invalid_budget", "Missing required argument 'target_id'", false, "Provide target_id")
		}

		briefing, err := s.Engine.GetImpactBriefing(args.TargetID, args.Depth)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return newStructuredError("node_not_found", fmt.Sprintf("Failed to get impact briefing: %v", err), true, "Use locate_boundary to search for valid node IDs")
			}
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to get impact briefing: %v", err), true, "Check target_id")
		}

		reqFmt := slices.FormatJSON
		if strings.ToLower(args.Format) == "caveman" || strings.ToLower(args.Format) == "llm" {
			reqFmt = slices.FormatCaveman
		} else if strings.ToLower(args.Format) == "md" || strings.ToLower(args.Format) == "markdown" {
			reqFmt = slices.FormatMarkdown
		}

		outJSON, err := briefing.Emit(reqFmt, args.Budget)
		if err != nil {
			if errors.Is(err, slices.ErrBudgetExceeded) {
				return newStructuredError("budget_exceeded", err.Error(), false, "rerun with a larger budget")
			}
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to render impact briefing: %v", err), false, "Check budget parameter")
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
			return newStructuredError("invalid_budget", fmt.Sprintf("Invalid arguments for get_pr_impact: %v", err), false, "Check arguments format")
		}
		if args.Range == "" {
			return newStructuredError("invalid_budget", "Missing required argument 'range'", false, "Provide a git revision range")
		}
		diff := args.Diff
		if diff == "" {
			dir := args.RepoPath
			if dir == "" {
				dir = "."
			}
			out, err := gitdiff.Run(context.Background(), dir, args.Range)
			if err != nil {
				return newStructuredError("envelope_error", fmt.Sprintf("Failed to read diff: %v", err), true, "Check repo_path and range")
			}
			diff = out
		}
		changes, err := gitdiff.Parse(diff)
		if err != nil {
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to parse diff: %v", err), false, "Provide a valid unified diff")
		}
		briefing, err := s.Engine.GetPRImpact(args.Range, args.RepoID, changes, args.Depth)
		if err != nil {
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to get PR impact: %v", err), true, "Check diff and range")
		}
		out, err := briefing.Emit(slices.FormatJSON, args.Budget)
		if err != nil {
			if errors.Is(err, slices.ErrBudgetExceeded) {
				return newStructuredError("budget_exceeded", err.Error(), false, "rerun with a larger budget")
			}
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to render PR impact: %v", err), false, "Check budget parameter")
		}
		return CallToolResult{Content: []ToolContent{{Type: "text", Text: string(out)}}}

	case "get_tour":
		var args struct {
			Focus    string `json:"focus"`
			MaxSteps int    `json:"max_steps"`
			Budget   int    `json:"budget"`
		}
		if len(params.Arguments) > 0 {
			if err := json.Unmarshal(params.Arguments, &args); err != nil {
				return newStructuredError("invalid_budget", fmt.Sprintf("Invalid arguments for get_tour: %v", err), false, "Check arguments format")
			}
		}
		tour, err := s.Engine.GetTour(args.Focus, args.MaxSteps)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return newStructuredError("node_not_found", fmt.Sprintf("Failed to get tour: %v", err), true, "Use locate_boundary to search for valid node IDs")
			}
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to get tour: %v", err), true, "Check focus node ID")
		}
		out, err := tour.Emit(slices.FormatJSON, args.Budget)
		if err != nil {
			if errors.Is(err, slices.ErrBudgetExceeded) {
				return newStructuredError("budget_exceeded", err.Error(), false, "rerun with a larger budget")
			}
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to render tour: %v", err), false, "Check budget parameter")
		}
		return CallToolResult{Content: []ToolContent{{Type: "text", Text: string(out)}}}

	case "get_doc":
		var args struct {
			ServiceID string `json:"service_id"`
			Budget    int    `json:"budget"`
		}
		if len(params.Arguments) > 0 {
			if err := json.Unmarshal(params.Arguments, &args); err != nil {
				return newStructuredError("invalid_budget", fmt.Sprintf("Invalid arguments for get_doc: %v", err), false, "Check arguments format")
			}
		}
		if args.ServiceID == "" {
			return newStructuredError("invalid_budget", "Missing required argument 'service_id'", false, "Provide service_id")
		}
		doc, err := s.Engine.GetDoc(args.ServiceID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return newStructuredError("node_not_found", fmt.Sprintf("Failed to get doc: %v", err), true, "Use locate_boundary to search for valid service IDs")
			}
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to get doc: %v", err), true, "Check service_id")
		}
		out, err := doc.Emit(slices.FormatJSON, args.Budget)
		if err != nil {
			if errors.Is(err, slices.ErrBudgetExceeded) {
				return newStructuredError("budget_exceeded", err.Error(), false, "rerun with a larger budget")
			}
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to render doc: %v", err), false, "Check budget parameter")
		}
		return CallToolResult{Content: []ToolContent{{Type: "text", Text: string(out)}}}

	case "diff_maps":
		var args struct {
			IndexA string `json:"index_a"`
			IndexB string `json:"index_b"`
			Budget int    `json:"budget"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return newStructuredError("invalid_budget", fmt.Sprintf("Invalid arguments for diff_maps: %v", err), false, "Check arguments format")
		}
		if args.IndexA == "" || args.IndexB == "" {
			return newStructuredError("invalid_budget", "Missing required arguments 'index_a' and 'index_b'", false, "Provide index_a and index_b paths")
		}
		st := store.NewStore("")
		a, err := st.LoadFile(args.IndexA)
		if err != nil {
			return mapLoadIndexError("index_a", err)
		}
		b, err := st.LoadFile(args.IndexB)
		if err != nil {
			return mapLoadIndexError("index_b", err)
		}
		out, err := archdiff.Compare(a, b).Emit("json", args.Budget)
		if err != nil {
			if errors.Is(err, slices.ErrBudgetExceeded) {
				return newStructuredError("budget_exceeded", err.Error(), false, "rerun with a larger budget")
			}
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to render diff: %v", err), false, "Check budget parameter")
		}
		return CallToolResult{Content: []ToolContent{{Type: "text", Text: string(out)}}}

	case "get_context_slice":
		var args struct {
			NodeID string `json:"node_id"`
			Tier   string `json:"tier"`
			View   string `json:"view"`
			Budget int    `json:"budget"`
			Format string `json:"format"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return newStructuredError("invalid_budget", fmt.Sprintf("Invalid arguments for get_context_slice: %v", err), false, "Check arguments format")
		}
		if args.NodeID == "" {
			return newStructuredError("invalid_budget", "Missing required argument 'node_id'", false, "Provide node_id")
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
			if strings.Contains(err.Error(), "not found") {
				return newStructuredError("node_not_found", fmt.Sprintf("Failed to get context slice: %v", err), true, "Use locate_boundary to search for valid node IDs")
			}
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to get context slice: %v", err), true, "Check node_id")
		}

		reqFmt := slices.FormatJSON
		if strings.ToLower(args.Format) == "caveman" || strings.ToLower(args.Format) == "llm" {
			reqFmt = slices.FormatCaveman
		} else if strings.ToLower(args.Format) == "md" || strings.ToLower(args.Format) == "markdown" {
			reqFmt = slices.FormatMarkdown
		}

		outJSON, err := slice.Emit(reqFmt, args.Budget)
		if err != nil {
			if errors.Is(err, slices.ErrBudgetExceeded) {
				return newStructuredError("budget_exceeded", err.Error(), false, "rerun with a larger budget")
			}
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to render context slice: %v", err), false, "Check budget parameter")
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
			Budget int    `json:"budget"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return newStructuredError("invalid_budget", fmt.Sprintf("Invalid arguments for explain_evidence: %v", err), false, "Check arguments format")
		}
		if args.NodeID == "" && args.EdgeID == "" {
			return newStructuredError("invalid_budget", "Either node_id or edge_id must be provided", false, "Provide either node_id or edge_id")
		}

		report, err := s.Engine.ExplainEvidence(args.NodeID, args.EdgeID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return newStructuredError("node_not_found", fmt.Sprintf("Failed to explain evidence: %v", err), true, "Check node_id or edge_id exists")
			}
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to explain evidence: %v", err), true, "Check node_id or edge_id exists")
		}

		outJSON, err := report.Emit(slices.FormatJSON, args.Budget)
		if err != nil {
			if errors.Is(err, slices.ErrBudgetExceeded) {
				return newStructuredError("budget_exceeded", err.Error(), false, "rerun with a larger budget")
			}
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to serialize canonical JSON: %v", err), false, "Check budget parameter")
		}

		return CallToolResult{
			Content: []ToolContent{
				{Type: "text", Text: string(outJSON)},
			},
		}

	case "locate_boundary":
		var args struct {
			Query  string `json:"query"`
			Budget int    `json:"budget"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return newStructuredError("invalid_budget", fmt.Sprintf("Invalid arguments for locate_boundary: %v", err), false, "Check arguments format")
		}
		if strings.TrimSpace(args.Query) == "" {
			return newStructuredError("invalid_budget", "Missing required argument 'query'", false, "Provide a non-empty query")
		}

		matches, err := s.Engine.LocateBoundary(args.Query)
		if err != nil {
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to locate boundary: %v", err), true, "Check query format")
		}

		// Q1: a call without a budget stays byte-identical to the CLI's locate
		// JSON (a bare array); only a budgeted call gains the envelope object.
		if args.Budget <= 0 {
			out, err := json.MarshalIndent(matches, "", "  ")
			if err != nil {
				return newStructuredError("envelope_error", fmt.Sprintf("Failed to format matches: %v", err), false, "Check query format")
			}
			out = append(out, '\n')
			return CallToolResult{Content: []ToolContent{{Type: "text", Text: string(out)}}}
		}

		result := slices.BoundaryResult{Matches: matches}
		out, err := result.Emit(slices.FormatJSON, args.Budget)
		if err != nil {
			if errors.Is(err, slices.ErrBudgetExceeded) {
				return newStructuredError("budget_exceeded", err.Error(), false, "rerun with a larger budget")
			}
			return newStructuredError("envelope_error", fmt.Sprintf("Failed to format matches: %v", err), false, "Check budget parameter")
		}

		return CallToolResult{
			Content: []ToolContent{
				{Type: "text", Text: string(out)},
			},
		}

	default:
		return newStructuredError("tool_not_found", fmt.Sprintf("Unknown tool: %s", params.Name), false, "Check available tools with tools/list")
	}
}

func mapLoadIndexError(targetName string, err error) CallToolResult {
	if errors.Is(err, store.ErrCorruptIndex) || strings.Contains(err.Error(), "index_corrupt") {
		return newStructuredError("index_corrupt", fmt.Sprintf("Corrupt index file %s: %v", targetName, err), false, "Index file is corrupted; re-index or recover")
	}
	if errors.Is(err, store.ErrUnsupportedSchema) || strings.Contains(err.Error(), "index_schema_unsupported") {
		return newStructuredError("index_schema_unsupported", fmt.Sprintf("Unsupported index schema for %s: %v", targetName, err), false, "Upgrade index schema to version 2 using 'gira index'")
	}
	if errors.Is(err, store.ErrToolMismatch) || strings.Contains(err.Error(), "index_tool_mismatch") {
		return newStructuredError("index_tool_mismatch", fmt.Sprintf("Tool mismatch for %s: %v", targetName, err), false, "Re-index with current tool binary")
	}
	if errors.Is(err, store.ErrIndexAmbiguous) || strings.Contains(err.Error(), "index_ambiguous") {
		return newStructuredError("index_ambiguous", fmt.Sprintf("Ambiguous index for %s: %v", targetName, err), false, "Specify unambiguous index path or run from directory with gira.yaml")
	}
	return newStructuredError("index_missing", fmt.Sprintf("Failed to load %s: %v", targetName, err), true, "Run gira index first")
}
