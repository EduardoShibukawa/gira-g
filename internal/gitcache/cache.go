package gitcache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

// FileEntry represents an extracted file entry from git tree.
type FileEntry struct {
	Path string
	Mode string
	SHA  string
}

// CacheManager manages bare repository caches on the local filesystem.
type CacheManager struct {
	BaseDir string
}

// gitReadEnvironment removes inherited variables that can redirect repository
// discovery/object lookup or change the interpretation of pinned reads.
// EnsureRepo's clone/fetch commands intentionally do not use this environment:
// they retain the user's credentials and Git configuration.
func gitReadEnvironment() []string {
	blocked := map[string]bool{
		"GIT_DIR":                       true,
		"GIT_WORK_TREE":                 true,
		"GIT_COMMON_DIR":                true,
		"GIT_INDEX_FILE":                true,
		"GIT_OBJECT_DIRECTORY":          true,
		"GIT_ALTERNATE_OBJECT_DIRECTORIES": true,
		"GIT_NAMESPACE":                 true,
		"GIT_PREFIX":                    true,
		"GIT_CEILING_DIRECTORIES":        true,
		"GIT_DISCOVERY_ACROSS_FILESYSTEM": true,
		"GIT_REPLACE_REF_BASE":           true,
		"GIT_GRAFT_FILE":                 true,
		"GIT_SHALLOW_FILE":               true,
		"GIT_CONFIG":                     true,
	}
	env := make([]string, 0, len(os.Environ()))
	for _, item := range os.Environ() {
		key, _, ok := strings.Cut(item, "=")
		if !ok || blocked[key] || strings.HasPrefix(key, "GIT_CONFIG_") {
			continue
		}
		env = append(env, item)
	}
	return env
}

func newGitReadCommand(ctx context.Context, args ...string) *exec.Cmd {
	cmd := newGitReadCommand(ctx, args...)
	cmd.Env = gitReadEnvironment()
	return cmd
}

// NewCacheManager initializes a CacheManager with the provided base directory,
// or defaults to $XDG_CACHE_HOME/gira/repos or ~/.cache/gira/repos.
func NewCacheManager(baseDir string) (*CacheManager, error) {
	if strings.TrimSpace(baseDir) == "" {
		xdg := os.Getenv("XDG_CACHE_HOME")
		if xdg != "" {
			baseDir = filepath.Join(xdg, "gira", "repos")
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to determine user home directory: %w", err)
			}
			baseDir = filepath.Join(home, ".cache", "gira", "repos")
		}
	}
	return &CacheManager{BaseDir: baseDir}, nil
}

// NormalizeURL produces a deterministic normalized URL representation.
func NormalizeURL(rawURL string) string {
	u := strings.TrimSpace(rawURL)
	u = strings.TrimSuffix(u, "/")
	// Convert file paths to clean slashes if applicable
	if strings.HasPrefix(u, "file://") {
		pathPart := strings.TrimPrefix(u, "file://")
		pathPart = filepath.ToSlash(filepath.Clean(pathPart))
		return "file://" + pathPart
	}
	return u
}

// RepoHash calculates the deterministic SHA256 hex digest of the normalized repository URL.
func RepoHash(rawURL string) string {
	normalized := NormalizeURL(rawURL)
	h := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(h[:])
}

// RepoDir returns the bare repository cache directory path for the given repository.
func (cm *CacheManager) RepoDir(repo manifest.RepositoryDeclaration) string {
	hash := RepoHash(repo.URL)
	return filepath.Join(cm.BaseDir, hash+".git")
}

// HasCommit checks whether the bare repository at repoDir contains the specified commit SHA.
func (cm *CacheManager) HasCommit(ctx context.Context, repoDir, commitSHA string) bool {
	cmd := newGitReadCommand(ctx, "--no-replace-objects", "--git-dir="+repoDir, "cat-file", "-e", commitSHA+"^{commit}")
	return cmd.Run() == nil
}

// EnsureRepo ensures that the repository is cached locally and that the pinned commit exists.
func (cm *CacheManager) EnsureRepo(ctx context.Context, repo manifest.RepositoryDeclaration, offline bool) error {
	repoDir := cm.RepoDir(repo)

	dirExists := false
	if fi, err := os.Stat(repoDir); err == nil && fi.IsDir() {
		dirExists = true
	}

	if !dirExists {
		if offline {
			return ErrRepoNotFound{RepoID: repo.ID, CacheDir: repoDir}
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(repoDir), 0755); err != nil {
			return fmt.Errorf("failed to create cache directory parent: %w", err)
		}

		// Clone bare repository
		cmd := exec.CommandContext(ctx, "git", "clone", "--bare", repo.URL, repoDir)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			// Clean up partially created directory if any
			_ = os.RemoveAll(repoDir)
			return UnreachableError{
				RepoID: repo.ID,
				URL:    repo.URL,
				Reason: strings.TrimSpace(stderr.String()),
				Err:    err,
			}
		}
	}

	// Verify commit exists
	if cm.HasCommit(ctx, repoDir, repo.SHA) {
		return nil
	}

	// Commit not present: if offline, return error immediately
	if offline {
		return ErrCommitMissing{RepoID: repo.ID, SHA: repo.SHA}
	}

	// Fetch updates from origin
	cmd := exec.CommandContext(ctx, "git", "--git-dir="+repoDir, "fetch", "origin")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return UnreachableError{
			RepoID: repo.ID,
			URL:    repo.URL,
			Reason: fmt.Sprintf("failed to fetch origin: %s", strings.TrimSpace(stderr.String())),
			Err:    err,
		}
	}

	// Check again after fetch
	if !cm.HasCommit(ctx, repoDir, repo.SHA) {
		return ErrCommitMissing{RepoID: repo.ID, SHA: repo.SHA}
	}

	return nil
}

