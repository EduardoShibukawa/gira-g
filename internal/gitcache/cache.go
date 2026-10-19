package gitcache

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
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

// gitRedirectEnv contains variables that can redirect repository/object lookup
// or change the interpretation of path arguments. They must not influence
// reads from a pinned repository. Clone/fetch keep user Git configuration, but
// still discard these redirection variables.
var gitRedirectEnv = map[string]bool{
	"GIT_DIR": true,
	"GIT_WORK_TREE": true,
	"GIT_COMMON_DIR": true,
	"GIT_INDEX_FILE": true,
	"GIT_OBJECT_DIRECTORY": true,
	"GIT_ALTERNATE_OBJECT_DIRECTORIES": true,
	"GIT_NAMESPACE": true,
	"GIT_PREFIX": true,
	"GIT_CEILING_DIRECTORIES": true,
	"GIT_NO_REPLACE_OBJECTS": true,
	"GIT_REPLACE_REF_BASE": true,
	"GIT_LITERAL_PATHSPECS": true,
	"GIT_GLOB_PATHSPECS": true,
	"GIT_NOGLOB_PATHSPECS": true,
	"GIT_ICASE_PATHSPECS": true,
	"GIT_ATTR_SOURCE": true,
}

// GitEnvForRead returns an environment for deterministic reads from pinned Git
// objects. It ignores user/system config and replacement refs, while preserving
// non-Git environment needed to run the process.
func GitEnvForRead() []string {
	return gitEnvironment(false)
}

// GitEnvForNetwork returns an environment for clone/fetch. It strips repository
// redirection variables but preserves user Git config and credential settings.
func GitEnvForNetwork() []string {
	return gitEnvironment(true)
}

func gitEnvironment(allowUserConfig bool) []string {
	env := make([]string, 0, len(os.Environ())+3)
	for _, item := range os.Environ() {
		key, _, ok := strings.Cut(item, "=")
		if !ok || gitRedirectEnv[key] {
			continue
		}
		if !allowUserConfig && (key == "GIT_CONFIG" || strings.HasPrefix(key, "GIT_CONFIG_")) {
			continue
		}
		env = append(env, item)
	}
	if !allowUserConfig {
		// os.DevNull is "NUL" on Windows and "/dev/null" on Unix.
		env = append(env,
			"GIT_CONFIG_GLOBAL="+os.DevNull,
			"GIT_CONFIG_NOSYSTEM=1",
			"GIT_NO_REPLACE_OBJECTS=1",
		)
	}
	return env
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
	cmd := exec.CommandContext(ctx, "git", "--git-dir="+repoDir, "cat-file", "-e", commitSHA+"^{commit}")
	cmd.Env = GitEnvForRead()
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
		cmd.Env = GitEnvForNetwork()
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
	cmd.Env = GitEnvForNetwork()
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

	args := []string{"--git-dir=" + repoDir, "ls-tree", "-r", "-z", "--full-tree", "--full-name", repo.SHA}
	if repo.PathScope != "" && repo.PathScope != "." {
		args = append(args, repo.PathScope)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = GitEnvForRead()
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
	cmd := exec.CommandContext(ctx, "git", "--git-dir="+repoDir, "show", target)
	cmd.Env = GitEnvForRead()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to read file %q at commit %s for repo %s: %s: %w", filePath, repo.SHA, repo.ID, strings.TrimSpace(stderr.String()), err)
	}

	return stdout.Bytes(), nil
}

// archiveMember records one entry that git archive encoded in the tar stream.
type archiveMember struct {
	name string
	kind byte
	size int64
}

