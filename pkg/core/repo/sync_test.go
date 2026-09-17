package repo

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mesosphere/mesosphere/pkg/git"
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

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	gitEnv(cmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func initRepo(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "init", "-b", "main")
	runGit(t, dir, "config", "user.name", "Mesosphere Test")
	runGit(t, dir, "config", "user.email", "test@mesosphere.local")
}

func TestPublishRepositorySuccess(t *testing.T) {
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
	c := Git
	if err := c.Commit(local, "a.md", "initial"); err != nil {
		t.Fatal(err)
	}
	runGit(t, local, "remote", "add", "origin", bare)
	runGit(t, local, "push", "-u", "origin", "main")

	if err := os.WriteFile(filepath.Join(local, "a.md"), []byte("b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := c.Commit(local, "a.md", "update"); err != nil {
		t.Fatal(err)
	}

	st, err := GetRepoStatus(local)
	if err != nil {
		t.Fatal(err)
	}
	if st.Ahead < 1 {
		t.Errorf("ahead = %d, want >= 1", st.Ahead)
	}

	if err := PublishRepository(local); err != nil {
		t.Fatalf("PublishRepository: %v", err)
	}

	st, err = GetRepoStatus(local)
	if err != nil {
		t.Fatal(err)
	}
	if st.Ahead != 0 || st.Behind != 0 {
		t.Errorf("after publish ahead=%d behind=%d", st.Ahead, st.Behind)
	}
}

func TestPublishRepositoryConflict(t *testing.T) {
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
	if err := Git.Commit(a, "f.md", "base"); err != nil {
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
	if err := Git.Commit(a, "f.md", "from a"); err != nil {
		t.Fatal(err)
	}
	if err := Git.Push(a); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(b, "f.md"), []byte("from-b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Git.Commit(b, "f.md", "from b"); err != nil {
		t.Fatal(err)
	}

	err := PublishRepository(b)
	if err == nil {
		t.Fatal("expected conflict")
	}
	var ce *git.ConflictError
	if !errors.As(err, &ce) && !strings.Contains(strings.ToLower(err.Error()), "conflict") {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestSyncRepository(t *testing.T) {
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
	if err := Git.Commit(local, "a.md", "initial"); err != nil {
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

	if err := os.WriteFile(filepath.Join(local, "a.md"), []byte("synced\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Git.Commit(local, "a.md", "remote change"); err != nil {
		t.Fatal(err)
	}
	if err := Git.Push(local); err != nil {
		t.Fatal(err)
	}

	if err := SyncRepository(clone); err != nil {
		t.Fatalf("SyncRepository: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(clone, "a.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != "synced" {
		t.Errorf("content = %q", data)
	}
}
