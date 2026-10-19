package manifest

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	// ErrManifestOutsideGitRoot is returned when a manifest with declarative overrides
	// is located outside of any Git repository worktree.
	ErrManifestOutsideGitRoot = errors.New("manifest_outside_git_root: manifest with declarative overrides must reside within a git repository root")

	// ErrGitRootUnavailable is returned when Git root discovery fails due to subprocess errors,
	// missing git binary, dubious ownership, or invalid git state.
	ErrGitRootUnavailable = errors.New("git_root_unavailable: failed to discover git repository root")

	// ErrManifestPathRequired is returned when declarative overrides are present but no manifest
	// path was provided to resolve the locator.
	ErrManifestPathRequired = errors.New("manifest_path_required: manifest path is required when declarative overrides are present")
)

// GitRootDiscoveryHook allows tests to override git root discovery. In production, it is nil.
var GitRootDiscoveryHook func(manifestPath string) (string, error)

// HasDeclarativeOverrides returns true if the manifest defines any relationship or metadata overrides.
func (m *Manifest) HasDeclarativeOverrides() bool {
	if m == nil {
		return false
	}
	ov := m.Overrides
	if ov == nil && m.TargetSystem.Overrides != nil {
		ov = m.TargetSystem.Overrides
	}
	if ov == nil {
		return false
	}
	return len(ov.Relationships) > 0 || len(ov.Metadata) > 0
}

// GitCommandEnv returns an execution environment with all repository-redirection and
// configuration variables stripped to prevent adversarial discovery hijacking.
func GitCommandEnv() []string {
	blocked := map[string]bool{
		"GIT_DIR":                          true,
		"GIT_WORK_TREE":                    true,
		"GIT_COMMON_DIR":                   true,
		"GIT_INDEX_FILE":                   true,
		"GIT_OBJECT_DIRECTORY":             true,
		"GIT_ALTERNATE_OBJECT_DIRECTORIES": true,
		"GIT_NAMESPACE":                    true,
		"GIT_PREFIX":                       true,
		"GIT_CEILING_DIRECTORIES":          true,
		"GIT_NO_REPLACE_OBJECTS":           true,
		"GIT_REPLACE_REF_BASE":             true,
		"GIT_LITERAL_PATHSPECS":            true,
		"GIT_GLOB_PATHSPECS":               true,
		"GIT_NOGLOB_PATHSPECS":             true,
		"GIT_ICASE_PATHSPECS":              true,
		"GIT_ATTR_SOURCE":                  true,
		"GIT_CONFIG":                       true,
		"GIT_CONFIG_GLOBAL":                true,
		"GIT_CONFIG_SYSTEM":                true,
		"GIT_CONFIG_NOSYSTEM":              true,
	}
	env := make([]string, 0, len(os.Environ())+3)
	for _, item := range os.Environ() {
		key, _, ok := strings.Cut(item, "=")
		if !ok || blocked[key] || strings.HasPrefix(key, "GIT_CONFIG_") {
			continue
		}
		env = append(env, item)
	}
	env = append(env,
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_NO_REPLACE_OBJECTS=1",
	)
	return env
}

// DefaultGitRootDiscoverer discovers the git repository root for the given manifest path
// using git rev-parse --show-toplevel under a strictly sanitized environment.
func DefaultGitRootDiscoverer(manifestPath string) (string, error) {
	if strings.TrimSpace(manifestPath) == "" {
		return "", ErrManifestPathRequired
	}

	absPath, err := filepath.Abs(manifestPath)
	if err != nil {
		return "", fmt.Errorf("%w: failed to determine absolute path for %q: %v", ErrGitRootUnavailable, manifestPath, err)
	}

	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("failed to read manifest file: %w", err)
		}
		return "", fmt.Errorf("%w: failed to resolve symlink for %q: %v", ErrGitRootUnavailable, manifestPath, err)
	}

	manifestDir := filepath.Dir(realPath)

	cmd := exec.Command("git", "-C", manifestDir, "rev-parse", "--show-toplevel")
	cmd.Env = GitCommandEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		outStr := string(out)
		if strings.Contains(outStr, "not a git repository") ||
			strings.Contains(outStr, "this operation must be run in a work tree") {
			return "", ErrManifestOutsideGitRoot
		}
		if strings.Contains(outStr, "dubious ownership") {
			return "", fmt.Errorf("%w: %s", ErrGitRootUnavailable, strings.TrimSpace(outStr))
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("%w: git rev-parse failed: %s (%v)", ErrGitRootUnavailable, strings.TrimSpace(outStr), err)
		}
		return "", fmt.Errorf("%w: failed to execute git: %v", ErrGitRootUnavailable, err)
	}

	gitRoot := strings.TrimSpace(string(out))
	if gitRoot == "" {
		return "", ErrManifestOutsideGitRoot
	}

	realGitRoot, evalErr := filepath.EvalSymlinks(gitRoot)
	if evalErr == nil {
		gitRoot = realGitRoot
	}
	absGitRoot, absErr := filepath.Abs(gitRoot)
	if absErr == nil {
		gitRoot = absGitRoot
	}

	return gitRoot, nil
}

// ResolveManifestLocator derives a normalized, forward-slash relative locator for
// manifestPath relative to the enclosing Git repository root.
//
// If Git discovery fails or the manifest is outside the repository root, it returns
// the corresponding stable fail-closed error.
func ResolveManifestLocator(manifestPath string) (string, error) {
	if strings.TrimSpace(manifestPath) == "" {
		return "", ErrManifestPathRequired
	}

	absPath, err := filepath.Abs(manifestPath)
	if err != nil {
		return "", fmt.Errorf("%w: failed to determine absolute path for %q: %v", ErrGitRootUnavailable, manifestPath, err)
	}

	realManifestPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("failed to read manifest file: %w", err)
		}
		return "", fmt.Errorf("%w: failed to resolve symlink for %q: %v", ErrGitRootUnavailable, manifestPath, err)
	}

	var gitRoot string
	if GitRootDiscoveryHook != nil {
		gitRoot, err = GitRootDiscoveryHook(manifestPath)
	} else {
		gitRoot, err = DefaultGitRootDiscoverer(manifestPath)
	}
	if err != nil {
		return "", err
	}

	realGitRoot, evalErr := filepath.EvalSymlinks(gitRoot)
	if evalErr == nil {
		gitRoot = realGitRoot
	}
	absGitRoot, absErr := filepath.Abs(gitRoot)
	if absErr == nil {
		gitRoot = absGitRoot
	}

	realManifestAbs, absErr := filepath.Abs(realManifestPath)
	if absErr == nil {
		realManifestPath = realManifestAbs
	}

	rel, err := filepath.Rel(gitRoot, realManifestPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || strings.HasPrefix(rel, "../") {
		return "", ErrManifestOutsideGitRoot
	}

	slashRel := filepath.ToSlash(filepath.Clean(rel))
	if slashRel == "." || slashRel == "" || strings.HasPrefix(slashRel, "../") {
		return "", ErrManifestOutsideGitRoot
	}

	return slashRel, nil
}
