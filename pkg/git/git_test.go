package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
}

func gitEnv(cmd *exec.Cmd) {
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Mesosphere Test",
		"GIT_AUTHOR_EMAIL=test@mesosphere.local",
		"GIT_COMMITTER_NAME=Mesosphere Test",
		"GIT_COMMITTER_EMAIL=test@mesosphere.local",
	)
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	gitEnv(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func initRepo(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "init", "-b", "main")
	runGit(t, dir, "config", "user.name", "Mesosphere Test")
	runGit(t, dir, "config", "user.email", "test@mesosphere.local")
}

func TestCommitHistoryAndVersion(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	initRepo(t, dir)

	doc := filepath.Join(dir, "readme.md")
	if err := os.WriteFile(doc, []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	c := &Client{}
	if err := c.Commit(dir, "readme.md", "add readme"); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	if err := os.WriteFile(doc, []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := c.Commit(dir, "readme.md", "update readme"); err != nil {
		t.Fatalf("Commit 2: %v", err)
	}

	hist, err := c.GetDocHistory(dir, "readme.md")
	if err != nil {
		t.Fatalf("GetDocHistory: %v", err)
	}
	if len(hist) != 2 {
		t.Fatalf("history len = %d, want 2", len(hist))
	}
	if hist[0].Message != "update readme" {
		t.Errorf("newest message = %q", hist[0].Message)
	}

	old, err := c.GetDocVersionContent(dir, "readme.md", hist[1].Hash)
	if err != nil {
		t.Fatalf("GetDocVersionContent: %v", err)
	}
	if old != "v1\n" {
		t.Errorf("old content = %q, want v1", old)
	}

	st, err := c.CheckRepoStatus(dir)
	if err != nil {
		t.Fatalf("CheckRepoStatus: %v", err)
	}
	if st.HasUncommitted {
		t.Error("expected clean working tree")
	}
	if st.Branch != "main" {
		t.Errorf("branch = %q, want main", st.Branch)
	}

	if err := os.WriteFile(doc, []byte("dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err = c.CheckRepoStatus(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !st.HasUncommitted {
		t.Error("expected uncommitted changes")
	}
}

func TestFetchPullPushAndAheadBehind(t *testing.T) {
	requireGit(t)
	root := t.TempDir()
	bare := filepath.Join(root, "remote.git")
	if err := os.MkdirAll(bare, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, bare, "init", "--bare", "-b", "main")

	local := filepath.Join(root, "local")
	if err := os.MkdirAll(local, 0o755); err != nil {
		t.Fatal(err)
	}
	initRepo(t, local)
	if err := os.WriteFile(filepath.Join(local, "a.md"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := &Client{}
	if err := c.Commit(local, "a.md", "initial"); err != nil {
		t.Fatal(err)
	}
	runGit(t, local, "remote", "add", "origin", bare)
	runGit(t, local, "push", "-u", "origin", "main")

	clone := filepath.Join(root, "clone")
	cmd := exec.Command("git", "clone", bare, clone)
	gitEnv(cmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("clone: %v\n%s", err, out)
	}
	runGit(t, clone, "config", "user.name", "Mesosphere Test")
	runGit(t, clone, "config", "user.email", "test@mesosphere.local")

	if err := os.WriteFile(filepath.Join(local, "a.md"), []byte("local-change\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := c.Commit(local, "a.md", "local update"); err != nil {
		t.Fatal(err)
	}
	if err := c.Push(local); err != nil {
		t.Fatalf("Push: %v", err)
	}

	if err := c.Fetch(clone); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	st, err := c.CheckRepoStatus(clone)
	if err != nil {
		t.Fatal(err)
	}
	if st.Behind < 1 {
		t.Errorf("behind = %d, want >= 1", st.Behind)
	}

	if err := c.Pull(clone); err != nil {
		t.Fatalf("Pull: %v", err)
	}
	st, err = c.CheckRepoStatus(clone)
	if err != nil {
		t.Fatal(err)
	}
	if st.Behind != 0 || st.Ahead != 0 {
		t.Errorf("after pull ahead=%d behind=%d", st.Ahead, st.Behind)
	}

	content, err := c.GetDocVersionContent(clone, "a.md", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if content != "local-change\n" {
		t.Errorf("cloned content = %q", content)
	}
}

func TestPullConflict(t *testing.T) {
	requireGit(t)
	root := t.TempDir()
	bare := filepath.Join(root, "remote.git")
	if err := os.MkdirAll(bare, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, bare, "init", "--bare", "-b", "main")

	a := filepath.Join(root, "a")
	b := filepath.Join(root, "b")
	if err := os.MkdirAll(a, 0o755); err != nil {
		t.Fatal(err)
	}
	initRepo(t, a)
	if err := os.WriteFile(filepath.Join(a, "f.md"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := &Client{}
	if err := c.Commit(a, "f.md", "base"); err != nil {
		t.Fatal(err)
	}
	runGit(t, a, "remote", "add", "origin", bare)
	runGit(t, a, "push", "-u", "origin", "main")

	cmd := exec.Command("git", "clone", bare, b)
	gitEnv(cmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("clone: %v\n%s", err, out)
	}
	runGit(t, b, "config", "user.name", "Mesosphere Test")
	runGit(t, b, "config", "user.email", "test@mesosphere.local")

	if err := os.WriteFile(filepath.Join(a, "f.md"), []byte("from-a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := c.Commit(a, "f.md", "from a"); err != nil {
		t.Fatal(err)
	}
	if err := c.Push(a); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(b, "f.md"), []byte("from-b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := c.Commit(b, "f.md", "from b"); err != nil {
		t.Fatal(err)
	}

	err := c.Pull(b)
	if err == nil {
		t.Fatal("expected pull conflict")
	}
	var ce *ConflictError
	if !errors.As(err, &ce) && !strings.Contains(strings.ToLower(err.Error()), "conflict") {
		t.Fatalf("expected conflict, got %v", err)
	}
}
