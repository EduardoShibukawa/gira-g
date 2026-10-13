package corpus_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/corpus"
	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for p, c := range files {
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestMaterializeIsDeterministic(t *testing.T) {
	src := writeTree(t, map[string]string{"a.txt": "a", "src/b/main.go": "package main\n"})
	sha1, err := corpus.Materialize(context.Background(), src, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sha2, err := corpus.Materialize(context.Background(), src, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if sha1 != sha2 || len(sha1) != 40 {
		t.Fatalf("same tree must give the same 40-char SHA: %q vs %q", sha1, sha2)
	}
}

func TestPerturbRenamesLongestFirstAndEdits(t *testing.T) {
	src := writeTree(t, map[string]string{"src/cart/x.go": "ProductCartService CartService\nkeep me\n"})
	dst := t.TempDir()
	p := corpus.Perturbation{
		Renames: []corpus.Rename{{From: "Cart", To: "Basket"}, {From: "ProductCart", To: "Inventory"}, {From: "cart", To: "basket"}},
		Edits:   []corpus.Edit{{Path: "src/cart/x.go", Find: "keep me\n", Replace: ""}},
	}
	if err := corpus.Perturb(src, dst, p); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "src/basket/x.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "InventoryService BasketService\n" {
		t.Fatalf("unexpected perturbation: %q", got)
	}
}

func TestPerturbFailsOnDriftedEdit(t *testing.T) {
	src := writeTree(t, map[string]string{"a.txt": "x"})
	err := corpus.Perturb(src, t.TempDir(), corpus.Perturbation{Edits: []corpus.Edit{{Path: "a.txt", Find: "missing", Replace: ""}}})
	if err == nil {
		t.Fatal("an edit that does not match must fail")
	}
	err = corpus.Perturb(src, t.TempDir(), corpus.Perturbation{Edits: []corpus.Edit{{Path: "gone.txt", Find: "x", Replace: ""}}})
	if err == nil {
		t.Fatal("an edit on a missing file must fail")
	}
}

func TestLeaksMatchesMemoryLeakRule(t *testing.T) {
	for p, want := range map[string]bool{
		"README.md":              true,
		"src/cart/README":        true,
		"AGENTS.md":              true,
		"docs/arch.md":           true,
		"src/frontend/docs/x.md": true,
		"src/checkout/main.go":   false,
		"docsite/index.md":       false,
		"pb/demo.proto":          false,
	} {
		if got := corpus.Leaks(p); got != want {
			t.Errorf("Leaks(%q) = %t, want %t", p, got, want)
		}
	}
}

func TestFetchUpstreamDropsLeaksFromPinnedCommit(t *testing.T) {
	ctx := context.Background()
	src := t.TempDir()
	for p, body := range map[string]string{
		"README.md":        "leak",
		"docs/guide.md":    "leak",
		"AGENTS.md":        "leak",
		"src/a/main.go":    "package main\n",
		"src/a/README.txt": "leak",
		"compose.yaml":     "services: {}\n",
	} {
		full := filepath.Join(src, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	repoDir := filepath.Join(t.TempDir(), "upstream")
	sha, err := corpus.Materialize(ctx, src, repoDir)
	if err != nil {
		t.Fatal(err)
	}
	cm, err := gitcache.NewCacheManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	dst := t.TempDir()
	n, err := corpus.FetchUpstream(ctx, cm, manifest.RepositoryDeclaration{ID: "u", URL: repoDir, SHA: sha, PathScope: "."}, dst, false)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("want 2 files written, got %d", n)
	}
	for _, p := range []string{"src/a/main.go", "compose.yaml"} {
		if _, err := os.Stat(filepath.Join(dst, p)); err != nil {
			t.Errorf("%s missing: %v", p, err)
		}
	}
}
