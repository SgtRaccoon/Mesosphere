package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SaveDocument writes content to docPath and commits it.
// If commitMsg is empty, uses "update(mesosphere): <filename>".
func (e *Engine) SaveDocument(repoPath, docPath, content, commitMsg string) error {
	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("resolve repo: %w", err)
	}
	target := filepath.Join(absRepo, filepath.FromSlash(docPath))
	rel, err := filepath.Rel(absRepo, target)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("document path escapes repository: %s", docPath)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create document directory: %w", err)
	}
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write document: %w", err)
	}

	if strings.TrimSpace(commitMsg) == "" {
		commitMsg = fmt.Sprintf("update(mesosphere): %s", filepath.Base(target))
	}
	return e.git().Commit(absRepo, filepath.ToSlash(rel), commitMsg)
}
