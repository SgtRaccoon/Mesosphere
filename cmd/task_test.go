package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mesosphere/mesosphere/pkg/core/config"
	"github.com/mesosphere/mesosphere/pkg/core/tasks"
)

func TestTaskListAndGetJSON(t *testing.T) {
	dir := t.TempDir()
	payload := `{"tasks":[{"id":"TASK-101","title":"Ship","scope":"cli","validation_gate":"tests","status":"TO DO"}]}`
	if err := os.WriteFile(filepath.Join(dir, "tasks.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}

	origLoad, origWD := loadConfig, getwd
	t.Cleanup(func() {
		loadConfig, getwd = origLoad, origWD
	})
	loadConfig = func() (*config.Config, error) {
		return &config.Config{Repositories: []config.RepositoryConfig{
			{ID: "main", Path: dir},
		}}, nil
	}
	getwd = func() (string, error) { return dir, nil }

	execRoot := func(args ...string) (string, error) {
		c := NewRootCommand()
		buf := new(bytes.Buffer)
		c.SetOut(buf)
		c.SetErr(buf)
		c.SetArgs(args)
		err := c.Execute()
		return buf.String(), err
	}

	out, err := execRoot("task", "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var list []tasks.CommonTask
	if err := json.Unmarshal([]byte(out), &list); err != nil {
		t.Fatalf("list json: %v\n%s", err, out)
	}
	if len(list) != 1 || list[0].CardID != "TASK-101" {
		t.Fatalf("list = %+v", list)
	}

	out, err = execRoot("task", "get", "TASK-101", "--repo", "main")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	var one tasks.CommonTask
	if err := json.Unmarshal([]byte(out), &one); err != nil {
		t.Fatalf("get json: %v\n%s", err, out)
	}
	if one.Title != "Ship" {
		t.Errorf("title = %q", one.Title)
	}
}
