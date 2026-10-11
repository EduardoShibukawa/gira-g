package gitcache

import "fmt"

// ErrCommitMissing is returned when a pinned commit SHA is not found in the cache.
type ErrCommitMissing struct {
	RepoID string
	SHA    string
}

func (e ErrCommitMissing) Error() string {
	return fmt.Sprintf("commit %s missing from repository %q cache", e.SHA, e.RepoID)
}

// ErrRepoNotFound is returned when a repository cache does not exist locally (e.g. in offline mode).
type ErrRepoNotFound struct {
	RepoID   string
	CacheDir string
}

func (e ErrRepoNotFound) Error() string {
	return fmt.Sprintf("repository %q cache directory not found at %s", e.RepoID, e.CacheDir)
}

// UnreachableError represents a network failure or remote repository access error.
type UnreachableError struct {
	RepoID string
	URL    string
	Reason string
	Err    error
}

func (e UnreachableError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("repository %q at %s unreachable: %s: %v", e.RepoID, e.URL, e.Reason, e.Err)
	}
	return fmt.Sprintf("repository %q at %s unreachable: %s", e.RepoID, e.URL, e.Reason)
}

func (e UnreachableError) Unwrap() error {
	return e.Err
}
