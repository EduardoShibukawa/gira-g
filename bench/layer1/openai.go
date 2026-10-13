package layer1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
)

// OpenAIModel talks to an OpenAI-compatible chat completions endpoint, such as a
// local Ollama server (http://localhost:11434). Exact model IDs are pinned by the
// caller and recorded in telemetry.
type OpenAIModel struct {
	Endpoint    string // base URL, without /v1/chat/completions
	Model       string
	Temperature float64
	APIKey      string
	// PaidApproval must name the owner's `approved:` entry (with its cost estimate)
	// before any non-local endpoint is called (ADR 0012 rule 5).
	PaidApproval string
	Client       *http.Client
}

// ID is the pinned model ID.
func (m OpenAIModel) ID() string { return m.Model }

// Local reports whether the endpoint is on this machine.
func Local(endpoint string) bool {
	u, err := url.Parse(endpoint)
	if err != nil {
		return false
	}
	host := u.Hostname()
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// Complete sends one chat completion request.
func (m OpenAIModel) Complete(ctx context.Context, msgs []Message, tools []ToolSpec) (Reply, error) {
	if !Local(m.Endpoint) && m.PaidApproval == "" {
		return Reply{}, fmt.Errorf("refusing non-local endpoint %s: a paid run needs an owner approval reference", m.Endpoint)
	}
	type fn struct {
		Name        string         `json:"name"`
		Description string         `json:"description,omitempty"`
		Parameters  map[string]any `json:"parameters,omitempty"`
		Arguments   string         `json:"arguments,omitempty"`
	}
	type call struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Function fn     `json:"function"`
	}
	type msg struct {
		Role       string `json:"role"`
		Content    string `json:"content"`
		ToolCalls  []call `json:"tool_calls,omitempty"`
		ToolCallID string `json:"tool_call_id,omitempty"`
	}
	req := struct {
		Model       string           `json:"model"`
		Messages    []msg            `json:"messages"`
		Tools       []map[string]any `json:"tools,omitempty"`
		Temperature float64          `json:"temperature"`
	}{Model: m.Model, Temperature: m.Temperature}
	for _, x := range msgs {
		om := msg{Role: x.Role, Content: x.Content, ToolCallID: x.ToolCallID}
		for _, c := range x.ToolCalls {
			args, _ := json.Marshal(c.Arguments)
			om.ToolCalls = append(om.ToolCalls, call{ID: c.ID, Type: "function", Function: fn{Name: c.Name, Arguments: string(args)}})
		}
		req.Messages = append(req.Messages, om)
	}
	for _, t := range tools {
		req.Tools = append(req.Tools, map[string]any{"type": "function", "function": fn{Name: t.Name, Description: t.Description, Parameters: t.Parameters}})
	}
	body, err := json.Marshal(req)
	if err != nil {
		return Reply{}, err
	}
	hr, err := http.NewRequestWithContext(ctx, http.MethodPost, m.Endpoint+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Reply{}, err
	}
	hr.Header.Set("Content-Type", "application/json")
	if m.APIKey != "" {
		hr.Header.Set("Authorization", "Bearer "+m.APIKey)
	}
	client := m.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(hr)
	if err != nil {
		return Reply{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Reply{}, fmt.Errorf("model endpoint: %s", resp.Status)
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []call `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			Prompt     int `json:"prompt_tokens"`
			Completion int `json:"completion_tokens"`
			Details    struct {
				Cached int `json:"cached_tokens"`
			} `json:"prompt_tokens_details"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Reply{}, err
	}
	if len(out.Choices) == 0 {
		return Reply{}, fmt.Errorf("model endpoint returned no choices")
	}
	r := Reply{Content: out.Choices[0].Message.Content, Usage: Usage{Input: out.Usage.Prompt, Output: out.Usage.Completion, CacheRead: out.Usage.Details.Cached}}
	for _, c := range out.Choices[0].Message.ToolCalls {
		args := map[string]any{}
		_ = json.Unmarshal([]byte(c.Function.Arguments), &args)
		r.ToolCalls = append(r.ToolCalls, ToolCall{ID: c.ID, Name: c.Function.Name, Arguments: args})
	}
	return r, nil
}
