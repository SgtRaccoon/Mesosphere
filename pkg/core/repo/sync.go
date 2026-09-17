package repo

import (
	"github.com/mesosphere/mesosphere/pkg/git"
)

// Git is the git client used for sync/publish. Tests may replace it.
var Git = &git.Client{}

// RepoStatus is an alias of git.RepoStatus for core callers.
type RepoStatus = git.RepoStatus

// SyncRepository pulls remote changes into the local working copy.
func SyncRepository(repoPath string) error {
	return Git.Pull(repoPath)
}

// PublishRepository pulls then pushes. Conflicts surface as *git.ConflictError.
func PublishRepository(repoPath string) error {
	if err := Git.Pull(repoPath); err != nil {
		return err
	}
	return Git.Push(repoPath)
}

// GetRepoStatus reports uncommitted files and ahead/behind vs upstream.
func GetRepoStatus(repoPath string) (*RepoStatus, error) {
	return Git.CheckRepoStatus(repoPath)
}
