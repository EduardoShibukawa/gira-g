package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestArchitectureCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runArchitecture([]string{"--format", "md"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runArchitecture error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Architectural Reference Overview") {
		t.Errorf("expected architectural overview title, got:\n%.200s", out)
	}
	if !strings.Contains(out, "Zero-Token Fast Pass") {
		t.Errorf("expected Zero-Token Fast Pass invariant in output")
	}
	if !strings.Contains(out, "```mermaid") {
		t.Errorf("expected mermaid block in output")
	}
	if !strings.Contains(out, "Critical Lifecycles & Intent Flows") {
		t.Errorf("expected lifecycles in output")
	}
}

func TestDocModuleAndSymbolQueries(t *testing.T) {
	// Test module query
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runDoc([]string{"manipulator", "--format", "md"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runDoc(manipulator) error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "internal/manipulator") || !strings.Contains(out, "AST Manipulator") {
		t.Errorf("unexpected doc output: %s", out)
	}

	// Test symbol query
	stdout.Reset()
	stderr.Reset()
	err = runDoc([]string{"FindSymbols", "--format", "md"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runDoc(FindSymbols) error: %v", err)
	}
	out = stdout.String()
	if !strings.Contains(out, "FindSymbols") || !strings.Contains(out, "func FindSymbols") {
		t.Errorf("unexpected symbol doc output: %s", out)
	}
}

func TestFlowsIntentQueries(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runFlows([]string{"ast_edit", "--format", "md"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runFlows(ast_edit) error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "Lifecycle of a Surgical AST Edit") {
		t.Errorf("unexpected flows output: %s", out)
	}
	if !strings.Contains(out, "sequenceDiagram") {
		t.Errorf("expected sequenceDiagram in flows output")
	}

	// Test JSON format
	stdout.Reset()
	stderr.Reset()
	err = runFlows([]string{"impact", "--format", "json"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runFlows(impact, json) error: %v", err)
	}
	out = stdout.String()
	if !strings.Contains(out, `"intent": "impact"`) {
		t.Errorf("expected json intent in output: %s", out)
	}
}
