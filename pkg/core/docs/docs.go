// Package docs discovers markdown/text documents and retrieves git versions.
package docs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mesosphere/mesosphere/pkg/git"
)

// DocumentSummary is a discovered document in a repository.
type DocumentSummary struct {
	Path  string `json:"path"`
	Title string `json:"title"`
}

// DocumentDetail is document text at a given version.
type DocumentDetail struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Content string `json:"content"`
}

// Engine lists and reads documents using the git wrapper.
type Engine struct {
	Git *git.Client
}

func (e *Engine) git() *git.Client {
	if e != nil && e.Git != nil {
		return e.Git
	}
	return &git.Client{}
}

// ListDocuments walks repoPath for .md and .txt files (relative paths).
func (e *Engine) ListDocuments(repoPath string) ([]DocumentSummary, error) {
	var docs []DocumentSummary
	err := filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if skipDir(name) {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".md" && ext != ".txt" {
			return nil
		}
		rel, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		title := titleFromFile(path, rel)
		docs = append(docs, DocumentSummary{Path: rel, Title: title})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	if docs == nil {
		docs = []DocumentSummary{}
	}
	return docs, nil
}

func skipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", "dist", ".mesosphere":
		return true
	default:
		return false
	}
}

func titleFromFile(absPath, rel string) string {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return rel
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			return strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
		return line
	}
	return rel
}

// ListDocumentVersions returns git history for docPath (relative to repo).
func (e *Engine) ListDocumentVersions(repoPath, docPath string) ([]git.CommitInfo, error) {
	return e.git().GetDocHistory(repoPath, docPath)
}

// GetDocument returns working-tree content when version is empty, otherwise git show.
func (e *Engine) GetDocument(repoPath, docPath, version string) (*DocumentDetail, error) {
	version = strings.TrimSpace(version)
	var content string
	if version == "" {
		data, err := os.ReadFile(filepath.Join(repoPath, filepath.FromSlash(docPath)))
		if err != nil {
			return nil, fmt.Errorf("read document: %w", err)
		}
		content = string(data)
		version = "HEAD"
	} else {
		out, err := e.git().GetDocVersionContent(repoPath, docPath, version)
		if err != nil {
			return nil, err
		}
		content = out
	}
	return &DocumentDetail{Path: filepath.ToSlash(docPath), Version: version, Content: content}, nil
}
