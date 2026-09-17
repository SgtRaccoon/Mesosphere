package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mesosphere/mesosphere/pkg/core/config"
	"github.com/mesosphere/mesosphere/pkg/git"
)

func TestDocListRequiresContext(t *testing.T) {
	origLoad, origWD := loadConfig, getwd
	t.Cleanup(func() {
		loadConfig, getwd = origLoad, origWD
	})
	loadConfig = func() (*config.Config, error) {
		return &config.Config{Repositories: []config.RepositoryConfig{
			{ID: "main", Path: filepath.Join(t.TempDir(), "configured")},
		}}, nil
	}
	getwd = func() (string, error) { return t.TempDir(), nil }

	c := NewRootCommand()
	buf := new(bytes.Buffer)
	c.SetOut(buf)
	c.SetErr(buf)
	c.SetArgs([]string{"doc", "list"})
	err := c.Execute()
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestDocListGetVersions(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	cmd := exec.Command("git", "-C", dir, "init", "-b", "main")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	run := func(args ...string) {
		t.Helper()
		c := exec.Command("git", append([]string{"-C", dir}, args...)...)
		c.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=t@t.local",
			"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=t@t.local",
		)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("config", "user.name", "Test")
	run("config", "user.email", "t@t.local")

	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Hello\nv1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	g := &git.Client{}
	if err := g.Commit(dir, "readme.md", "first"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Hello\nv2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := g.Commit(dir, "readme.md", "second"); err != nil {
		t.Fatal(err)
	}

	origLoad, origWD := loadConfig, getwd
	t.Cleanup(func() {
		loadConfig, getwd = origLoad, origWD
	})
	loadConfig = func() (*config.Config, error) {
		return &config.Config{Repositories: []config.RepositoryConfig{
			{ID: "main", Name: "Main", Path: dir},
		}}, nil
	}
	getwd = func() (string, error) { return t.TempDir(), nil }

	execRoot := func(args ...string) (string, error) {
		c := NewRootCommand()
		buf := new(bytes.Buffer)
		c.SetOut(buf)
		c.SetErr(buf)
		c.SetArgs(args)
		err := c.Execute()
		return buf.String(), err
	}

	out, err := execRoot("doc", "list", "--repo", "main")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out, "readme.md") {
		t.Fatalf("list output: %s", out)
	}

	out, err = execRoot("doc", "get", "readme.md", "--repo", "main")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !strings.Contains(out, "v2") {
		t.Fatalf("get current: %s", out)
	}

	vers, err := execRoot("doc", "versions", "readme.md", "--repo", "main")
	if err != nil {
		t.Fatalf("versions: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(vers), "\n")
	if len(lines) != 2 {
		t.Fatalf("versions lines = %d (%s)", len(lines), vers)
	}
	oldHash := strings.Fields(lines[1])[0]
	out, err = execRoot("doc", "get", "readme.md", "--repo", "main", "--version", oldHash)
	if err != nil {
		t.Fatalf("get version: %v", err)
	}
	if !strings.Contains(out, "v1") {
		t.Fatalf("historical: %s", out)
	}
}
