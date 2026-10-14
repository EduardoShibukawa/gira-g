package slices

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// SC2: the excerpt reads the cited lines plus the window, and reports
// unavailable for a missing file, a path escape, or an empty checkout.
func TestSCExplainSource(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a", "b.go")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	for i := 1; i <= 100; i++ {
		fmt.Fprintf(&sb, "line %d\n", i)
	}
	if err := os.WriteFile(p, []byte(sb.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	loc := domain.Locator{Repo: "r", CommitSHA: "c", Path: "a/b.go", StartLine: 50, EndLine: 52}
	ex, ok := Excerpt(dir, loc, 2)
	if !ok {
		t.Fatal("expected an excerpt")
	}
	for _, want := range []string{"line 48", "line 50", "line 54"} {
		if !strings.Contains(ex, want) {
			t.Fatalf("window missing %q:\n%s", want, ex)
		}
	}
	for _, no := range []string{"line 47", "line 55"} {
		if strings.Contains(ex, no) {
			t.Fatalf("window too wide (%q):\n%s", no, ex)
		}
	}

	if _, ok := Excerpt(dir, domain.Locator{Path: "missing.go", StartLine: 1, EndLine: 1}, 2); ok {
		t.Fatal("missing file must be unavailable")
	}
	if _, ok := Excerpt(dir, domain.Locator{Path: "../escape.go", StartLine: 1, EndLine: 1}, 2); ok {
		t.Fatal("path traversal must be unavailable")
	}
	if _, ok := Excerpt("", loc, 2); ok {
		t.Fatal("empty checkout root must be unavailable")
	}

	// A single-point locator (EndLine == 0) still shows the cited line.
	one, ok := Excerpt(dir, domain.Locator{Path: "a/b.go", StartLine: 100, EndLine: 0}, 2)
	if !ok {
		t.Fatal("single-point locator must read")
	}
	if !strings.Contains(one, "line 100") {
		t.Fatalf("single-point locator must show the cited line:\n%s", one)
	}

	// A symlink escaping the checkout is rejected.
	secret := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(secret, []byte("secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(secret, link); err != nil {
		t.Fatal(err)
	}
	if _, ok := Excerpt(dir, domain.Locator{Path: "link.txt", StartLine: 1, EndLine: 1}, 2); ok {
		t.Fatal("a symlink escaping the checkout must be unavailable")
	}
}
