// Package windowsgates holds throwaway experiments that record how the
// platform behaves. Tests never fail on an "unsupported" outcome: they log
// one JSON line per observation (prefix RESULT) and, when RESULTS_FILE is
// set, append the same line to that file.
package windowsgates

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var resultsMu sync.Mutex

// rec logs one structured observation.
func rec(t *testing.T, experiment string, fields map[string]any) {
	t.Helper()
	m := map[string]any{"os": runtime.GOOS, "arch": runtime.GOARCH, "experiment": experiment}
	for k, v := range fields {
		m[k] = v
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	t.Logf("RESULT %s", b)
	if f := os.Getenv("RESULTS_FILE"); f != "" {
		resultsMu.Lock()
		defer resultsMu.Unlock()
		fh, err := os.OpenFile(f, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatalf("open RESULTS_FILE: %v", err)
		}
		defer fh.Close()
		_, _ = fh.Write(append(b, '\n'))
	}
}

// errStr renders an error (or nil) for JSON.
func errStr(err error) any {
	if err == nil {
		return nil
	}
	return err.Error()
}

type cmdResult struct {
	Exit   int    `json:"exit"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Err    any    `json:"start_error"`
}

// run executes a command with a clean environment: the parent environment
// minus every GIT_* variable, plus extra. Exit code -1 means it never ran.
func run(dir string, stdin []byte, extra []string, name string, args ...string) cmdResult {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	var env []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(strings.ToUpper(kv), "GIT_") {
			continue
		}
		env = append(env, kv)
	}
	cmd.Env = append(env, extra...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	err := cmd.Run()
	res := cmdResult{Stdout: so.String(), Stderr: se.String()}
	var ee *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &ee):
		res.Exit = ee.ExitCode()
	default:
		res.Exit = -1
		res.Err = err.Error()
	}
	return res
}

// runPipe executes cmd1 piped to cmd2 in targetDir with filtered GIT_* environment.
func runPipe(dir string, env []string, targetDir string, name1 string, args1 []string, name2 string, args2 []string) (exit1, exit2 int, stderr1, stderr2 string) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	c1 := exec.CommandContext(ctx, name1, args1...)
	c1.Dir = dir
	var filteredEnv []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(strings.ToUpper(kv), "GIT_") {
			continue
		}
		filteredEnv = append(filteredEnv, kv)
	}
	c1.Env = append(filteredEnv, env...)

	c2 := exec.CommandContext(ctx, name2, args2...)
	c2.Dir = targetDir
	c2.Env = append(filteredEnv, env...)

	pipe, err := c1.StdoutPipe()
	if err != nil {
		return -1, -1, err.Error(), ""
	}
	c2.Stdin = pipe
	var s1, s2 bytes.Buffer
	c1.Stderr = &s1
	c2.Stderr = &s2

	if err := c2.Start(); err != nil {
		return -1, -1, "", err.Error()
	}
	err1 := c1.Run()
	err2 := c2.Wait()

	exit1 = 0
	var ee1 *exec.ExitError
	if err1 != nil {
		if errors.As(err1, &ee1) {
			exit1 = ee1.ExitCode()
		} else {
			exit1 = -1
		}
	}
	exit2 = 0
	var ee2 *exec.ExitError
	if err2 != nil {
		if errors.As(err2, &ee2) {
			exit2 = ee2.ExitCode()
		} else {
			exit2 = -1
		}
	}
	return exit1, exit2, s1.String(), s2.String()
}

// isolatedHome creates an empty HOME and returns the env that points every
// per-user location at it. System config is left alone unless noSystem.
func isolatedHome(t *testing.T, noSystem bool) (home string, env []string) {
	t.Helper()
	home = t.TempDir()
	env = []string{
		"HOME=" + home,
		"USERPROFILE=" + home,
		"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"),
		"XDG_CACHE_HOME=" + filepath.Join(home, ".cache"),
		"GIT_TERMINAL_PROMPT=0",
	}
	if noSystem {
		env = append(env, "GIT_CONFIG_NOSYSTEM=1")
	}
	return home, env
}

var identity = []string{
	"GIT_AUTHOR_NAME=gates", "GIT_AUTHOR_EMAIL=gates@example.invalid",
	"GIT_COMMITTER_NAME=gates", "GIT_COMMITTER_EMAIL=gates@example.invalid",
	"GIT_AUTHOR_DATE=2026-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2026-01-01T00:00:00Z",
}

func git(dir string, env []string, args ...string) cmdResult {
	return run(dir, nil, env, "git", args...)
}

func trim(s string) string { return strings.TrimRight(s, "\r\n") }

// treeEntry is one index entry to commit without touching the working tree,
// so names the filesystem cannot hold (tab on Windows) still get committed.
type treeEntry struct {
	mode, path, content string
}

// makeRepo builds a non-bare repository at dir with the given entries. Trees
// are written with mktree and committed with commit-tree: no index and no
// worktree file is involved, so names the filesystem cannot hold (tab on
// Windows) are committed too. `update-index` is not usable for them: Git for
// Windows ignores such paths (see TestTabFilenameConstruction). Returns the
// commit sha.
func makeRepo(t *testing.T, dir string, env []string, entries []treeEntry) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if r := git(dir, env, "init", "-q", "-b", "main", dir); r.Exit != 0 {
		t.Fatalf("git init: %+v", r)
	}
	var build func(es []treeEntry) string
	build = func(es []treeEntry) string {
		subdirs := map[string][]treeEntry{}
		var order []string
		var buf bytes.Buffer
		for _, e := range es {
			head, rest, nested := strings.Cut(e.path, "/")
			if nested {
				if _, ok := subdirs[head]; !ok {
					order = append(order, head)
				}
				subdirs[head] = append(subdirs[head], treeEntry{e.mode, rest, e.content})
				continue
			}
			r := run(dir, []byte(e.content), env, "git", "hash-object", "-w", "--stdin")
			if r.Exit != 0 {
				t.Fatalf("hash-object: %+v", r)
			}
			buf.WriteString(e.mode + " blob " + trim(r.Stdout) + "\t" + head + "\x00")
		}
		for _, d := range order {
			buf.WriteString("040000 tree " + build(subdirs[d]) + "\t" + d + "\x00")
		}
		r := run(dir, buf.Bytes(), env, "git", "mktree", "-z")
		if r.Exit != 0 {
			t.Fatalf("mktree: %+v", r)
		}
		return trim(r.Stdout)
	}
	tree := build(entries)
	cenv := append(append([]string{}, env...), identity...)
	r := run(dir, nil, cenv, "git", "commit-tree", "-m", "fixture", tree)
	if r.Exit != 0 {
		t.Fatalf("commit-tree: %+v", r)
	}
	sha := trim(r.Stdout)
	if r := git(dir, env, "update-ref", "refs/heads/main", sha); r.Exit != 0 {
		t.Fatalf("update-ref: %+v", r)
	}
	return sha
}
