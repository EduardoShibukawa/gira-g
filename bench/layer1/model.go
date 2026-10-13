// Package layer1 is the agent task benchmark runner (spec B, Layer 1). It runs the
// 12 ADR 0007 tasks for each arm and model, speaks MCP as a client to the arm's
// tool server, records telemetry, and scores answers with a deterministic set
// comparison against the ground truth. No LLM judges primary accuracy.
package layer1

import "context"

// Message is one turn of a conversation.
type Message struct {
	Role       string     `json:"role"` // system, user, assistant, tool
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall is a model's request to run one tool.
type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ToolSpec describes a tool offered to the model.
type ToolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// Usage is the token accounting a model API reports for one completion.
type Usage struct {
	Input     int `json:"input"`
	Output    int `json:"output"`
	CacheRead int `json:"cache_read"`
}

// Reply is one completion: either tool calls or a final answer.
type Reply struct {
	Content   string
	ToolCalls []ToolCall
	Usage     Usage
}

// Model is a chat model with tool use. ID is pinned and recorded per run.
type Model interface {
	ID() string
	Complete(ctx context.Context, msgs []Message, tools []ToolSpec) (Reply, error)
}
