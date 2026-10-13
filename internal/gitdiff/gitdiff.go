// Package gitdiff reads a git revision range as changed line ranges, so that an
// impact briefing can start from a diff instead of a node ID (spec A.1,
// `gira impact --from-diff`). Line ranges are on the base side, because the index
// the briefing runs on describes the base revision.
package gitdiff

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// Hunk is a changed range of base-side lines. Count 0 is a pure insertion after
// line Start (Start 0 means before the first line).
type Hunk struct {
	Start int `json:"start"`
	Count int `json:"count"`
}

// FileChange lists the hunks of one file. Path is the base-side path; for an added
// file it is the new path and Added is set.
type FileChange struct {
	Path    string `json:"path"`
	Added   bool   `json:"added,omitempty"`
	Deleted bool   `json:"deleted,omitempty"`
	Hunks   []Hunk `json:"hunks"`
}

// Touches reports whether the change touches lines [start, end] of the base file.
// A locator without a line range (start 0) stands for the whole file.
func (c FileChange) Touches(start, end int) bool {
	if start <= 0 || c.Added || c.Deleted {
		return true
	}
	if end < start {
		end = start
	}
	for _, h := range c.Hunks {
		if h.Count == 0 {
			// An insertion between lines Start and Start+1 is inside [start, end]
			// only when both neighbours are.
			if start <= h.Start && h.Start+1 <= end {
				return true
			}
			continue
		}
		if h.Start <= end && start <= h.Start+h.Count-1 {
			return true
		}
	}
	return false
}

// Run returns the zero-context diff of a revision range ("base..head" or
// "base...head") in the repository at dir.
func Run(ctx context.Context, dir, revRange string) (string, error) {
	if !strings.Contains(revRange, "..") || strings.HasPrefix(revRange, "-") {
		return "", fmt.Errorf("invalid revision range %q: want <base>..<head>", revRange)
	}
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "diff", "--no-color", "--no-ext-diff", "--no-renames", "-U0", revRange, "--")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git diff %s: %s: %w", revRange, strings.TrimSpace(stderr.String()), err)
	}
	return stdout.String(), nil
}

// Parse reads a unified diff (any context size) into per-file changes, sorted by path.
func Parse(diff string) ([]FileChange, error) {
	var out []FileChange
	var cur *FileChange
	var oldPath, newPath string
	flush := func() {
		if cur != nil {
			out = append(out, *cur)
			cur = nil
		}
	}
	sc := bufio.NewScanner(strings.NewReader(diff))
	sc.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "diff --git "):
			flush()
			oldPath, newPath = "", ""
		case strings.HasPrefix(line, "--- "):
			oldPath = diffPath(line[4:])
		case strings.HasPrefix(line, "+++ "):
			newPath = diffPath(line[4:])
			fc := FileChange{Path: oldPath}
			switch {
			case oldPath == "":
				fc.Path, fc.Added = newPath, true
			case newPath == "":
				fc.Deleted = true
			}
			cur = &fc
		case strings.HasPrefix(line, "@@ "):
			if cur == nil {
				return nil, fmt.Errorf("hunk header before file header: %q", line)
			}
			h, err := parseHunk(line)
			if err != nil {
				return nil, err
			}
			cur.Hunks = append(cur.Hunks, h)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	flush()
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// diffPath strips the a/ or b/ prefix; /dev/null becomes "".
func diffPath(s string) string {
	s = strings.TrimSpace(strings.SplitN(s, "\t", 2)[0])
	if s == "/dev/null" {
		return ""
	}
	if unq, err := strconv.Unquote(s); err == nil {
		s = unq
	}
	if len(s) > 2 && (s[:2] == "a/" || s[:2] == "b/") {
		s = s[2:]
	}
	return s
}

// parseHunk reads the base side of "@@ -start[,count] +start[,count] @@".
func parseHunk(line string) (Hunk, error) {
	fields := strings.Fields(line)
	if len(fields) < 3 || !strings.HasPrefix(fields[1], "-") {
		return Hunk{}, fmt.Errorf("malformed hunk header %q", line)
	}
	spec := fields[1][1:]
	h := Hunk{Count: 1}
	parts := strings.SplitN(spec, ",", 2)
	var err error
	if h.Start, err = strconv.Atoi(parts[0]); err != nil {
		return Hunk{}, fmt.Errorf("malformed hunk header %q", line)
	}
	if len(parts) == 2 {
		if h.Count, err = strconv.Atoi(parts[1]); err != nil {
			return Hunk{}, fmt.Errorf("malformed hunk header %q", line)
		}
	}
	return h, nil
}
