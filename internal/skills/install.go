package skills

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Hosts are the supported host families (spec A.3).
var Hosts = []string{"claude", "cursor", "agents"}

// Options configures an install.
type Options struct {
	// Root is the project directory the host files are written under.
	Root string
	// Command is how the host should start the MCP server (default "gira").
	Command string
	// Index, when set, is passed to `gira serve mcp --index`.
	Index string
	// Hook, for the claude host, registers the opt-in PreToolUse nudge.
	Hook bool
}

// Report lists what an install did, by path relative to Root.
type Report struct {
	Host      string   `json:"host"`
	Created   []string `json:"created"`
	Updated   []string `json:"updated"`
	Unchanged []string `json:"unchanged"`
}

// planFiles builds the expected host file renderers from the canonical
// catalog. It is shared by Install (which writes) and Check (which only
// reads), so the two can never disagree on what "in sync" means.
func planFiles(host string, opts Options) (map[string]func([]byte) ([]byte, error), error) {
	skills, err := Catalog()
	if err != nil {
		return nil, err
	}
	if opts.Command == "" {
		opts.Command = "gira"
	}
	files := map[string]func([]byte) ([]byte, error){}
	switch host {
	case "claude":
		for _, s := range skills {
			files[filepath.Join(".claude", "skills", s.Name, "SKILL.md")] = skillFile(s)
		}
		files[".mcp.json"] = mcpEntry(opts, map[string]any{"type": "stdio"})
		if opts.Hook {
			files[filepath.Join(".claude", "settings.json")] = nudgeHook(opts)
		}
	case "agents":
		for _, s := range skills {
			files[filepath.Join(".agents", "skills", s.Name, "SKILL.md")] = skillFile(s)
		}
	case "cursor":
		files[filepath.Join(".cursor", "rules", "gira.mdc")] = cursorRules(skills)
		files[filepath.Join(".cursor", "mcp.json")] = mcpEntry(opts, nil)
	default:
		return nil, fmt.Errorf("unknown host %q: want one of %s", host, strings.Join(Hosts, ", "))
	}
	return files, nil
}

// Install generates the host files for host from the canonical catalog.
func Install(host string, opts Options) (*Report, error) {
	if opts.Root == "" {
		opts.Root = "."
	}
	files, err := planFiles(host, opts)
	if err != nil {
		return nil, err
	}

	rep := &Report{Host: host, Created: []string{}, Updated: []string{}, Unchanged: []string{}}
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, rel := range paths {
		full := filepath.Join(opts.Root, rel)
		old, err := os.ReadFile(full)
		exists := err == nil
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		next, err := files[rel](old)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", rel, err)
		}
		switch {
		case exists && bytes.Equal(old, next):
			rep.Unchanged = append(rep.Unchanged, filepath.ToSlash(rel))
			continue
		case exists:
			rep.Updated = append(rep.Updated, filepath.ToSlash(rel))
		default:
			rep.Created = append(rep.Created, filepath.ToSlash(rel))
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(full, next, 0o644); err != nil {
			return nil, err
		}
	}
	return rep, nil
}

// skillFile renders a SKILL.md as two owned regions: the frontmatter fields (YAML
// comment markers inside the frontmatter) and the body (HTML comment markers).
func skillFile(s Skill) func([]byte) ([]byte, error) {
	front := region{id: "frontmatter", open: "# ", content: s.Frontmatter}
	body := region{id: "body", open: "<!-- ", close: " -->", content: strings.TrimLeft(s.Body, "\n")}
	return func(old []byte) ([]byte, error) {
		return applyRegions(old, func() string {
			return "---\n" + front.render() + "---\n" + body.render()
		}, []region{front, body})
	}
}

// cursorRules renders every skill into one Cursor rules file.
func cursorRules(skills []Skill) func([]byte) ([]byte, error) {
	var names []string
	var b strings.Builder
	for _, s := range skills {
		names = append(names, s.Name)
		b.WriteString(fmt.Sprintf("## Skill: %s\n\nUse when: %s\n\n", s.Name, s.Description))
		body := strings.TrimLeft(s.Body, "\n")
		// Demote the skill's own headings one level under its "## Skill" heading.
		for _, l := range strings.SplitAfter(body, "\n") {
			if strings.HasPrefix(l, "#") {
				l = "#" + l
			}
			b.WriteString(l)
		}
		b.WriteString("\n")
	}
	front := region{id: "frontmatter", open: "# ", content: fmt.Sprintf("description: GIRA architecture skills (%s). Deterministic, evidence-backed answers from the gira MCP server or CLI.\nalwaysApply: false\n", strings.Join(names, ", "))}
	body := region{id: "skills", open: "<!-- ", close: " -->", content: b.String()}
	return func(old []byte) ([]byte, error) {
		return applyRegions(old, func() string {
			return "---\n" + front.render() + "---\n" + body.render()
		}, []region{front, body})
	}
}

// mcpEntry registers `gira serve mcp` under mcpServers.gira, the only key the
// installer owns in an MCP config file.
func mcpEntry(opts Options, extra map[string]any) func([]byte) ([]byte, error) {
	args := []string{"serve", "mcp"}
	if opts.Index != "" {
		args = append(args, "--index", opts.Index)
	}
	entry := map[string]any{"command": opts.Command, "args": args}
	for k, v := range extra {
		entry[k] = v
	}
	return func(old []byte) ([]byte, error) {
		return setJSONKey(old, []string{"mcpServers", "gira"}, entry)
	}
}

// NudgeCommand marks the hook entry the installer owns in .claude/settings.json.
const NudgeCommand = "skills nudge"

// nudgeHook appends one PreToolUse entry that runs `gira skills nudge` before
// edits. Entries already running the nudge are left alone; nothing else in the
// settings file changes.
func nudgeHook(opts Options) func([]byte) ([]byte, error) {
	cmd := opts.Command + " " + NudgeCommand
	if opts.Index != "" {
		cmd += " --index " + opts.Index
	}
	entry := map[string]any{
		"matcher": "Edit|MultiEdit|Write",
		"hooks":   []any{map[string]any{"type": "command", "command": cmd}},
	}
	return func(old []byte) ([]byte, error) {
		return appendJSONArrayItem(old, []string{"hooks", "PreToolUse"}, entry, func(el any) bool {
			m, _ := el.(map[string]any)
			hooks, _ := m["hooks"].([]any)
			for _, h := range hooks {
				if hm, ok := h.(map[string]any); ok && strings.Contains(fmt.Sprint(hm["command"]), NudgeCommand) {
					return true
				}
			}
			return false
		})
	}
}
