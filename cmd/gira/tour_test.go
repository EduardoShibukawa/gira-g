package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/corpus"
	"github.com/EduardoShibukawa/gira-g/internal/eval"
	"github.com/EduardoShibukawa/gira-g/internal/mcp"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

var locatorRE = regexp.MustCompile(`^([^@]+)@([0-9a-f]{40}):([^#]+)(?:#L(\d+)(?:-L(\d+))?)?$`)

// resolveLocator checks a locator against the materialized corpus repository:
// the commit is the pinned one, the file exists there and the range is in bounds.
func resolveLocator(t *testing.T, repo, sha, loc string) error {
	t.Helper()
	m := locatorRE.FindStringSubmatch(loc)
	if m == nil {
		return fmt.Errorf("malformed locator %q", loc)
	}
	if m[2] != sha {
		return fmt.Errorf("%s: cites commit %s, pinned is %s", loc, m[2], sha)
	}
	content := gitIn(t, repo, "show", sha+":"+m[3])
	lines := strings.Count(content, "\n") + 1
	if m[4] != "" {
		start, _ := strconv.Atoi(m[4])
		end := start
		if m[5] != "" {
			end, _ = strconv.Atoi(m[5])
		}
		if start < 1 || end < start || end > lines {
			return fmt.Errorf("%s: range out of bounds (file has %d lines)", loc, lines)
		}
	}
	return nil
}

// SC8: tours are byte-identical, every locator resolves in bounds at the pinned
// SHA, every why comes from the template renderer, and the focused tour covers at
// least 80% of the ground-truth order flow.
func TestSC8ToursAreDeterministicAndVerifiable(t *testing.T) {
	gt, index := hermeticIndex(t)
	_, c, err := eval.Hermetic(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	repo := filepath.Join(t.TempDir(), "repo")
	sha, err := corpus.Materialize(context.Background(), c.Dir, repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(gt.Flows) == 0 {
		t.Fatal("ground truth declares no flow")
	}
	flow := gt.Flows[0]

	mm, err := loadIndex(index)
	if err != nil {
		t.Fatal(err)
	}
	server := mcp.NewServer("gira", "test", slices.NewEngine(mm))

	for _, focus := range []string{"", flow.Entry} {
		args := []string{"tour", "--index", index}
		if focus != "" {
			args = append(args, "--focus", focus)
		}
		a, err := runCLI(t, args...)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := runCLI(t, args...)
		if !bytes.Equal(a, b) {
			t.Fatalf("tour %q is not byte-identical across runs", focus)
		}
		ma, _ := runCLI(t, append(args, "--format", "md")...)
		mb, _ := runCLI(t, append(args, "--format", "md")...)
		if !bytes.Equal(ma, mb) {
			t.Fatalf("markdown tour %q is not byte-identical across runs", focus)
		}
		text, isErr := mcpCall(t, server, "get_tour", map[string]any{"focus": focus})
		if isErr || text != string(a) {
			t.Fatalf("get_tour differs from the CLI tour %q", focus)
		}

		var tour slices.Tour
		if err := json.Unmarshal(a, &tour); err != nil {
			t.Fatal(err)
		}
		if len(tour.Steps) == 0 {
			t.Fatalf("tour %q is empty", focus)
		}
		for _, s := range tour.Steps {
			if s.Why != slices.RenderWhy(s.WhyTemplate, s.WhyArgs) || s.Why == "unknown" {
				t.Errorf("step %d: why %q is not the template renderer's output", s.Order, s.Why)
			}
			// I1: the why claim cites its own locator, or says its locator is unknown.
			if s.WhyLocator == "" {
				if !strings.Contains(s.Why, "unknown") {
					t.Errorf("step %d: why %q has neither a locator nor an unknown marker", s.Order, s.Why)
				}
			} else if err := resolveLocator(t, repo, sha, s.WhyLocator); err != nil {
				t.Errorf("step %d why_locator: %v", s.Order, err)
			}
			if len(s.Read) == 0 {
				t.Errorf("step %d has nothing to read", s.Order)
			}
			for _, loc := range s.Read {
				if err := resolveLocator(t, repo, sha, loc); err != nil {
					t.Errorf("step %d: %v", s.Order, err)
				}
			}
		}

		if focus == flow.Entry {
			cov, missed := eval.TourCoverage(&tour, flow)
			t.Logf("order-flow coverage of the tour focused on %s: %.1f%%, missed %v", focus, cov*100, missed)
			if cov < 0.8 {
				t.Fatalf("critical-path coverage %.1f%% < 80%%, missed %v", cov*100, missed)
			}
		}
	}
}

// Flows cite one locator per hop, resolving at the pinned SHA, and are stable.
func TestFlowsCiteResolvableLocators(t *testing.T) {
	_, index := hermeticIndex(t)
	_, c, err := eval.Hermetic(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	repo := filepath.Join(t.TempDir(), "repo")
	sha, err := corpus.Materialize(context.Background(), c.Dir, repo)
	if err != nil {
		t.Fatal(err)
	}
	a, err := runCLI(t, "flows", "service:checkout", "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := runCLI(t, "flows", "service:checkout", "--index", index)
	if !bytes.Equal(a, b) {
		t.Fatal("flows are not byte-identical")
	}
	var f slices.Flows
	if err := json.Unmarshal(a, &f); err != nil {
		t.Fatal(err)
	}
	reachedKafka := false
	for _, fl := range f.Flows {
		for _, h := range fl.Steps[1:] {
			if h.Locator == "" {
				t.Errorf("hop to %s has no locator", h.NodeID)
				continue
			}
			if err := resolveLocator(t, repo, sha, h.Locator); err != nil {
				t.Error(err)
			}
			reachedKafka = reachedKafka || h.NodeID == "service:kafka"
		}
	}
	if !reachedKafka {
		t.Fatal("checkout's flows must reach the kafka broker")
	}
}

func TestFlowAnchorsResolveInSnapshot(t *testing.T) {
	gt, c, err := eval.Hermetic(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, v := eval.CheckFlowAnchors(gt.Flows, eval.DirReader(c.Dir)); len(v) > 0 {
		t.Fatal(v)
	}
}
