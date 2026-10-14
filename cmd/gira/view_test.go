package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// SC4: `gira view --format html` writes a self-contained index.html, honors
// --no-open, and opens the browser only without it.
func TestSCViewExport(t *testing.T) {
	_, index := hermeticIndex(t)

	dir := t.TempDir()
	if err := Run([]string{"view", "--index", index, "--output", dir, "--no-open"}, strings.NewReader(""), io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	html, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatalf("view must write index.html: %v", err)
	}
	for _, bad := range []string{"<script src=", "<link ", `src="http`, `href="http`, "@import", "url(http"} {
		if bytes.Contains(html, []byte(bad)) {
			t.Fatalf("HTML view is not self-contained: contains %q", bad)
		}
	}

	if runtime.GOOS != "linux" {
		return
	}

	// A fake xdg-open records that it was invoked.
	bin := t.TempDir()
	marker := filepath.Join(t.TempDir(), "opened")
	script := "#!/bin/sh\necho \"$1\" > " + marker + "\n"
	if err := os.WriteFile(filepath.Join(bin, "xdg-open"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("DISPLAY", ":0")
	t.Setenv("CI", "")
	t.Setenv("NO_BROWSER", "")

	if err := Run([]string{"view", "--index", index, "--output", t.TempDir()}, strings.NewReader(""), io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("view should open the browser when --no-open is absent")
	}

	_ = os.Remove(marker)
	if err := Run([]string{"view", "--index", index, "--output", t.TempDir(), "--no-open"}, strings.NewReader(""), io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("--no-open must not open the browser")
	}
}
