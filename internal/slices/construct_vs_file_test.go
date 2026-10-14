package slices_test

import (
	"context"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/extractors"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

// F7: an unsupported construct inside a parsed (supported) Go file is not an
// unsupported file. Export and LocateBoundary count only unsupported files;
// the briefing keeps the construct's explicit unknown row (J3: never a silent
// miss).
func TestUnsupportedConstructIsNotUnsupportedFile(t *testing.T) {
	const goPath = "src/catalog/main.go"
	const pyPath = "src/worker/app.py"
	files := []extractors.TargetFile{
		{Repo: "r", CommitSHA: "c", Path: goPath, Content: []byte("package main\n\nfunc main() {\n\tstartOpAMPClient()\n}\n")},
		{Repo: "r", CommitSHA: "c", Path: pyPath, Content: []byte("print('hi')\n")},
	}
	mm, err := extractors.DefaultPipeline().ProcessFiles(context.Background(), "t", files)
	if err != nil {
		t.Fatal(err)
	}

	// Export (md + html): one Unsupported Boundary (the .py file), never the Go file.
	md := slices.ExportMarkdown(mm)
	if !strings.Contains(md, "- Unsupported Boundaries: 1\n") {
		t.Errorf("md header must count only unsupported files (1):\n%s", md)
	}
	sec := md[strings.Index(md, "## Unsupported Boundaries & Known Unknowns"):]
	if end := strings.Index(sec[1:], "\n## "); end >= 0 {
		sec = sec[:end+1]
	}
	if !strings.Contains(sec, "(1)") || strings.Contains(sec, goPath) || !strings.Contains(sec, pyPath) {
		t.Errorf("md Unsupported Boundaries section must list only %s:\n%s", pyPath, sec)
	}
	if !strings.Contains(md, goPath) {
		t.Errorf("md must still surface the construct at %s (no silent miss):\n%s", goPath, md)
	}
	html := slices.ExportHTML(mm)
	if !strings.Contains(html, "Unsupported Boundaries: <strong>1</strong>") ||
		!strings.Contains(html, "<h2>Unsupported Boundaries & Known Unknowns (1)</h2>") {
		t.Errorf("html must count only unsupported files (1)")
	}

	engine := slices.NewEngine(mm)

	// LocateBoundary: the Go file is not an unsupported boundary match.
	for _, q := range []string{"catalog", "main.go", "OpAMP"} {
		ms, err := engine.LocateBoundary(q)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range ms {
			if m.Kind == "unsupported" && m.ID == goPath {
				t.Errorf("LocateBoundary(%q) reported the parsed Go file as an Unsupported Boundary: %+v", q, m)
			}
		}
	}
	ms, err := engine.LocateBoundary("app.py")
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 1 || ms[0].Kind != "unsupported" || ms[0].ID != pyPath {
		t.Errorf("LocateBoundary(app.py) must still find the unsupported file, got %+v", ms)
	}

	// Briefing: the construct keeps its explicit unknown row with a locator.
	var svc string
	for _, n := range mm.Nodes {
		for _, ev := range n.Evidence {
			if ev.Locator.Path == goPath && strings.HasPrefix(n.ID, "service:") {
				svc = n.ID
			}
		}
	}
	if svc == "" {
		t.Fatalf("no service node extracted from %s: %+v", goPath, mm.Nodes)
	}
	b, err := engine.GetImpactBriefing(svc, 1)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, ub := range b.UnsupportedBoundaries {
		if ub.Path == goPath && ub.Locator.StartLine == 4 && ub.Reason != "" {
			found = true
		}
	}
	if !found {
		t.Errorf("briefing for %s lost the construct's unknown row at %s#L4: %+v", svc, goPath, b.UnsupportedBoundaries)
	}
}
