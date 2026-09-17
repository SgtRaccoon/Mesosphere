package tasks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mesosphere/mesosphere/pkg/git"
)

// Git is the git client used for task commits. Tests may replace it.
var Git = &git.Client{}

// SaveTask validates rawJSON, writes it to the task's file, and commits.
// Empty commitMsg becomes "update(mesosphere): task <taskID>".
func SaveTask(repoPath, taskID, rawJSON, commitMsg string) error {
	if !json.Valid([]byte(rawJSON)) {
		return fmt.Errorf("invalid JSON payload for task %s", taskID)
	}

	task, err := GetTask(repoPath, "", taskID)
	if err != nil {
		return err
	}

	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("resolve repo: %w", err)
	}
	target := filepath.Join(absRepo, filepath.FromSlash(task.FilePath))
	rel, err := filepath.Rel(absRepo, target)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("task path escapes repository: %s", task.FilePath)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create task directory: %w", err)
	}
	if err := os.WriteFile(target, []byte(rawJSON), 0o644); err != nil {
		return fmt.Errorf("write task file: %w", err)
	}

	if strings.TrimSpace(commitMsg) == "" {
		commitMsg = fmt.Sprintf("update(mesosphere): task %s", taskID)
	}
	return Git.Commit(absRepo, filepath.ToSlash(rel), commitMsg)
}