// ListTree extracts the list of files at the pinned commit, scoped to repo.PathScope if specified.
func (cm *CacheManager) ListTree(ctx context.Context, repo manifest.RepositoryDeclaration) ([]FileEntry, error) {
	repoDir := cm.RepoDir(repo)

	args := []string{"--no-replace-objects", "--git-dir=" + repoDir, "ls-tree", "-r", "-z", "--full-tree", "--full-name", repo.SHA}
	if repo.PathScope != "" && repo.PathScope != "." {
		args = append(args, repo.PathScope)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git ls-tree failed for repo %s: %s: %w", repo.ID, strings.TrimSpace(stderr.String()), err)
	}

	var entries []FileEntry
	// With -z, Git emits each record terminated by NUL and leaves pathname
	// bytes unquoted. Newlines, tabs and backslashes in names are therefore
	// data, not record delimiters or quoting syntax.
	for _, record := range bytes.Split(stdout.Bytes(), []byte{0}) {
		if len(record) == 0 {
			continue
		}
		// Format: <mode> <type> <object>\t<path>\0
		parts := bytes.SplitN(record, []byte{'\t'}, 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed git ls-tree record for repo %s", repo.ID)
		}
		metaParts := strings.Fields(string(parts[0]))
		if len(metaParts) != 3 {
			return nil, fmt.Errorf("malformed git ls-tree metadata for repo %s", repo.ID)
		}
		if metaParts[1] != "blob" {
			continue
		}
		entries = append(entries, FileEntry{
			Path: string(parts[1]),
			Mode: metaParts[0],
			SHA:  metaParts[2],
		})
	}

	return entries, nil
}

// ReadFile reads the content of a file at the pinned commit SHA.
func (cm *CacheManager) ReadFile(ctx context.Context, repo manifest.RepositoryDeclaration, filePath string) ([]byte, error) {
	repoDir := cm.RepoDir(repo)

	// Clean path
	cleanPath := filepath.ToSlash(filepath.Clean(filePath))
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	target := fmt.Sprintf("%s:%s", repo.SHA, cleanPath)
	cmd := newGitReadCommand(ctx, "--no-replace-objects", "--git-dir="+repoDir, "show", target)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to read file %q at commit %s for repo %s: %s: %w", filePath, repo.SHA, repo.ID, strings.TrimSpace(stderr.String()), err)
	}

	return stdout.Bytes(), nil
}

// ExportTree extracts a pinned tree into targetDir through an isolated staging
// directory. targetDir must not already exist. On any archive/extraction failure,
// staged files are removed and no partial tree is published at targetDir.
func (cm *CacheManager) ExportTree(ctx context.Context, repo manifest.RepositoryDeclaration, targetDir string) error {
	if strings.TrimSpace(targetDir) == "" {
		return fmt.Errorf("target directory cannot be empty")
	}
	targetDir = filepath.Clean(targetDir)
	parentDir := filepath.Dir(targetDir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create target parent directory: %w", err)
	}
	if _, err := os.Lstat(targetDir); err == nil {
		return fmt.Errorf("target directory already exists: %s", targetDir)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect target directory %s: %w", targetDir, err)
	}

	stageDir, err := os.MkdirTemp(parentDir, "."+filepath.Base(targetDir)+".gira-export-*")
	if err != nil {
		return fmt.Errorf("failed to create extraction staging directory: %w", err)
	}
	published := false
	defer func() {
		if !published {
			_ = os.RemoveAll(stageDir)
		}
	}()

	repoDir := cm.RepoDir(repo)
	archiveCmd := newGitReadCommand(ctx, "--no-replace-objects", "--git-dir="+repoDir, "archive", repo.SHA)
	if repo.PathScope != "" && repo.PathScope != "." {
		archiveCmd.Args = append(archiveCmd.Args, repo.PathScope)
	}

	tarCmd := exec.CommandContext(ctx, "tar", "-x", "-C", stageDir)
	pipe, err := archiveCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	tarCmd.Stdin = pipe

	var archiveStderr, tarStderr bytes.Buffer
	archiveCmd.Stderr = &archiveStderr
	tarCmd.Stderr = &tarStderr

	if err := tarCmd.Start(); err != nil {
		return fmt.Errorf("failed to start tar command: %w", err)
	}

	archiveErr := archiveCmd.Run()
	// Always reap tar even when git archive fails; tar may have extracted a
	// valid prefix of the stream. Its output is still only in stageDir here.
	tarErr := tarCmd.Wait()
	if archiveErr != nil {
		return fmt.Errorf("git archive failed: %s: %w", strings.TrimSpace(archiveStderr.String()), archiveErr)
	}
	if tarErr != nil {
		return fmt.Errorf("tar extract failed: %s: %w", strings.TrimSpace(tarStderr.String()), tarErr)
	}

	if err := os.Rename(stageDir, targetDir); err != nil {
		return fmt.Errorf("failed to publish extracted tree into %s: %w", targetDir, err)
	}
	published = true
	return nil
}
