package windowsgates

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGitSymlinkCheckout commits a symlink (mode 120000) and clones it with
// the default configuration to see what materializes. The commit is built
// through the index so it does not depend on the filesystem supporting links.
func TestGitSymlinkCheckout(t *testing.T) {
	_, env := isolatedHome(t, false) // system config stays: it is what a user's git has
	src := filepath.Join(t.TempDir(), "src")
	makeRepo(t, src, env, []treeEntry{
		{"100644", "target.txt", "payload\n"},
		{"120000", "link_to_file", "target.txt"},
		{"120000", "link_to_dir", "dir"},
		{"100644", "dir/inner.txt", "inner\n"},
	})

	for _, variant := range []struct {
		name string
		args []string
	}{
		{"default", nil},
		{"core.symlinks=true", []string{"-c", "core.symlinks=true"}},
		{"core.symlinks=false", []string{"-c", "core.symlinks=false"}},
	} {
		dst := filepath.Join(t.TempDir(), "clone")
		args := append(append([]string{}, variant.args...), "clone", "-q", src, dst)
		r := git("", env, args...)
		f := map[string]any{"variant": variant.name, "clone_exit": r.Exit, "clone_stderr": trim(r.Stderr)}
		if r.Exit == 0 {
			cs := git(dst, env, "config", "--local", "--get", "core.symlinks")
			f["clone_core_symlinks"] = map[string]any{"exit": cs.Exit, "value": trim(cs.Stdout)}
			ls := git(dst, env, "ls-files", "-s")
			f["ls_files_s"] = trim(ls.Stdout)
			for _, n := range []string{"link_to_file", "link_to_dir"} {
				p := filepath.Join(dst, n)
				e := map[string]any{}
				if li, err := os.Lstat(p); err != nil {
					e["lstat_error"] = err.Error()
				} else {
					e["is_symlink"], e["mode"], e["is_dir"] = li.Mode()&os.ModeSymlink != 0, li.Mode().String(), li.IsDir()
					if li.Mode()&os.ModeSymlink != 0 {
						rl, err := os.Readlink(p)
						e["readlink"], e["readlink_error"] = rl, errStr(err)
					} else if !li.IsDir() {
						b, _ := os.ReadFile(p)
						e["regular_file_content"] = string(b)
					}
				}
				f[n] = e
			}
			st := git(dst, env, "status", "--porcelain")
			f["status_porcelain_exit"], f["status_porcelain"] = st.Exit, trim(st.Stdout)
		}
		rec(t, "git_symlink_checkout", f)
	}

	// Reads that gira would do on a bare cache: no checkout involved.
	bare := filepath.Join(t.TempDir(), "bare.git")
	if r := git("", env, "clone", "-q", "--bare", src, bare); r.Exit != 0 {
		t.Fatalf("bare clone: %+v", r)
	}
	lt := git(bare, env, "ls-tree", "-r", "-z", "--full-name", "HEAD")
	cf := git(bare, env, "cat-file", "-p", "HEAD:link_to_file")
	rec(t, "git_symlink_bare_read", map[string]any{
		"ls_tree_exit": lt.Exit, "ls_tree_z": lt.Stdout, "ls_tree_stderr": trim(lt.Stderr),
		"cat_file_link_exit": cf.Exit, "cat_file_link_stdout": cf.Stdout,
	})
}
