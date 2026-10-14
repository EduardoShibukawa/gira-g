package slices

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// SourceUnavailable is the marker shown when a locator's source cannot be read.
const SourceUnavailable = "source unavailable"

// SourceWindow is the default number of context lines added on each side of a
// locator's cited range (owner decision Q2: capped ±40).
const SourceWindow = 40

// Excerpt reads the lines a locator cites from a local checkout, widened by
// window lines on each side and clamped to the file. It returns (excerpt, true),
// or ("", false) when the checkout, the file, the range is missing, or the path
// escapes the checkout. ADR 0001: the index keeps locators only; this reads
// source at render time and never persists it.
func Excerpt(repoRoot string, loc domain.Locator, window int) (string, bool) {
	if repoRoot == "" || loc.Path == "" {
		return "", false
	}
	clean := filepath.Clean(filepath.FromSlash(loc.Path))
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return "", false
	}
	lines, err := readLines(filepath.Join(repoRoot, clean))
	if err != nil {
		return "", false
	}
	start := loc.StartLine - window
	if start < 1 {
		start = 1
	}
	end := loc.EndLine + window
	if end < start {
		end = start
	}
	if start > len(lines) {
		return "", false
	}
	if end > len(lines) {
		end = len(lines)
	}
	var sb strings.Builder
	for i := start; i <= end; i++ {
		fmt.Fprintf(&sb, "%5d  %s\n", i, lines[i-1])
	}
	return sb.String(), true
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		out = append(out, sc.Text())
	}
	return out, sc.Err()
}
