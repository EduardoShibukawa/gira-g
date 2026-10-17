package windowsgates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	nonASCIIName = "proto/café.proto"
	tabName      = "proto/a\tb.proto"
)

// TestDevNullAsGlobalConfig records how git reacts to GIT_CONFIG_GLOBAL
// pointing at the null device, which is what the product's read path will use.
func TestDevNullAsGlobalConfig(t *testing.T) {
	// Fixture repo + bare cache built with an isolated, empty user config.
	_, setupEnv := isolatedHome(t, true)
	src := filepath.Join(t.TempDir(), "src")
	commit := makeRepo(t, src, setupEnv, []treeEntry{
		{"100644", nonASCIIName, "syntax = \"proto3\";\n"},
		{"100644", tabName, "syntax = \"proto3\";\n// tab\n"},
		{"100644", "plain.proto", "syntax = \"proto3\";\n// plain\n"},
	})
	cache := filepath.Join(t.TempDir(), "cache.git")
	if r := git("", setupEnv, "clone", "-q", "--bare", src, cache); r.Exit != 0 {
		t.Fatalf("bare clone: %+v", r)
	}

	// Each variant gets its own "user" home holding a global config that would
	// change read output, so the write probe of one variant cannot leak into another.
	userCfgBody := []byte("[core]\n\tquotepath = false\n\tsymlinks = true\n[x]\n\ty = from-user\n")
	userEnv := func(home string, extra ...string) []string {
		return append([]string{
			"HOME=" + home, "USERPROFILE=" + home,
			"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"),
			"GIT_TERMINAL_PROMPT=0",
		}, extra...)
	}

	type variant struct {
		name  string
		extra func(t *testing.T) []string
	}
	lit := func(kv ...string) func(*testing.T) []string { return func(*testing.T) []string { return kv } }
	variants := []variant{
		// What the product will set (plan 2b): os.DevNull + NOSYSTEM.
		{"os.DevNull+NOSYSTEM", lit("GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")},
		// What internal/corpus/corpus.go:26-27 sets today, on every OS.
		{"literal_/dev/null+SYSTEM=/dev/null", lit("GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")},
		{"os.DevNull_only_no_NOSYSTEM", lit("GIT_CONFIG_GLOBAL=" + os.DevNull)},
		// Control: user config is live.
		{"control_user_config_live", lit("GIT_CONFIG_NOSYSTEM=1")},
		// Control: explicit empty file instead of the null device.
		{"empty_regular_file+NOSYSTEM", func(t *testing.T) []string {
			return []string{"GIT_CONFIG_GLOBAL=" + mkEmpty(t), "GIT_CONFIG_NOSYSTEM=1"}
		}},
	}
	if os.DevNull != "NUL" {
		// Extra comparison only meaningful off Windows: the literal spelling NUL.
		variants = append(variants, variant{"literal_NUL+NOSYSTEM", lit("GIT_CONFIG_GLOBAL=NUL", "GIT_CONFIG_NOSYSTEM=1")})
	}

	blobR := git(cache, setupEnv, "rev-parse", "HEAD:plain.proto")
	blob := trim(blobR.Stdout)

	cmds := []struct {
		name string
		args []string
	}{
		{"config_global_list", []string{"config", "--global", "--list"}},
		{"config_global_get_core.quotepath", []string{"config", "--global", "--get", "core.quotepath"}},
		{"config_effective_origin_core.quotepath", []string{"config", "--show-origin", "--get-all", "core.quotepath"}},
		{"config_effective_origin_x.y", []string{"config", "--show-origin", "--get-all", "x.y"}},
		{"ls-tree_-z_HEAD", []string{"ls-tree", "-z", "HEAD"}},
		{"ls-tree_-r_-z_--full-name_HEAD", []string{"ls-tree", "-r", "-z", "--full-name", "HEAD"}},
		{"ls-tree_-r_-z_--full-name_HEAD_proto", []string{"ls-tree", "-r", "-z", "--full-name", "HEAD", "proto/"}},
		{"cat-file_-p_blob_oid", []string{"cat-file", "-p", blob}},
		{"cat-file_-p_HEAD:nonascii", []string{"cat-file", "-p", "HEAD:" + nonASCIIName}},
		{"cat-file_-p_HEAD:tab", []string{"cat-file", "-p", "HEAD:" + tabName}},
		{"ls-tree_-r_HEAD(no_-z)", []string{"ls-tree", "-r", "HEAD"}},
		{"-c_core.quotePath=true_ls-tree_-r_HEAD", []string{"-c", "core.quotePath=true", "ls-tree", "-r", "HEAD"}},
		{"-c_core.quotePath=false_ls-tree_-r_HEAD", []string{"-c", "core.quotePath=false", "ls-tree", "-r", "HEAD"}},
		{"config_global_write_x.y=z", []string{"config", "--global", "x.y", "z"}},
	}

	for _, v := range variants {
		home := t.TempDir()
		if err := os.WriteFile(filepath.Join(home, ".gitconfig"), userCfgBody, 0o644); err != nil {
			t.Fatal(err)
		}
		venv := userEnv(home, v.extra(t)...)
		// Taken before the write probe, which mutates config. The property
		// the product needs: user config is NOT visible, bytes are the same.
		eff := git(cache, venv, "config", "--get-all", "x.y")
		lt := git(cache, venv, "ls-tree", "-r", "-z", "--full-name", "HEAD")
		rec(t, "devnull_summary", map[string]any{
			"variant": v.name, "commit": commit,
			"user_config_visible": strings.Contains(eff.Stdout, "from-user"), "ls_tree_z_exit": lt.Exit, "ls_tree_z": lt.Stdout,
		})
		for _, c := range cmds {
			r := git(cache, venv, c.args...)
			rec(t, "devnull_git", map[string]any{
				"variant": v.name, "cmd": c.name, "args": c.args,
				"exit": r.Exit, "stdout": r.Stdout, "stderr": trim(r.Stderr), "start_error": r.Err,
			})
		}
	}

	// Non-ASCII/tab names must survive on the filesystem-free path: also try
	// a real clone (read path of cache creation keeps the user config).
	cloneHome := t.TempDir()
	if err := os.WriteFile(filepath.Join(cloneHome, ".gitconfig"), userCfgBody, 0o644); err != nil {
		t.Fatal(err)
	}
	cl := git("", userEnv(cloneHome, "GIT_CONFIG_NOSYSTEM=1"), "clone", "-q", "--bare", src, filepath.Join(t.TempDir(), "c2.git"))
	rec(t, "clone_with_user_config", map[string]any{"exit": cl.Exit, "stderr": trim(cl.Stderr)})
}

func mkEmpty(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "empty.gitconfig")
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}
