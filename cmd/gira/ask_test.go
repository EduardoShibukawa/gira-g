package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestAskQuery(t *testing.T) {
	mm := &domain.MacroMap{
		TargetSystem: "duh-homelab/gira-g",
		Nodes: []domain.Node{
			{
				ID:          "pkg:docsite",
				Name:        "docsite",
				Kind:        "package",
				Description: "Living documentation builder and Cytoscape exporter",
				Evidence: []domain.EvidenceRecord{
					{Locator: domain.Locator{Path: "internal/docsite/landing.go"}},
				},
			},
			{
				ID:          "cmd:setup",
				Name:        "setup",
				Kind:        "command",
				Description: "One-shot developer workspace initialization",
				Evidence: []domain.EvidenceRecord{
					{Locator: domain.Locator{Path: "cmd/gira/main.go"}},
				},
			},
		},
		Relationships: []domain.Relationship{
			{
				ID:       "rel:setup->docsite",
				SourceID: "cmd:setup",
				TargetID: "pkg:docsite",
				Verb:     "calls",
			},
		},
	}

	t.Run("purpose question", func(t *testing.T) {
		resp := processAskQuery("Qual o objetivo desse projeto?", mm, false)
		if resp.Intent != "purpose" {
			t.Fatalf("expected intent 'purpose', got %s", resp.Intent)
		}
		if !strings.Contains(resp.Answer, "Objetivo e Propósito do Projeto") {
			t.Errorf("expected purpose title in answer, got: %s", resp.Answer)
		}
	})

	t.Run("quickstart question with diagram", func(t *testing.T) {
		resp := processAskQuery("Como eu utilizo o sistema?", mm, true)
		if resp.Intent != "quickstart" {
			t.Fatalf("expected intent 'quickstart', got %s", resp.Intent)
		}
		if resp.Diagram == "" || !strings.Contains(resp.Diagram, "graph") {
			t.Errorf("expected mermaid diagram, got: %s", resp.Diagram)
		}
	})

	t.Run("component impact question", func(t *testing.T) {
		resp := processAskQuery("O que quebra se alterar docsite?", mm, true)
		if resp.Intent != "impact" {
			t.Fatalf("expected intent 'impact', got %s", resp.Intent)
		}
		if len(resp.MatchedNodes) == 0 || resp.MatchedNodes[0] != "pkg:docsite" {
			t.Errorf("expected matched node pkg:docsite, got %v", resp.MatchedNodes)
		}
		if !strings.Contains(resp.Answer, "cmd:setup") {
			t.Errorf("expected caller cmd:setup in answer, got: %s", resp.Answer)
		}
	})

	t.Run("runAsk CLI invocation", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := runAsk([]string{"--json", "qual o propósito?"}, strings.NewReader(""), &stdout, &stderr)
		if err != nil {
			t.Fatalf("unexpected error running ask: %v", err)
		}
		outStr := stdout.String()
		if !strings.Contains(outStr, `"intent": "purpose"`) {
			t.Errorf("expected json intent purpose, got: %s", outStr)
		}
	})
}
