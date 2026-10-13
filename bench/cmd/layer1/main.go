// Command layer1 runs the Layer 1 agent task benchmark. Owner-run only.
//
//	go run ./cmd/layer1 --corpus out/corpus --corpus-sha <sha> --gira <bin> --index out/B/macro-map.json --model mock
//	go run ./cmd/layer1 ... --model openai --endpoint http://localhost:11434 --model-id llama3.1:8b
//
// A non-local endpoint is refused unless --approved-paid-run names the owner's
// approval entry, which must carry a cost estimate (ADR 0012 rule 5).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/bench/layer1"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "layer1:", err)
		os.Exit(1)
	}
}

func run() error {
	gtPath := flag.String("ground-truth", "../fixtures/evaluation/reference-corpus/ground-truth.v2.json", "ground-truth fixture")
	corpus := flag.String("corpus", "out/corpus", "corpus copy (README*, AGENTS.md and docs/ removed)")
	corpusSHA := flag.String("corpus-sha", "", "pinned corpus SHA, recorded in telemetry")
	gira := flag.String("gira", "gira", "gira binary for arm B")
	index := flag.String("index", "out/B/macro-map.json", "arm B index")
	arms := flag.String("arms", "A,B", "arms to run (A and B are supported by this runner)")
	model := flag.String("model", "mock", "mock or openai")
	endpoint := flag.String("endpoint", "http://localhost:11434", "OpenAI-compatible endpoint")
	modelID := flag.String("model-id", "", "pinned model ID")
	temperature := flag.Float64("temperature", 0, "sampling temperature")
	reps := flag.Int("reps", 5, "repetitions per task x arm x model cell (at least 5)")
	out := flag.String("out", "out/layer1", "output directory (git-ignored)")
	paid := flag.String("approved-paid-run", "", "owner approval reference for a paid endpoint")
	flag.Parse()

	if *reps < 5 {
		return fmt.Errorf("--reps must be at least 5")
	}
	data, err := os.ReadFile(*gtPath)
	if err != nil {
		return err
	}
	var gt layer1.GroundTruth
	if err := json.Unmarshal(data, &gt); err != nil {
		return err
	}
	var m layer1.Model = layer1.MockModel{}
	if *model == "openai" {
		if *modelID == "" {
			return fmt.Errorf("--model-id is required: model IDs are pinned at run time")
		}
		m = layer1.OpenAIModel{Endpoint: *endpoint, Model: *modelID, Temperature: *temperature, APIKey: os.Getenv("BENCH_API_KEY"), PaidApproval: *paid}
	}
	ctx := context.Background()
	runner := layer1.Runner{Model: m, GT: &gt, CorpusSHA: *corpusSHA, Temperature: *temperature, TranscriptDir: filepath.Join(*out, "transcripts")}
	base := layer1.BaseTools{Root: *corpus}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	tf, err := os.Create(filepath.Join(*out, "telemetry.jsonl"))
	if err != nil {
		return err
	}
	defer tf.Close()

	metrics := map[string]map[string][]float64{}
	for _, id := range strings.Split(*arms, ",") {
		arm := layer1.Arm{ID: id, Base: base}
		switch id {
		case "A":
		case "B":
			c, err := layer1.StartMCP(ctx, *gira, "serve", "mcp", "--index", *index)
			if err != nil {
				return err
			}
			defer c.Close()
			arm.MCP = c
		default:
			return fmt.Errorf("arm %s needs its container query phase (scripts/container.sh); not wired in this runner yet", id)
		}
		for _, task := range gt.Tasks {
			for rep := 1; rep <= *reps; rep++ {
				tel := runner.RunTask(ctx, arm, task, rep)
				line, _ := json.Marshal(tel)
				if err := layer1.ValidateTelemetry(line); err != nil {
					return fmt.Errorf("invalid telemetry: %w", err)
				}
				if _, err := tf.Write(append(line, '\n')); err != nil {
					return err
				}
				cell := fmt.Sprintf("%s/%s/%s", task.ID, id, m.ID())
				if metrics[cell] == nil {
					metrics[cell] = map[string][]float64{}
				}
				for k, v := range map[string]float64{
					"tokens_input": float64(tel.TokensInput), "tokens_output": float64(tel.TokensOutput), "tool_calls": float64(tel.ToolCalls),
					"turns": float64(tel.Turns), "wall_latency_ms": float64(tel.WallLatencyMS), "precision": tel.Precision, "recall": tel.Recall, "f1": tel.F1,
				} {
					metrics[cell][k] = append(metrics[cell][k], v)
				}
			}
		}
	}
	summary := map[string]map[string]layer1.Summary{}
	cells := make([]string, 0, len(metrics))
	for c := range metrics {
		cells = append(cells, c)
	}
	sort.Strings(cells)
	for _, c := range cells {
		summary[c] = map[string]layer1.Summary{}
		for k, xs := range metrics[c] {
			summary[c][k] = layer1.Summarize(xs)
		}
	}
	sdata, _ := json.MarshalIndent(map[string]any{"model_id": m.ID(), "corpus_sha": *corpusSHA, "temperature": *temperature, "base_tools": "grep, glob, read (no shell)", "cells": summary}, "", "  ")
	return os.WriteFile(filepath.Join(*out, "summary.json"), append(sdata, '\n'), 0o644)
}
