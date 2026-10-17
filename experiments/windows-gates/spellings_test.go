package windowsgates

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSpellingsSameManifest covers SC7: different spellings of the same
// manifest must be recognisable as one file. For every pair we record
// os.SameFile, canonical-path equality and whether a non-blocking exclusive
// lock taken through spelling A blocks spelling B.
func TestSpellingsSameManifest(t *testing.T) {
	root := t.TempDir()
	// Resolve the temp root once so symlinked temp dirs (macOS /var) do not
	// masquerade as a symlink spelling.
	if r, err := filepath.EvalSymlinks(root); err == nil {
		root = r
	}
	proj := filepath.Join(root, "proj")
	other := filepath.Join(root, "other")
	for _, d := range []string{proj, other} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	manifest := filepath.Join(proj, "gira.yaml")
	if err := os.WriteFile(manifest, []byte("version: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	type spelling struct{ name, path string }
	sp := []spelling{{"absolute", manifest}}

	// Relative spelling: from another cwd, ../proj/./gira.yaml.
	oldwd, _ := os.Getwd()
	if err := os.Chdir(other); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldwd)
	sp = append(sp, spelling{"relative_dotdot", filepath.Join("..", "proj", ".", "gira.yaml")})

	if err := os.Symlink("gira.yaml", filepath.Join(proj, "link_rel.yaml")); err == nil {
		sp = append(sp, spelling{"file_symlink_relative", filepath.Join(proj, "link_rel.yaml")})
	} else {
		rec(t, "spellings_setup", map[string]any{"spelling": "file_symlink_relative", "error": err.Error()})
	}
	if err := os.Symlink(manifest, filepath.Join(other, "link_abs.yaml")); err == nil {
		sp = append(sp, spelling{"file_symlink_absolute", filepath.Join(other, "link_abs.yaml")})
	} else {
		rec(t, "spellings_setup", map[string]any{"spelling": "file_symlink_absolute", "error": err.Error()})
	}
	if err := os.Symlink(proj, filepath.Join(other, "dir_link")); err == nil {
		sp = append(sp, spelling{"dir_symlink_absolute", filepath.Join(other, "dir_link", "gira.yaml")})
	} else {
		rec(t, "spellings_setup", map[string]any{"spelling": "dir_symlink_absolute", "error": err.Error()})
	}
	if err := os.Symlink(filepath.Join("..", "proj"), filepath.Join(other, "dir_link_rel")); err == nil {
		sp = append(sp, spelling{"dir_symlink_relative", filepath.Join(other, "dir_link_rel", "gira.yaml")})
	} else {
		rec(t, "spellings_setup", map[string]any{"spelling": "dir_symlink_relative", "error": err.Error()})
	}
	// Case variant: same file on case-insensitive filesystems (Windows, default macOS).
	sp = append(sp, spelling{"case_variant", filepath.Join(proj, "GIRA.YAML")})

	canonical := func(p string) string {
		abs, err := filepath.Abs(p)
		if err != nil {
			return "ERR:" + err.Error()
		}
		ev, err := filepath.EvalSymlinks(abs)
		if err != nil {
			return "ERR:" + err.Error()
		}
		return ev
	}

	for _, a := range sp {
		rec(t, "spelling_canonical", map[string]any{
			"spelling": a.name, "path": a.path,
			"abs": mustAbs(a.path), "evalsymlinks_abs": canonical(a.path),
			"canonical_equals_absolute": canonical(a.path) == canonical(manifest),
		})
	}

	base := sp[0]
	for _, b := range sp[1:] {
		f := map[string]any{"a": base.name, "b": b.name, "path_a": base.path, "path_b": b.path}
		ai, aerr := os.Stat(base.path)
		bi, berr := os.Stat(b.path)
		f["stat_error_a"], f["stat_error_b"] = errStr(aerr), errStr(berr)
		if aerr == nil && berr == nil {
			f["samefile"] = os.SameFile(ai, bi)
		}
		f["canonical_equal"] = canonical(base.path) == canonical(b.path)
		lr := lockContention(base.path, b.path)
		for k, v := range lr {
			f[k] = v
		}
		rec(t, "spelling_pair", f)
	}
}

func mustAbs(p string) string {
	a, err := filepath.Abs(p)
	if err != nil {
		return "ERR:" + err.Error()
	}
	return a
}

// lockContention holds an exclusive non-blocking lock via pathA and records
// whether pathB can also take one (acquired=true means NO exclusion).
func lockContention(pathA, pathB string) map[string]any {
	out := map[string]any{}
	fa, err := os.OpenFile(pathA, os.O_RDWR, 0)
	if err != nil {
		out["lock_open_a_error"] = err.Error()
		return out
	}
	defer fa.Close()
	ok, err := tryLockExclusive(fa)
	out["lock_a_acquired"], out["lock_a_error"] = ok, errStr(err)
	if !ok {
		return out
	}
	defer unlock(fa)
	fb, err := os.OpenFile(pathB, os.O_RDWR, 0)
	if err != nil {
		out["lock_open_b_error"] = err.Error()
		return out
	}
	defer fb.Close()
	okb, err := tryLockExclusive(fb)
	out["lock_b_acquired_while_a_held"], out["lock_b_error"] = okb, errStr(err)
	out["mutually_exclusive"] = !okb && err == nil
	if okb {
		unlock(fb)
	}
	return out
}
