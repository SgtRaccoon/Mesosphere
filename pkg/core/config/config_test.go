package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigMissingCreatesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Version != DefaultVersion {
		t.Errorf("version = %q, want %q", cfg.Version, DefaultVersion)
	}
	if cfg.Settings.Server.Port != DefaultPort {
		t.Errorf("port = %d, want %d", cfg.Settings.Server.Port, DefaultPort)
	}
	if !cfg.Settings.Server.AutoOpenBrowser {
		t.Error("expected auto_open_browser true")
	}
	if cfg.Settings.UI.ActiveTab != "docs" {
		t.Errorf("active_tab = %q, want docs", cfg.Settings.UI.ActiveTab)
	}
	if cfg.Repositories == nil {
		t.Error("repositories should be empty slice, not nil")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("default config not written: %v", err)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	original := DefaultConfig()
	original.Settings.UI.ActiveTab = "tasks"
	original.Settings.UI.SplitViewEnabled = true
	original.Settings.UI.Panes = []PaneState{
		{RepoID: "mesosphere-main", Mode: "docs", ActiveDocPath: "docs/architecture.md"},
		{RepoID: "mesosphere-main", Mode: "tasks"},
	}
	original.Repositories = []RepositoryConfig{
		{
			ID:        "mesosphere-main",
			Name:      "Mesosphere Main Repo",
			Path:      "/path/to/local/repo",
			RemoteURL: "git@github.com:org/repo.git",
			Credentials: CredentialsConfig{
				SSHKeyPath: "~/.ssh/id_rsa",
			},
		},
		{
			ID:   "secondary-repo",
			Name: "Secondary Project",
			Path: "/path/to/another/repo",
		},
	}

	if err := SaveConfig(path, original); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if loaded.Settings.UI.ActiveTab != "tasks" {
		t.Errorf("active_tab = %q, want tasks", loaded.Settings.UI.ActiveTab)
	}
	if !loaded.Settings.UI.SplitViewEnabled {
		t.Error("expected split_view_enabled true")
	}
	if len(loaded.Repositories) != 2 {
		t.Fatalf("repositories = %d, want 2", len(loaded.Repositories))
	}
	if loaded.Repositories[0].Credentials.SSHKeyPath != "~/.ssh/id_rsa" {
		t.Errorf("ssh_key_path = %q", loaded.Repositories[0].Credentials.SSHKeyPath)
	}
	if len(loaded.Settings.UI.Panes) != 2 {
		t.Fatalf("panes = %d, want 2", len(loaded.Settings.UI.Panes))
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("settings: [unterminated"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSaveConfigNil(t *testing.T) {
	if err := SaveConfig(filepath.Join(t.TempDir(), "c.yaml"), nil); err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestDefaultPathUsesMesosphereHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MESOSPHERE_HOME", home)
	got, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "config.yaml")
	if got != want {
		t.Errorf("DefaultPath = %q, want %q", got, want)
	}
}
