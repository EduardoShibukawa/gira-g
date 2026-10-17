package windowsgates

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func modeStr(fi os.FileInfo) string { return fi.Mode().String() }

func readThrough(p string) map[string]any {
	f, err := os.Open(p)
	if err != nil {
		return map[string]any{"open_error": err.Error()}
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	return map[string]any{"content": string(b), "read_error": errStr(err)}
}

// TestSymlinkOS exercises os-level symlink behaviour: creation, resolution,
// Stat/Lstat, identity and read-through.
func TestSymlinkOS(t *testing.T) {
	root := t.TempDir()
	targetFile := filepath.Join(root, "target.txt")
	targetDir := filepath.Join(root, "tdir")
	if err := os.WriteFile(targetFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "inner.txt"), []byte("inner"), 0o644); err != nil {
		t.Fatal(err)
	}
	tfi, _ := os.Stat(targetFile)
	tdi, _ := os.Stat(targetDir)

	cases := []struct {
		name, link, target string
		isDir              bool
	}{
		{"file_relative", "flink_rel", "target.txt", false},
		{"file_absolute", "flink_abs", targetFile, false},
		{"dir_relative", "dlink_rel", "tdir", true},
		{"dir_absolute", "dlink_abs", targetDir, true},
	}
	for _, c := range cases {
		link := filepath.Join(root, c.link)
		f := map[string]any{"case": c.name, "target_arg": c.target}
		err := os.Symlink(c.target, link)
		f["symlink_error"] = errStr(err)
		if err != nil {
			f["created"] = false
			rec(t, "symlink_os", f)
			continue
		}
		f["created"] = true
		rl, err := os.Readlink(link)
		f["readlink"], f["readlink_error"] = rl, errStr(err)
		ev, err := filepath.EvalSymlinks(link)
		f["evalsymlinks_equals_expected"] = err == nil && sameString(ev, targetFile, targetDir, c.isDir)
		f["evalsymlinks"], f["evalsymlinks_error"] = ev, errStr(err)
		li, lerr := os.Lstat(link)
		si, serr := os.Stat(link)
		f["lstat_error"], f["stat_error"] = errStr(lerr), errStr(serr)
		if lerr == nil {
			f["lstat_mode"], f["lstat_is_symlink"] = modeStr(li), li.Mode()&os.ModeSymlink != 0
		}
		if serr == nil {
			f["stat_mode"], f["stat_is_dir"] = modeStr(si), si.IsDir()
			want := tfi
			if c.isDir {
				want = tdi
			}
			f["samefile_stat_link_vs_target"] = os.SameFile(si, want)
		}
		if lerr == nil && serr == nil {
			f["samefile_lstat_vs_stat_of_link"] = os.SameFile(li, si)
		}
		if c.isDir {
			f["read_through"] = readThrough(filepath.Join(link, "inner.txt"))
			ents, err := os.ReadDir(link)
			f["readdir_through_count"], f["readdir_error"] = len(ents), errStr(err)
		} else {
			f["read_through"] = readThrough(link)
		}
		rec(t, "symlink_os", f)
	}

	// Dangling and symlink-to-symlink chain.
	dang := filepath.Join(root, "dangling")
	err := os.Symlink("nope.txt", dang)
	f := map[string]any{"case": "dangling", "symlink_error": errStr(err)}
	if err == nil {
		_, serr := os.Stat(dang)
		_, lerr := os.Lstat(dang)
		_, eerr := filepath.EvalSymlinks(dang)
		f["stat_error"], f["lstat_error"], f["evalsymlinks_error"] = errStr(serr), errStr(lerr), errStr(eerr)
	}
	rec(t, "symlink_os", f)

	chain := filepath.Join(root, "chain")
	err = os.Symlink("flink_rel", chain)
	f = map[string]any{"case": "chain", "symlink_error": errStr(err)}
	if err == nil {
		ev, eerr := filepath.EvalSymlinks(chain)
		f["evalsymlinks_equals_target"], f["evalsymlinks_error"] = eerr == nil && ev != "" && sameString(ev, targetFile, "", false), errStr(eerr)
		f["evalsymlinks"] = ev
		f["read_through"] = readThrough(chain)
	}
	rec(t, "symlink_os", f)
}

// sameString compares an EvalSymlinks result to the expected target, tolerating
// the temp dir itself being a symlink or short-name spelling (macOS /var, Windows 8.3).
func sameString(got, file, dir string, isDir bool) bool {
	want := file
	if isDir {
		want = dir
	}
	if got == want {
		return true
	}
	a, err1 := os.Stat(got)
	b, err2 := os.Stat(want)
	return err1 == nil && err2 == nil && os.SameFile(a, b) && strings.EqualFold(filepath.Base(got), filepath.Base(want))
}
