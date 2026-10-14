package eval_test

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/eval"
)

// SC1 (spec gira-recall-joins J1): stdlib Health never becomes a Contract, so
// task-01 and task-07 report no contract:Health in unexpected, in both
// hermetic variants.
func TestSC1HealthDenylist(t *testing.T) {
	for _, v := range hermeticVariants(t) {
		for _, n := range v.mm.Nodes {
			if n.ID == "contract:Health" {
				t.Errorf("%s: contract:Health reached the map", v.name)
			}
		}
		ev := eval.NewEvaluator(v.mm, v.gt.Ignore, nil, eval.DefaultThresholds())
		checked := 0
		for _, task := range v.gt.Tasks {
			if !strings.HasPrefix(task.ID, "task-01-") && !strings.HasPrefix(task.ID, "task-07-") {
				continue
			}
			checked++
			for _, id := range ev.EvaluateTask(context.Background(), task).UnexpectedIDs {
				if id == "contract:Health" {
					t.Errorf("%s %s: contract:Health in unexpected", v.name, task.ID)
				}
			}
		}
		if checked != 2 {
			t.Errorf("%s: checked %d of task-01/task-07", v.name, checked)
		}
	}
}

// SC2 (J2): no service→service consumes edge rests on compose depends_on
// alone; an edge that cites a depends_on line also cites its corroborating
// line. task-02 and task-11 report no ordering-only edges.
func TestSC2DependsOnCorroboration(t *testing.T) {
	for _, v := range hermeticVariants(t) {
		ordering := dependsOnLines(t, v.dir)
		corroborated := 0
		for _, r := range v.mm.Relationships {
			if r.Verb != domain.VerbConsumes || !strings.HasPrefix(r.SourceID, "service:") || !strings.HasPrefix(r.TargetID, "service:") {
				continue
			}
			cites, other := false, false
			for _, e := range r.Evidence {
				if ordering[e.Locator.Path+"#"+strconv.Itoa(e.Locator.StartLine)] {
					cites = true
				} else {
					other = true
				}
			}
			if cites && !other {
				t.Errorf("%s: %s rests on depends_on only", v.name, r.ID)
			}
			if cites && other {
				corroborated++
			}
		}
		if corroborated == 0 {
			t.Errorf("%s: no corroborated depends_on edge cites both lines; the law would pass vacuously", v.name)
		}

		frontend, catalog := "service:frontend", "service:product-catalog"
		if v.name == "perturbed" {
			frontend, catalog = "service:storefront", "service:inventory"
		}
		ordOnly := map[string]bool{frontend + ">service:quote": true, frontend + ">service:image-provider": true, "service:agent>" + catalog: true}
		for _, r := range v.mm.Relationships {
			if r.Verb == domain.VerbConsumes && ordOnly[r.SourceID+">"+r.TargetID] {
				t.Errorf("%s: ordering-only edge %s survived", v.name, r.ID)
			}
		}
		ev := eval.NewEvaluator(v.mm, v.gt.Ignore, nil, eval.DefaultThresholds())
		for _, task := range v.gt.Tasks {
			if !strings.HasPrefix(task.ID, "task-02-") && !strings.HasPrefix(task.ID, "task-11-") {
				continue
			}
			for _, id := range ev.EvaluateTask(context.Background(), task).UnexpectedIDs {
				if id == "service:quote" || id == "service:image-provider" || id == catalog {
					t.Errorf("%s %s: ordering-only %s in unexpected", v.name, task.ID, id)
				}
			}
		}
	}
}

// dependsOnLines returns "path#line" for every entry line inside a
// depends_on block of the corpus's top-level compose files.
func dependsOnLines(t *testing.T, dir string) map[string]bool {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(dir, "compose*.yaml"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no compose files under %s: %v", dir, err)
	}
	out := map[string]bool{}
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			t.Fatal(err)
		}
		rel := filepath.Base(p)
		block := -1 // indent of the open depends_on key, -1 when none
		sc := bufio.NewScanner(f)
		for n := 1; sc.Scan(); n++ {
			line := sc.Text()
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			indent := len(line) - len(strings.TrimLeft(line, " "))
			if block >= 0 && indent > block {
				out[rel+"#"+strconv.Itoa(n)] = true
				continue
			}
			block = -1
			if trimmed == "depends_on:" {
				block = indent
			}
		}
		f.Close()
	}
	return out
}
