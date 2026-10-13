package corpus_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/corpus"
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
