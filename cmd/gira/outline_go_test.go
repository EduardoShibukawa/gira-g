package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOutlineGoSymbols(t *testing.T) {
	src := "package demo\n\nconst Max = 3\n\nvar global int\n\ntype Svc struct{}\n\nfunc (s *Svc) Run() error { return nil }\n\nfunc Helper(a int) int { return a }\n"
	p := filepath.Join(t.TempDir(), "demo.go")
	if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := runOutline([]string{p}, &out, &errb); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"L3: const Max", "L5: var global", "L7: type Svc", "L9: method (*Svc) Run", "L11: func Helper"} {
		if !strings.Contains(got, want) {
			t.Errorf("outline missing %q; got:\n%s", want, got)
		}
	}
}
