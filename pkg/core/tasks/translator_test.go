package tasks

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mesosphere/mesosphere/pkg/git"
)

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestTranslateJSONAndMarkdown(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "tasks.json", `{
  "tasks": [
    {
      "id": "TASK-101",
      "title": "Ship config",
      "scope": "pkg/core/config",
      "validation_gate": "tests pass",
      "status": "IN PROGRESS",
      "is_blocked": true,
      "blocker_ids": ["TASK-100"]
    }
  ]
}`)
	write(t, dir, "todo.md", "- [ ] Write docs\n- [x] Init repo\n")
	write(t, dir, "skip.go", "package x\n")

	list, err := ListTasks(dir, "repo-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("got %d tasks: %+v", len(list), list)
	}

	got, err := GetTask(dir, "repo-1", "TASK-101")
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Ship config" || got.Scope != "pkg/core/config" {
		t.Errorf("mapping: %+v", got)
	}
	if got.ValidationGate != "tests pass" {
		t.Errorf("gate = %q", got.ValidationGate)
	}
	if !got.IsBlocked || len(got.BlockerID) != 1 || got.BlockerID[0] != "TASK-100" {
		t.Errorf("blocked = %+v", got)
	}
	if got.Status != StatusInProgress {
		t.Errorf("status = %q", got.Status)
	}

	write(t, dir, "docs/spec-kit.md", "# spec-kit notes\n\nNot JSON.\n")
	list, err = ListTasks(dir, "repo-1")
	if err != nil {
		t.Fatalf("unparseable detected file should be skipped: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("after skip got %d tasks", len(list))
	}

	done, err := GetTask(dir, "repo-1", "todo.md#2")
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != StatusDone {
		t.Errorf("checkbox done status = %q", done.Status)
	}
}

func TestFrameworkTranslators(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		content string
		id      string
		fw      string
	}{
		{"speckit", ".specify/tasks.json", `{"speckit": true, "tasks": [{"id": "SK-1", "title": "A", "status": "todo"}]}`, "SK-1", "github-spec-kit"},
		{"openspec", "openspec/change.yaml", "openspec: true\ntasks:\n  - id: OS-1\n    title: B\n    status: done\n", "OS-1", "openspec"},
		{"bmad", "bmad/stories.yml", "bmad: true\nstories:\n  - id: BM-1\n    title: C\n    status: TO DO\n", "BM-1", "bmad-method"},
		{"kiro", ".kiro/tasks.json", `{"tasks":[{"id":"KI-1","title":"D","status":"open"}]}`, "KI-1", "amazon-kiro"},
		{"cosmos", "cosmos/cards.json", `{"cards":[{"id":"CX-1","title":"E","status":"wip"}]}`, "CX-1", "augment-cosmos"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr := detect(defaultTranslators, tc.path, tc.content)
			if tr == nil {
				t.Fatal("no translator")
			}
			if tr.Name() != tc.fw {
				t.Fatalf("framework = %s, want %s", tr.Name(), tc.fw)
			}
			tasks, err := tr.Translate(tc.path, tc.content, "r")
			if err != nil {
				t.Fatal(err)
			}
			if len(tasks) != 1 || tasks[0].CardID != tc.id {
				t.Fatalf("tasks = %+v", tasks)
			}
		})
	}
}

func TestDeriveColumns(t *testing.T) {
	bin := DeriveColumns(nil, true)
	if len(bin) != 2 || bin[0] != StatusTodo || bin[1] != StatusDone {
		t.Fatalf("binary = %v", bin)
	}
	cols := DeriveColumns([]CommonTask{
		{Status: StatusTodo, IsBlocked: true},
		{Status: StatusInProgress},
	}, false)
	if len(cols) != 2 {
		t.Fatalf("cols = %v", cols)
	}
	for _, c := range cols {
		if c == "BLOCKED" {
			t.Fatal("must not emit BLOCKED column")
		}
	}
}

func TestSaveTaskCommitsJSON(t *testing.T) {
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

	orig := `{"tasks":[{"id":"TASK-101","title":"Old","status":"TO DO"}]}`
	write(t, dir, "tasks.json", orig)
	g := &git.Client{}
	if err := g.Commit(dir, "tasks.json", "seed"); err != nil {
		t.Fatal(err)
	}

	updated := `{"tasks":[{"id":"TASK-101","title":"New","status":"DONE"}]}`
	if err := SaveTask(dir, "TASK-101", updated, ""); err != nil {
		t.Fatalf("SaveTask: %v", err)
	}
	if err := SaveTask(dir, "TASK-101", "not-json", ""); err == nil {
		t.Fatal("expected invalid JSON error")
	}

	data, err := os.ReadFile(filepath.Join(dir, "tasks.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != updated {
		t.Errorf("disk = %s", data)
	}

hist, err := g.GetDocHistory(dir, "tasks.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(hist) != 2 {
		t.Fatalf("commits = %d", len(hist))
	}
	if hist[0].Message != "update(mesosphere): task TASK-101" {
		t.Errorf("message = %q", hist[0].Message)
	}
}
