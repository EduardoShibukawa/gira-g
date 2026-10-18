package windowsgates

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestTabFilenameConstruction diagnoses whether a tab-named entry can exist in
// a repository on this OS, via the index and via mktree, and whether it
// survives `clone --bare` (the shape of the gira cache). A tab name is legal
// in a Git tree and on Linux/macOS filesystems, not on Windows filesystems.
func TestTabFilenameConstruction(t *testing.T) {
	_, env := isolatedHome(t, true)
	cenv := append(append([]string{}, env...), identity...)
	const name = "a\tb.proto"

	snapshot := func(method, dir string, f map[string]any) {
		lt := git(dir, env, "ls-tree", "-r", "-z", "HEAD")
		f["ls_tree_r_z_exit"], f["ls_tree_r_z"] = lt.Exit, lt.Stdout
		bare := filepath.Join(t.TempDir(), "bare.git")
		cl := git("", env, "clone", "-q", "--bare", dir, bare)
		f["bare_clone_exit"], f["bare_clone_stderr"] = cl.Exit, trim(cl.Stderr)
		if cl.Exit == 0 {
			lb := git(bare, env, "ls-tree", "-r", "-z", "HEAD")
			f["bare_ls_tree_r_z_exit"], f["bare_ls_tree_r_z"] = lb.Exit, lb.Stdout

			// Test git archive piped to tar -x (ExportTree contract in internal/gitcache/cache.go)
			targetDir := filepath.Join(t.TempDir(), "target")
			_ = os.MkdirAll(targetDir, 0o755)
			e1, e2, se1, se2 := runPipe(bare, env, targetDir, "git", []string{"archive", "HEAD"}, "tar", []string{"-x", "-C", targetDir})
			f["archive_exit"] = e1
			f["archive_stderr"] = trim(se1)
			f["tar_extract_exit"] = e2
			f["tar_extract_stderr"] = trim(se2)
			entries, _ := os.ReadDir(targetDir)
			var extracted []string
			for _, en := range entries {
				extracted = append(extracted, en.Name())
			}
			f["extracted_files"] = extracted
		}
		fs := git(dir, env, "fsck", "--strict")
		f["fsck_exit"], f["fsck_stderr"] = fs.Exit, trim(fs.Stderr)
		f["method"] = method
		rec(t, "tab_filename", f)
	}

	// Method 1: index (update-index --index-info), no worktree file.
	d1 := filepath.Join(t.TempDir(), "idx")
	f1 := map[string]any{}
	oid := ""
	func() {
		if r := git(t.TempDir(), env, "init", "-q", "-b", "main", d1); r.Exit != 0 {
			f1["init_failed"] = r.Stderr
			return
		}
		r := run(d1, []byte("x\n"), env, "git", "hash-object", "-w", "--stdin")
		oid = trim(r.Stdout)
		var idx bytes.Buffer
		idx.WriteString("100644 " + oid + "\t" + name + "\x00")
		ui := run(d1, idx.Bytes(), env, "git", "update-index", "-z", "--add", "--index-info")
		f1["update_index_exit"], f1["update_index_stderr"] = ui.Exit, trim(ui.Stderr)
		fz := git(d1, env, "ls-files", "-z", "-s")
		f1["ls_files_z_s"] = fz.Stdout
		cm := git(d1, cenv, "commit", "-q", "--no-gpg-sign", "-m", "x")
		f1["commit_exit"], f1["commit_stderr"] = cm.Exit, trim(cm.Stderr)
		snapshot("update-index --index-info -z", d1, f1)
	}()

	// Method 2: mktree -z + commit-tree, bypassing index path verification.
	d2 := filepath.Join(t.TempDir(), "mk")
	f2 := map[string]any{}
	if r := git(t.TempDir(), env, "init", "-q", "-b", "main", d2); r.Exit != 0 {
		f2["init_failed"] = r.Stderr
		rec(t, "tab_filename", map[string]any{"method": "mktree", "init_failed": r.Stderr})
		return
	}
	r := run(d2, []byte("x\n"), env, "git", "hash-object", "-w", "--stdin")
	oid2 := trim(r.Stdout)
	mt := run(d2, []byte("100644 blob "+oid2+"\t"+name+"\x00"), env, "git", "mktree", "-z")
	f2["mktree_exit"], f2["mktree_stderr"] = mt.Exit, trim(mt.Stderr)
	if mt.Exit == 0 {
		ct := run(d2, nil, cenv, "git", "commit-tree", "-m", "x", trim(mt.Stdout))
		f2["commit_tree_exit"], f2["commit_tree_stderr"] = ct.Exit, trim(ct.Stderr)
		if ct.Exit == 0 {
			git(d2, env, "update-ref", "refs/heads/main", trim(ct.Stdout))
			snapshot("mktree -z + commit-tree", d2, f2)
			return
		}
	}
	rec(t, "tab_filename", map[string]any{"method": "mktree -z + commit-tree", "detail": f2})
}
