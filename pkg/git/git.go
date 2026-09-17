// Package git wraps the system git executable for repository operations.
package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// ConflictError indicates a pull/push failed due to merge conflicts.
type ConflictError struct {
	Op     string
	Output string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("git %s conflict: %s", e.Op, strings.TrimSpace(e.Output))
}

// CommitInfo is one history entry for a file.
type CommitInfo struct {
	Hash      string
	Author    string
	Timestamp time.Time
	Message   string
}

// RepoStatus describes working tree and tracking-branch divergence.
type RepoStatus struct {
	HasUncommitted bool
	Ahead          int
	Behind         int
	Branch         string
	Upstream       string
}

// Client executes git in a repository.
type Client struct {
	GitBin string
}

func (c *Client) bin() string {
	if c != nil && c.GitBin != "" {
		return c.GitBin
	}
	return "git"
}

func (c *Client) run(repoPath string, args ...string) (string, error) {
	cmdArgs := append([]string{"-C", repoPath}, args...)
	cmd := exec.Command(c.bin(), cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := stdout.String() + stderr.String()
	if err != nil {
		return out, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(out))
	}
	return out, nil
}

func isConflictOutput(out string) bool {
	lower := strings.ToLower(out)
	return strings.Contains(lower, "conflict") ||
		strings.Contains(lower, "fix conflicts") ||
		strings.Contains(lower, "unmerged")
}

// Commit stages filePath (relative to repo) and creates a commit.
func (c *Client) Commit(repoPath, filePath, message string) error {
	if message == "" {
		return errors.New("commit message is required")
	}
	if _, err := c.run(repoPath, "add", "--", filePath); err != nil {
		return err
	}
	_, err := c.run(repoPath, "commit", "-m", message, "--", filePath)
	return err
}

// Pull runs git pull and reports merge conflicts.
func (c *Client) Pull(repoPath string) error {
	out, err := c.run(repoPath, "pull")
	if err != nil && isConflictOutput(out) {
		return &ConflictError{Op: "pull", Output: out}
	}
	return err
}

// Push runs git push and reports rejected/conflict output.
func (c *Client) Push(repoPath string) error {
	out, err := c.run(repoPath, "push")
	if err != nil && isConflictOutput(out) {
		return &ConflictError{Op: "push", Output: out}
	}
	return err
}

// Fetch runs git fetch.
func (c *Client) Fetch(repoPath string) error {
	_, err := c.run(repoPath, "fetch")
	return err
}

// GetDocHistory returns commits that touched filePath, newest first.
func (c *Client) GetDocHistory(repoPath, filePath string) ([]CommitInfo, error) {
	out, err := c.run(repoPath, "log", "--pretty=format:%H%x09%an%x09%at%x09%s", "--", filePath)
	if err != nil {
		return nil, err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return []CommitInfo{}, nil
	}
	lines := strings.Split(out, "\n")
	history := make([]CommitInfo, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 4)
		if len(parts) < 4 {
			continue
		}
		unix, _ := strconv.ParseInt(parts[2], 10, 64)
		history = append(history, CommitInfo{
			Hash:      parts[0],
			Author:    parts[1],
			Timestamp: time.Unix(unix, 0).UTC(),
			Message:   parts[3],
		})
	}
	return history, nil
}

// GetDocVersionContent returns file contents at commitHash.
func (c *Client) GetDocVersionContent(repoPath, filePath, commitHash string) (string, error) {
	spec := commitHash + ":" + filepathToGit(filePath)
	out, err := c.run(repoPath, "show", spec)
	if err != nil {
		return "", err
	}
	return out, nil
}

func filepathToGit(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}

// CheckRepoStatus reports dirty files and ahead/behind vs upstream.
func (c *Client) CheckRepoStatus(repoPath string) (*RepoStatus, error) {
	st := &RepoStatus{}

	porcelain, err := c.run(repoPath, "status", "--porcelain")
	if err != nil {
		return nil, err
	}
	st.HasUncommitted = strings.TrimSpace(porcelain) != ""

	branchOut, err := c.run(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, err
	}
	st.Branch = strings.TrimSpace(branchOut)

	upOut, upErr := c.run(repoPath, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	if upErr != nil {
		return st, nil
	}
	st.Upstream = strings.TrimSpace(upOut)

	counts, err := c.run(repoPath, "rev-list", "--left-right", "--count", "@{upstream}...HEAD")
	if err != nil {
		return st, nil
	}
	fields := strings.Fields(strings.TrimSpace(counts))
	if len(fields) == 2 {
		st.Behind, _ = strconv.Atoi(fields[0])
		st.Ahead, _ = strconv.Atoi(fields[1])
	}
	return st, nil
}
