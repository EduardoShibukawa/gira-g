// Package human renders the human-facing terminal surface. It keeps the exact
// plain text the commands already emit and adds ANSI styling only when stdout is
// a terminal and NO_COLOR is unset, so pipes, CI and goldens see the
// byte-identical plain output (ADR 0008 §1).
package human

import (
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Styled reports whether w should receive ANSI styling. It is false when
// NO_COLOR is set (the cross-tool standard: any value, even empty) or when w is
// not a character device (pipes, buffers, CI).
func Styled(w io.Writer) bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// Decorate styles Markdown lines for a terminal: headings bold, unsupported
// warnings yellow. With styled=false it returns s unchanged (byte-identical
// plain output). The styled path forces a color profile because the caller has
// already decided the target is a color terminal.
func Decorate(s string, styled bool) string {
	if !styled {
		return s
	}
	r := lipgloss.NewRenderer(io.Discard)
	r.SetColorProfile(termenv.ANSI256)
	heading := r.NewStyle().Bold(true)
	warning := r.NewStyle().Foreground(lipgloss.Color("3"))
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		switch {
		case strings.Contains(l, "⚠️"):
			lines[i] = warning.Render(l)
		case strings.HasPrefix(l, "#"):
			lines[i] = heading.Render(l)
		}
	}
	return strings.Join(lines, "\n")
}
