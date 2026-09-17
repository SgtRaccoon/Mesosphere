package docs

import (
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

func initRepo(t *testing.T, dir string) *git.Client {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "init", "-b", "main")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	run := func(args ...string) {
		t.Helper()
		c := exec.Command("git", append([]string{"-C", dir}, args...)...)
		c.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=t@t.local",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=t@t.local",
		)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("config", "user.name", "Test")
	run("config", "user.email", "t@t.local")
	return &git.Client{}
}

func TestListDocumentsFindsMarkdownAndTxt(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docs", "arch.md"), []byte("# Architecture\n\nHello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("plain notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skip.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git", "ignored.md"), []byte("# no\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	eng := &Engine{}
	list, err := eng.ListDocuments(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("found %d docs: %+v", len(list), list)
	}
	byPath := map[string]DocumentSummary{}
	for _, d := range list {
		byPath[d.Path] = d
	}
	if byPath["docs/arch.md"].Title != "Architecture" {
		t.Errorf("title = %q", byPath["docs/arch.md"].Title)
	}
	if _, ok := byPath["notes.txt"]; !ok {
		t.Error("missing notes.txt")
	}
}

func TestGetDocumentVersions(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	g := initRepo(t, dir)
	eng := &Engine{Git: g}

	path := filepath.Join(dir, "readme.md")
	if err := os.WriteFile(path, []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := g.Commit(dir, "readme.md", "first"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := g.Commit(dir, "readme.md", "second"); err != nil {
		t.Fatal(err)
	}

	vers, err := eng.ListDocumentVersions(dir, "readme.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(vers) != 2 {
		t.Fatalf("versions = %d, want 2", len(vers))
	}

	old, err := eng.GetDocument(dir, "readme.md", vers[1].Hash)
	if err != nil {
		t.Fatal(err)
	}
	if old.Content != "v1\n" {
		t.Errorf("historical = %q", old.Content)
	}

	cur, err := eng.GetDocument(dir, "readme.md", "")
	if err != nil {
		t.Fatal(err)
	}
	if cur.Content != "v2\n" {
		t.Errorf("current = %q", cur.Content)
	}
	if cur.Version != "HEAD" {
		t.Errorf("version = %q", cur.Version)
	}
}

func TestSaveDocumentCommits(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	g := initRepo(t, dir)
	eng := &Engine{Git: g}

	if err := os.WriteFile(filepath.Join(dir, "doc.md"), []byte("orig\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := g.Commit(dir, "doc.md", "seed"); err != nil {
		t.Fatal(err)
	}

	if err := eng.SaveDocument(dir, "doc.md", "edited\n", ""); err != nil {
		t.Fatalf("SaveDocument: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "doc.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "edited\n" {
		t.Errorf("disk = %q", data)
	}

	vers, err := eng.ListDocumentVersions(dir, "doc.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(vers) != 2 {
		t.Fatalf("commits = %d, want 2", len(vers))
	}
	if vers[0].Message != "update(mesosphere): doc.md" {
		t.Errorf("message = %q", vers[0].Message)
	}
}
