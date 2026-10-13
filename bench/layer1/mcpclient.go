package layer1

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// MCPClient speaks MCP (JSON-RPC 2.0, newline-delimited, over stdio) to a tool
// server process.
type MCPClient struct {
	cmd    *exec.Cmd
	in     io.WriteCloser
	out    *bufio.Reader
	mu     sync.Mutex
	nextID int
}

// StartMCP launches a server command and performs the MCP handshake.
func StartMCP(ctx context.Context, name string, args ...string) (*MCPClient, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	c := &MCPClient{cmd: cmd, in: in, out: bufio.NewReader(out)}
	if _, err := c.call("initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"clientInfo":      map[string]any{"name": "gira-bench-layer1", "version": "1"},
		"capabilities":    map[string]any{},
	}); err != nil {
		c.Close()
		return nil, err
	}
	if err := c.notify("notifications/initialized"); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

func (c *MCPClient) notify(method string) error {
	data, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "method": method})
	_, err := c.in.Write(append(data, '\n'))
	return err
}

func (c *MCPClient) call(method string, params any) (json.RawMessage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextID++
	data, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": c.nextID, "method": method, "params": params})
	if _, err := c.in.Write(append(data, '\n')); err != nil {
		return nil, err
	}
	line, err := c.out.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("mcp %s: %w", method, err)
	}
	var resp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("mcp %s: %s", method, resp.Error.Message)
	}
	return resp.Result, nil
}

// Tools lists the server's tools as model tool specs.
func (c *MCPClient) Tools() ([]ToolSpec, error) {
	raw, err := c.call("tools/list", map[string]any{})
	if err != nil {
		return nil, err
	}
	var res struct {
		Tools []struct {
			Name        string         `json:"name"`
			Description string         `json:"description"`
			InputSchema map[string]any `json:"inputSchema"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	out := make([]ToolSpec, 0, len(res.Tools))
	for _, t := range res.Tools {
		out = append(out, ToolSpec{Name: t.Name, Description: t.Description, Parameters: t.InputSchema})
	}
	return out, nil
}

// Call runs one tool and returns its text content.
func (c *MCPClient) Call(name string, args map[string]any) (string, error) {
	raw, err := c.call("tools/call", map[string]any{"name": name, "arguments": args})
	if err != nil {
		return "", err
	}
	var res struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return "", err
	}
	text := ""
	for _, c := range res.Content {
		text += c.Text
	}
	if res.IsError {
		return "error: " + text, nil
	}
	return text, nil
}

// Close stops the server.
func (c *MCPClient) Close() {
	_ = c.in.Close()
	_ = c.cmd.Wait()
}
