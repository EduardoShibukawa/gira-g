package skills

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// Nudge reads a Claude Code PreToolUse hook payload and, when the file about to be
// edited is cited by the index, writes a hook response that adds context
// suggesting gira-impact-briefing. It never blocks: it sets no permission
// decision, and on any problem it writes nothing. repoRoot makes the edited path
// relative, to compare it with locator paths.
func Nudge(payload io.Reader, out io.Writer, mm *domain.MacroMap, repoRoot string) error {
	var in struct {
		ToolInput struct {
			FilePath string `json:"file_path"`
		} `json:"tool_input"`
	}
	if err := json.NewDecoder(payload).Decode(&in); err != nil || in.ToolInput.FilePath == "" || mm == nil {
		return nil
	}
	p := in.ToolInput.FilePath
	// A slash-rooted path is absolute on every OS; filepath.IsAbs alone does
	// not report it as such on Windows (no volume), which skipped the
	// relativization below and silenced the nudge. No-op on Linux.
	if (filepath.IsAbs(p) || strings.HasPrefix(p, "/")) && repoRoot != "" {
		if rel, err := filepath.Rel(repoRoot, p); err == nil && !strings.HasPrefix(rel, "..") {
			p = rel
		}
	}
	p = filepath.ToSlash(p)
	ids := map[string]bool{}
	for _, n := range mm.Nodes {
		for _, ev := range n.Evidence {
			if ev.Locator.Path == p {
				ids[n.ID] = true
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	list := make([]string, 0, len(ids))
	for id := range ids {
		list = append(list, id)
	}
	sort.Strings(list)
	if len(list) > 5 {
		list = append(list[:5], fmt.Sprintf("and %d more", len(ids)-5))
	}
	msg := fmt.Sprintf("GIRA maps %s to %s. Consider the gira-impact-briefing skill (gira impact <id> --budget 1500) before editing.", p, strings.Join(list, ", "))
	return json.NewEncoder(out).Encode(map[string]any{
		"hookSpecificOutput": map[string]any{"hookEventName": "PreToolUse", "additionalContext": msg},
	})
}
