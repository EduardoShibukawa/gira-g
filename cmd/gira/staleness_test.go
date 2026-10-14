package main

import (
	"bytes"
	"strings"
	"testing"
)

// SC3: human commands print the indexed commit(s), and a local checkout whose
// HEAD is not an indexed commit is flagged stale.
func TestSCStaleness(t *testing.T) {
	_, index := hermeticIndex(t)
	mm, err := loadIndex(index)
	if err != nil {
		t.Fatal(err)
	}
	if len(mm.IndexedCommits()) == 0 {
		t.Skip("the hermetic index cites no commits")
	}

	// The indexed-commit banner is always on stderr.
	var out, errOut bytes.Buffer
	if err := Run([]string{"slice", "service:checkout", "--index", index, "--format", "md", "--repo", ""}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errOut.String(), "indexed commit:") {
		t.Fatalf("missing the indexed-commit banner: %q", errOut.String())
	}

	// A checkout whose HEAD the index does not cite is stale.
	dir := t.TempDir()
	gitIn(t, dir, "init", "-q")
	gitIn(t, dir, "commit", "-q", "--allow-empty", "-m", "unrelated")
	var out2, err2 bytes.Buffer
	if err := Run([]string{"slice", "service:checkout", "--index", index, "--format", "md", "--repo", dir}, strings.NewReader(""), &out2, &err2); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(err2.String(), "stale:") {
		t.Fatalf("expected a stale banner: %q", err2.String())
	}
}
