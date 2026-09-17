package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mesosphere/mesosphere/pkg/core/config"
)

func TestRepoList(t *testing.T) {
	orig := loadConfig
	t.Cleanup(func() { loadConfig = orig })
	loadConfig = func() (*config.Config, error) {
		return &config.Config{Repositories: []config.RepositoryConfig{
			{ID: "main", Name: "Main", Path: "/tmp/main"},
		}}, nil
	}

	c := NewRootCommand()
	buf := new(bytes.Buffer)
	c.SetOut(buf)
	c.SetErr(buf)
	c.SetArgs([]string{"repo", "list"})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "main") || !strings.Contains(buf.String(), "/tmp/main") {
		t.Fatalf("output: %s", buf.String())
	}

	c = NewRootCommand()
	buf = new(bytes.Buffer)
	c.SetOut(buf)
	c.SetArgs([]string{"repo", "list", "--json"})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}
	var list []config.RepositoryConfig
	if err := json.Unmarshal(buf.Bytes(), &list); err != nil {
		t.Fatalf("json: %v\n%s", err, buf.String())
	}
	if len(list) != 1 || list[0].ID != "main" {
		t.Fatalf("%+v", list)
	}
}
