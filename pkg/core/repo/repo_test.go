package repo

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mesosphere/mesosphere/pkg/core/config"
)

func sampleConfig(t *testing.T) (*config.Config, string, string) {
	t.Helper()
	root := t.TempDir()
	main := filepath.Join(root, "main")
	other := filepath.Join(root, "other")
	if err := os.MkdirAll(filepath.Join(main, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		Repositories: []config.RepositoryConfig{
			{ID: "mesosphere-main", Name: "Main", Path: main},
			{ID: "secondary-repo", Name: "Other", Path: other},
		},
	}
	return cfg, main, other
}

func TestListRepositories(t *testing.T) {
	cfg, _, _ := sampleConfig(t)
	list := ListRepositories(cfg)
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	if ListRepositories(nil) == nil {
		t.Fatal("nil config should yield empty slice")
	}
}

func TestAddRepository(t *testing.T) {
	cfg := &config.Config{}
	got, err := AddRepository(cfg, config.RepositoryConfig{Name: "Demo", Path: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID == "" || got.Name != "Demo" || len(cfg.Repositories) != 1 {
		t.Fatalf("%+v cfg=%+v", got, cfg)
	}
	_, err = AddRepository(cfg, config.RepositoryConfig{ID: got.ID, Path: t.TempDir()})
	if !errors.Is(err, ErrRepoExists) {
		t.Fatalf("dup: %v", err)
	}
}

func TestGetRepositoryContextByCwd(t *testing.T) {
	cfg, main, _ := sampleConfig(t)
	nested := filepath.Join(main, "docs")

	got, err := GetRepositoryContext(cfg, nested, "")
	if err != nil {
		t.Fatalf("nested cwd: %v", err)
	}
	if got.ID != "mesosphere-main" {
		t.Errorf("id = %q", got.ID)
	}

	got, err = GetRepositoryContext(cfg, main, "")
	if err != nil {
		t.Fatalf("repo root cwd: %v", err)
	}
	if got.ID != "mesosphere-main" {
		t.Errorf("id = %q", got.ID)
	}
}

func TestGetRepositoryContextRepoFlag(t *testing.T) {
	cfg, _, other := sampleConfig(t)
	outside := t.TempDir()

	got, err := GetRepositoryContext(cfg, outside, "secondary-repo")
	if err != nil {
		t.Fatalf("flag id: %v", err)
	}
	if got.ID != "secondary-repo" {
		t.Errorf("id = %q", got.ID)
	}

	got, err = GetRepositoryContext(cfg, outside, other)
	if err != nil {
		t.Fatalf("flag path: %v", err)
	}
	if got.ID != "secondary-repo" {
		t.Errorf("id = %q", got.ID)
	}

	_, err = GetRepositoryContext(cfg, outside, "missing")
	if !errors.Is(err, ErrRepoNotFound) {
		t.Fatalf("missing flag: %v", err)
	}
}

func TestGetRepositoryContextUnconfiguredCwd(t *testing.T) {
	cfg, _, _ := sampleConfig(t)
	_, err := GetRepositoryContext(cfg, t.TempDir(), "")
	if !errors.Is(err, ErrNoContext) {
		t.Fatalf("want ErrNoContext, got %v", err)
	}
}

func TestGetRepositoryContextEmptyConfig(t *testing.T) {
	_, err := GetRepositoryContext(&config.Config{}, t.TempDir(), "")
	if !errors.Is(err, ErrNoContext) {
		t.Fatalf("want ErrNoContext, got %v", err)
	}
}