// readArchiveMembers validates archive paths before handing the tar to an
// external extractor. It also records the exact member names to check after
// extraction, since some platform tar builds can silently alter or omit names.
func readArchiveMembers(archivePath string) ([]archiveMember, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("open archive for validation: %w", err)
	}
	defer file.Close()

	reader := tar.NewReader(file)
	var members []archiveMember
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read archive header: %w", err)
		}
		name := strings.TrimSuffix(header.Name, "/")
		if name == "" || name == "." {
			continue
		}
		clean := path.Clean(name)
		if clean != name || clean == ".." || strings.HasPrefix(clean, "../") || path.IsAbs(name) {
			return nil, fmt.Errorf("unsafe archive entry path %q", header.Name)
		}
		platformPath := filepath.FromSlash(name)
		if filepath.IsAbs(platformPath) || filepath.VolumeName(platformPath) != "" {
			return nil, fmt.Errorf("unsafe archive entry path %q", header.Name)
		}
		if os.PathSeparator == '\\' && strings.ContainsRune(name, '\\') {
			return nil, fmt.Errorf("archive entry %q cannot be materialized on Windows because backslashes are path separators", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir, tar.TypeReg, tar.TypeRegA, tar.TypeSymlink, tar.TypeLink:
			// These are the entry types expected from a Git tree archive.
		default:
			return nil, fmt.Errorf("unsupported archive entry type %d for %q", header.Typeflag, header.Name)
		}
		if header.Typeflag == tar.TypeLink {
			linkName := path.Clean(header.Linkname)
			linkPath := filepath.FromSlash(header.Linkname)
			if linkName != header.Linkname || linkName == ".." || strings.HasPrefix(linkName, "../") || path.IsAbs(header.Linkname) || filepath.IsAbs(linkPath) || filepath.VolumeName(linkPath) != "" {
				return nil, fmt.Errorf("unsafe archive hard-link target %q for %q", header.Linkname, header.Name)
			}
			if os.PathSeparator == '\\' && strings.ContainsRune(header.Linkname, '\\') {
				return nil, fmt.Errorf("archive hard-link target %q cannot be materialized on Windows", header.Linkname)
			}
		}
		members = append(members, archiveMember{name: name, kind: header.Typeflag, size: header.Size})
	}
	return members, nil
}

// verifyArchiveMaterialized confirms that every non-directory archive member
// exists under the staging root with the expected type and file size. A tar
// implementation that exits 0 but cannot represent a pathname must fail closed.
func verifyArchiveMaterialized(stageDir string, members []archiveMember) error {
	for _, member := range members {
		fullPath := filepath.Join(stageDir, filepath.FromSlash(member.name))
		info, err := os.Lstat(fullPath)
		if err != nil {
			return fmt.Errorf("archive extraction incomplete: entry %q was not materialized: %w", member.name, err)
		}
		switch member.kind {
		case tar.TypeDir:
			if !info.IsDir() {
				return fmt.Errorf("archive extraction incomplete: directory entry %q was materialized as a non-directory", member.name)
			}
		case tar.TypeReg, tar.TypeRegA, tar.TypeLink:
			if !info.Mode().IsRegular() {
				return fmt.Errorf("archive extraction incomplete: regular entry %q was materialized as mode %s", member.name, info.Mode())
			}
			if member.kind != tar.TypeLink && info.Size() != member.size {
				return fmt.Errorf("archive extraction incomplete: entry %q size is %d bytes, archive declared %d", member.name, info.Size(), member.size)
			}
		case tar.TypeSymlink:
			if info.Mode()&os.ModeSymlink == 0 {
				return fmt.Errorf("archive extraction incomplete: symlink entry %q was materialized as mode %s", member.name, info.Mode())
			}
		}
	}
	return nil
}

