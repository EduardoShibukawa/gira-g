package skills

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// SC10: every canonical SKILL.md declares a trigger description, its operations,
// its output contract and its budget.
func TestCatalogDeclaresContract(t *testing.T) {
	cat, err := Catalog()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"gira-arch-diff", "gira-context", "gira-explain-component", "gira-impact-briefing", "gira-query", "gira-tour"}
	var got []string
	for _, s := range cat {
		got = append(got, s.Name)
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("catalog %v, want %v", got, want)
	}
}

func TestParseRejectsIncompleteSkill(t *testing.T) {
	_, err := Parse([]byte("---\nname: x\ndescription: y\n---\n# x\n"))
	if err == nil || !strings.Contains(err.Error(), "operations") || !strings.Contains(err.Error(), "## Budget") {
		t.Fatalf("incomplete skill accepted or error unclear: %v", err)
	}
}

func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			data, _ := os.ReadFile(p)
			rel, _ := filepath.Rel(root, p)
			out[filepath.ToSlash(rel)] = string(data)
		}
		return nil
	})
	return out
}

// SC10: each host is generated from the single canonical source, a second run
// changes nothing, and only content inside the markers (or the owned JSON key)
// changes when a user has edited the files.
func TestInstallIsIdempotentAndMarkerScoped(t *testing.T) {
	cat, err := Catalog()
	if err != nil {
		t.Fatal(err)
	}
	for _, host := range Hosts {
		t.Run(host, func(t *testing.T) {
			root := t.TempDir()
			rep, err := Install(host, Options{Root: root})
			if err != nil {
				t.Fatal(err)
			}
			if len(rep.Created) == 0 || len(rep.Updated)+len(rep.Unchanged) != 0 {
				t.Fatalf("first run: %+v", rep)
			}
			first := snapshot(t, root)

			// Generated from the canonical source: every skill body is present.
			all := ""
			for _, c := range first {
				all += c
			}
			for _, s := range cat {
				line := strings.SplitN(strings.TrimLeft(s.Body, "\n"), "\n", 3)[2]
				if !strings.Contains(all, strings.TrimSpace(strings.SplitN(line, "\n", 2)[0])) {
					t.Errorf("%s: body of %s not generated", host, s.Name)
				}
			}
			if host != "cursor" {
				for _, s := range cat {
					p := filepath.Join(root, "."+map[string]string{"claude": "claude", "agents": "agents"}[host], "skills", s.Name, "SKILL.md")
					data, err := os.ReadFile(p)
					if err != nil {
						t.Fatal(err)
					}
					got, err := Parse(data)
					if err != nil {
						t.Fatalf("%s: generated SKILL.md does not parse: %v", p, err)
					}
					if got.Name != s.Name || got.Description != s.Description || got.Budget == nil {
						t.Fatalf("%s: frontmatter drifted from the canonical source", p)
					}
				}
			}

			rep, err = Install(host, Options{Root: root})
			if err != nil {
				t.Fatal(err)
			}
			if len(rep.Created)+len(rep.Updated) != 0 {
				t.Fatalf("second run changed files: %+v", rep)
			}
			if !equalSnap(first, snapshot(t, root)) {
				t.Fatal("second run changed bytes")
			}

			// A user edits outside the markers and tampers inside them.
			for rel, content := range first {
				p := filepath.Join(root, rel)
				var edited string
				if strings.HasSuffix(rel, ".json") {
					var doc map[string]any
					_ = json.Unmarshal([]byte(content), &doc)
					doc["userSetting"] = true
					doc["mcpServers"].(map[string]any)["other"] = map[string]any{"command": "other"}
					doc["mcpServers"].(map[string]any)["gira"] = map[string]any{"command": "tampered"}
					b, _ := json.MarshalIndent(doc, "", "    ")
					edited = string(b) + "\n"
				} else {
					edited = "USER PREAMBLE\n" + strings.Replace(content, "## Operations", "## Tampered", 1) + "USER NOTES\n"
				}
				if err := os.WriteFile(p, []byte(edited), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := Install(host, Options{Root: root}); err != nil {
				t.Fatal(err)
			}
			for rel, content := range snapshot(t, root) {
				if strings.HasSuffix(rel, ".json") {
					var doc map[string]any
					if err := json.Unmarshal([]byte(content), &doc); err != nil {
						t.Fatal(err)
					}
					servers := doc["mcpServers"].(map[string]any)
					if doc["userSetting"] != true || servers["other"] == nil || servers["gira"].(map[string]any)["command"] != "gira" {
						t.Fatalf("%s: JSON edit not scoped to mcpServers.gira:\n%s", rel, content)
					}
					if !strings.Contains(content, "    \"userSetting\": true") {
						t.Fatalf("%s: user formatting outside the owned key changed:\n%s", rel, content)
					}
					continue
				}
				if !strings.HasPrefix(content, "USER PREAMBLE\n") || !strings.HasSuffix(content, "USER NOTES\n") {
					t.Fatalf("%s: content outside the markers was not kept", rel)
				}
				if strings.Contains(content, "## Tampered") {
					t.Fatalf("%s: content inside the markers was not restored", rel)
				}
			}
		})
	}
}

func equalSnap(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func TestSetJSONKeyInsertsWithoutTouchingOtherBytes(t *testing.T) {
	doc := []byte("{\n\t\"a\": [1,2],\n\t\"mcpServers\": {\n\t\t\"x\": {\"command\": \"x\"}\n\t}\n}\n")
	out, err := setJSONKey(doc, []string{"mcpServers", "gira"}, map[string]any{"command": "gira"})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(out, []byte("{\n\t\"a\": [1,2],\n\t\"mcpServers\": {\n\t\t\"x\": {\"command\": \"x\"},\n\t\t\"gira\": {")) {
		t.Fatalf("unexpected splice:\n%s", out)
	}
	var v map[string]any
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("invalid JSON after splice: %v\n%s", err, out)
	}
	again, _ := setJSONKey(out, []string{"mcpServers", "gira"}, map[string]any{"command": "gira"})
	if !bytes.Equal(again, out) {
		t.Fatal("setting an equal value must not change the document")
	}
}

func TestHookIsOptInAppendOnlyAndIdempotent(t *testing.T) {
	root := t.TempDir()
	if _, err := Install("claude", Options{Root: root}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude", "settings.json")); err == nil {
		t.Fatal("the hook must be opt-in")
	}
	settings := filepath.Join(root, ".claude", "settings.json")
	user := "{\n    \"permissions\": {\"allow\": [\"Bash(ls)\"]},\n    \"hooks\": {\n        \"PreToolUse\": [\n            {\"matcher\": \"Bash\", \"hooks\": [{\"type\": \"command\", \"command\": \"mine\"}]}\n        ]\n    }\n}\n"
	if err := os.WriteFile(settings, []byte(user), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install("claude", Options{Root: root, Hook: true}); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(settings)
	var doc struct {
		Permissions map[string]any `json:"permissions"`
		Hooks       struct {
			PreToolUse []map[string]any `json:"PreToolUse"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(first, &doc); err != nil {
		t.Fatalf("invalid settings: %v\n%s", err, first)
	}
	if len(doc.Hooks.PreToolUse) != 2 || doc.Permissions == nil || !strings.HasPrefix(string(first), "{\n    \"permissions\": {\"allow\": [\"Bash(ls)\"]},") {
		t.Fatalf("hook not appended next to the user's entries:\n%s", first)
	}
	rep, err := Install("claude", Options{Root: root, Hook: true})
	if err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(settings)
	if !bytes.Equal(first, second) || len(rep.Updated) != 0 {
		t.Fatalf("second hook install changed settings: %+v", rep)
	}
}

func TestNudgeNeverBlocks(t *testing.T) {
	mm := &domain.MacroMap{Nodes: []domain.Node{{ID: "service:cart", Evidence: []domain.EvidenceRecord{{Locator: domain.Locator{Path: "src/cart/main.go"}}}}}}
	var out bytes.Buffer
	if err := Nudge(strings.NewReader(`{"tool_input":{"file_path":"/repo/src/cart/main.go"}}`), &out, mm, "/repo"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "additionalContext") || !strings.Contains(out.String(), "service:cart") || strings.Contains(out.String(), "permissionDecision") {
		t.Fatalf("nudge output: %s", out.String())
	}
	out.Reset()
	_ = Nudge(strings.NewReader(`{"tool_input":{"file_path":"/repo/README.md"}}`), &out, mm, "/repo")
	_ = Nudge(strings.NewReader(`not json`), &out, mm, "/repo")
	if out.Len() != 0 {
		t.Fatalf("unmapped or malformed input must stay silent: %s", out.String())
	}
}

func TestQuestionRouteDocFlowsAndBudget(t *testing.T) {
	tool, args, cli, err := (Question{Op: "doc", ServiceID: "service:acme", Budget: 800}).route("idx.json")
	if err != nil {
		t.Fatal(err)
	}
	if tool != "get_doc" || args["service_id"] != "service:acme" || args["budget"] != 800 {
		t.Fatalf("unexpected doc route: tool=%s args=%v", tool, args)
	}
	wantCLI := "doc service:acme --budget 800 --index idx.json --format json"
	if strings.Join(cli, " ") != wantCLI {
		t.Fatalf("doc cli = %q, want %q", strings.Join(cli, " "), wantCLI)
	}

	tool, args, cli, err = (Question{Op: "flows", Entry: "service:frontend", Budget: 600}).route("idx.json")
	if err != nil {
		t.Fatal(err)
	}
	if tool != "get_flows" || args["entry"] != "service:frontend" || args["budget"] != 600 {
		t.Fatalf("unexpected flows route: tool=%s args=%v", tool, args)
	}
	wantFlowsCLI := "flows service:frontend --budget 600 --index idx.json --format json"
	if strings.Join(cli, " ") != wantFlowsCLI {
		t.Fatalf("flows cli = %q, want %q", strings.Join(cli, " "), wantFlowsCLI)
	}
}

