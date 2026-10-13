package layer1

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Arm is one benchmark arm: the base tools, plus an MCP tool server for arms B-D.
type Arm struct {
	ID   string
	Base BaseTools
	MCP  *MCPClient
}

// Runner runs tasks for one model.
type Runner struct {
	Model       Model
	GT          *GroundTruth
	CorpusSHA   string
	Temperature float64
	MaxTurns    int
	// TranscriptDir, when set, receives one transcript per run. It must be inside
	// the git-ignored output directory: transcripts are never committed.
	TranscriptDir string
}

// Prompt is the fixed task prompt.
func Prompt(t Task) string {
	return fmt.Sprintf("Task: %s\nTarget: %s\nDepth: %d\n\nFind the Impact Set of changing the target: the Components it depends on (upstream), the Components that depend on it (downstream) and the Contracts affected. Cite evidence as locators. Reply with only this JSON:\n%s\n",
		t.Title, t.TargetID, t.Depth, AnswerSchema)
}

const systemPrompt = "You answer architecture questions about the repository in your tools. Use the tools; never guess IDs you have not seen."

// RunTask runs one task once and returns its telemetry.
func (r Runner) RunTask(ctx context.Context, arm Arm, t Task, rep int) Telemetry {
	tel := Telemetry{
		Schema: TelemetrySchemaID, TaskID: t.ID, Arm: arm.ID, ModelID: r.Model.ID(), Repetition: rep,
		CorpusSHA: r.CorpusSHA, Temperature: r.Temperature, ToolCallsByTool: map[string]int{}, Hallucinated: []string{},
	}
	start := time.Now()
	defer func() { tel.WallLatencyMS = time.Since(start).Milliseconds() }()

	tools := arm.Base.Specs()
	if arm.MCP != nil {
		more, err := arm.MCP.Tools()
		if err != nil {
			tel.Error = err.Error()
			return tel
		}
		tools = append(tools, more...)
	}
	msgs := []Message{{Role: "system", Content: systemPrompt}, {Role: "user", Content: Prompt(t)}}
	maxTurns := r.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 12
	}
	var final string
	for tel.Turns < maxTurns {
		reply, err := r.Model.Complete(ctx, msgs, tools)
		tel.Turns++
		if err != nil {
			tel.Error = err.Error()
			break
		}
		tel.TokensInput += reply.Usage.Input
		tel.TokensOutput += reply.Usage.Output
		tel.CacheReadTokens += reply.Usage.CacheRead
		msgs = append(msgs, Message{Role: "assistant", Content: reply.Content, ToolCalls: reply.ToolCalls})
		if len(reply.ToolCalls) == 0 {
			final = reply.Content
			break
		}
		for _, call := range reply.ToolCalls {
			tel.ToolCalls++
			tel.ToolCallsByTool[call.Name]++
			out, err := r.runTool(arm, call)
			if err != nil {
				out = "error: " + err.Error()
			}
			msgs = append(msgs, Message{Role: "tool", Content: out, ToolCallID: call.ID})
		}
	}
	r.saveTranscript(t, arm, rep, msgs)
	if final == "" {
		if tel.Error == "" {
			tel.Error = "no final answer within the turn limit"
		}
		return tel
	}
	ans, err := ParseAnswer(final)
	if err != nil {
		tel.Error = err.Error()
		return tel
	}
	tel.Answered = true
	ignore := map[string]bool{}
	for _, id := range r.GT.Ignore {
		ignore[id] = true
	}
	s := ScoreAnswer(ans, t, ignore, r.GT.Universe(), func(p string, a, b int) error { return arm.Base.CheckLocator(p, a, b) })
	tel.Precision, tel.Recall, tel.F1 = s.Precision, s.Recall, s.F1
	tel.LocatorValidity, tel.LocatorsTotal, tel.Hallucinated = s.LocatorValidity, s.LocatorsTotal, s.HallucinatedNodes
	return tel
}

func (r Runner) runTool(arm Arm, call ToolCall) (string, error) {
	if arm.Base.Has(call.Name) {
		return arm.Base.Run(call.Name, call.Arguments)
	}
	if arm.MCP == nil {
		return "", fmt.Errorf("tool %q is not offered in arm %s", call.Name, arm.ID)
	}
	return arm.MCP.Call(call.Name, call.Arguments)
}

func (r Runner) saveTranscript(t Task, arm Arm, rep int, msgs []Message) {
	if r.TranscriptDir == "" {
		return
	}
	data, err := json.MarshalIndent(msgs, "", "  ")
	if err != nil {
		return
	}
	dir := filepath.Join(r.TranscriptDir, arm.ID, r.Model.ID())
	if os.MkdirAll(dir, 0o755) == nil {
		_ = os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s.%d.json", t.ID, rep)), data, 0o644)
	}
}
