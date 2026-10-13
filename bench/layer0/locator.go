package layer0

import (
	"fmt"
	"strings"
)

// FileReader reads one corpus file at the pinned SHA.
type FileReader func(path string) ([]byte, error)

// CheckLocator is Layer 0's locator accuracy rule: the cited file exists at the
// pinned SHA, the range is in bounds, and, when the ground truth records an
// anchor for the fact, the cited range contains it.
func CheckLocator(read FileReader, ev Evidence, anchor string) error {
	data, err := read(ev.Path)
	if err != nil {
		return fmt.Errorf("%s: file does not exist at the pinned SHA", ev.Path)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if ev.Start < 1 || ev.End < ev.Start || ev.End > len(lines) {
		return fmt.Errorf("%s#L%d-L%d: out of bounds (file has %d lines)", ev.Path, ev.Start, ev.End, len(lines))
	}
	if anchor != "" && !strings.Contains(strings.Join(lines[ev.Start-1:ev.End], "\n"), anchor) {
		return fmt.Errorf("%s#L%d-L%d: does not contain anchor %q", ev.Path, ev.Start, ev.End, anchor)
	}
	return nil
}