// ExportTree checks out the contents of the tree at pinned commit SHA into targetDir.
// Extraction is staged in a sibling temporary directory; targetDir becomes visible
// only after git archive, tar extraction, and an entry-by-entry materialization check
// succeed. This prevents partial or platform-mangled trees from being published.
func (cm *CacheManager) ExportTree(ctx context.Context, repo manifest.RepositoryDeclaration, targetDir string) error {
	if strings.TrimSpace(targetDir) == "" {
		return fmt.Errorf("export target path cannot be empty")
	}
	cleanTarget := filepath.Clean(targetDir)
	if cleanTarget == "." || filepath.Dir(cleanTarget) == cleanTarget {
		return fmt.Errorf("refusing to replace current directory or filesystem root")
	}
	parentDir := filepath.Dir(cleanTarget)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create target parent directory: %w", err)
	}

	targetExists := false
	targetMode := os.FileMode(0755)
	if info, err := os.Lstat(cleanTarget); err == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("export target %q is not a real directory", cleanTarget)
		}
		contents, err := os.ReadDir(cleanTarget)
		if err != nil {
			return fmt.Errorf("failed to inspect export target %q: %w", cleanTarget, err)
		}
		if len(contents) != 0 {
			return fmt.Errorf("export target directory %q is not empty", cleanTarget)
		}
		targetExists = true
		targetMode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect export target %q: %w", cleanTarget, err)
	}

	repoDir := cm.RepoDir(repo)
	stageDir, err := os.MkdirTemp(parentDir, ".gira-export-*")
	if err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}
	defer os.RemoveAll(stageDir)

	archiveFile, err := os.CreateTemp(parentDir, ".gira-archive-*.tar")
	if err != nil {
		return fmt.Errorf("failed to create archive staging file: %w", err)
	}
	archivePath := archiveFile.Name()
	defer os.Remove(archivePath)

	archiveCmd := exec.CommandContext(ctx, "git", "--git-dir="+repoDir, "archive", repo.SHA)
	archiveCmd.Env = GitEnvForRead()
	if repo.PathScope != "" && repo.PathScope != "." {
		archiveCmd.Args = append(archiveCmd.Args, repo.PathScope)
	}
	var archiveStderr bytes.Buffer
	archiveCmd.Stdout = archiveFile
	archiveCmd.Stderr = &archiveStderr

	archiveErr := archiveCmd.Run()
	closeErr := archiveFile.Close()
	if archiveErr != nil {
		return fmt.Errorf("git archive failed: %s: %w", strings.TrimSpace(archiveStderr.String()), archiveErr)
	}
	if closeErr != nil {
		return fmt.Errorf("failed to close staged archive: %w", closeErr)
	}

	members, err := readArchiveMembers(archivePath)
	if err != nil {
		return fmt.Errorf("failed to validate git archive: %w", err)
	}

	var tarStderr bytes.Buffer
	tarCmd := exec.CommandContext(ctx, "tar", "-x", "-C", stageDir, "-f", archivePath)
	tarCmd.Stderr = &tarStderr
	if err := tarCmd.Run(); err != nil {
		return fmt.Errorf("tar extract failed: %s: %w", strings.TrimSpace(tarStderr.String()), err)
	}
	if err := verifyArchiveMaterialized(stageDir, members); err != nil {
		return err
	}

	// MkdirTemp keeps the staging root private (0700). Restore the exported
	// root's historical permissions only once the complete tree is ready.
	if err := os.Chmod(stageDir, 0755); err != nil {
		return fmt.Errorf("failed to set exported tree permissions: %w", err)
	}
	if targetExists {
		if err := os.Remove(cleanTarget); err != nil {
			return fmt.Errorf("failed to replace empty export target %q: %w", cleanTarget, err)
		}
	}
	if err := os.Rename(stageDir, cleanTarget); err != nil {
		if targetExists {
			if restoreErr := os.Mkdir(cleanTarget, targetMode); restoreErr != nil {
				return fmt.Errorf("publish failed (%v) and empty export target could not be restored: %w", err, restoreErr)
			}
			if chmodErr := os.Chmod(cleanTarget, targetMode); chmodErr != nil {
				return fmt.Errorf("publish failed (%v) and restored target permissions could not be set: %w", err, chmodErr)
			}
		}
		return fmt.Errorf("failed to publish extracted tree at %q: %w", cleanTarget, err)
	}
	return nil
}
