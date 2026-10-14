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
// or ("", false) when the checkout, the file or the range is missing, when the
// path escapes the checkout (resolving symlinks), or when the target is not a
// regular file. It reads at most the lines it prints. ADR 0001: the index keeps
// locators only; this reads source at render time and never persists it.
func Excerpt(repoRoot string, loc domain.Locator, window int) (string, bool) {
	if repoRoot == "" || loc.Path == "" {
		return "", false
	}
	root, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		return "", false
	}
	clean := filepath.Clean(filepath.FromSlash(loc.Path))
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", false
	}
	full, err := filepath.EvalSymlinks(filepath.Join(root, clean))
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(root, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	if fi, err := os.Stat(full); err != nil || !fi.Mode().IsRegular() {
		return "", false
	}

	start := loc.StartLine - window
	if start < 1 {
		start = 1
	}
	end := loc.EndLine
	if end == 0 {
		end = loc.StartLine
	}
	end += window
	if end < start {
		end = start
	}

	f, err := os.Open(full)
	if err != nil {
		return "", false
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, 64*1024)
	var sb strings.Builder
	for n := 1; n <= end; n++ {
		line, err := r.ReadString('\n')
		if line == "" && err != nil {
			break
		}
		if n >= start {
			fmt.Fprintf(&sb, "%5d  %s", n, line)
			if !strings.HasSuffix(line, "\n") {
				sb.WriteString("\n")
			}
		}
		if err != nil {
			break
		}
	}
	if sb.Len() == 0 {
		return "", false
	}
	return sb.String(), true
}
