// Package config loads and persists Mesosphere configuration (~/.mesosphere/config.yaml).
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultVersion = "1.0"
	DefaultPort    = 8080
	configFileName = "config.yaml"
	appDirName     = ".mesosphere"
)

// Config is the root configuration document.
type Config struct {
	Version      string             `yaml:"version" json:"version"`
	Settings     Settings           `yaml:"settings" json:"settings"`
	Repositories []RepositoryConfig `yaml:"repositories" json:"repositories"`
}

// Settings holds server and UI preferences.
type Settings struct {
	Server ServerSettings `yaml:"server" json:"server"`
	UI     UISettings     `yaml:"ui" json:"ui"`
}

// ServerSettings controls the embedded HTTP server.
type ServerSettings struct {
	Port            int  `yaml:"port" json:"port"`
	AutoOpenBrowser bool `yaml:"auto_open_browser" json:"auto_open_browser"`
}

// UISettings persists UI state across sessions.
type UISettings struct {
	ActiveTab        string      `yaml:"active_tab" json:"active_tab"`
	SplitViewEnabled bool        `yaml:"split_view_enabled" json:"split_view_enabled"`
	Panes            []PaneState `yaml:"panes" json:"panes"`
}

// PaneState describes one split-view pane.
type PaneState struct {
	RepoID        string `yaml:"repo_id" json:"repo_id"`
	Mode          string `yaml:"mode" json:"mode"` // docs | tasks
	ActiveDocPath string `yaml:"active_doc_path,omitempty" json:"active_doc_path,omitempty"`
}

// RepositoryConfig describes a registered local git repository.
type RepositoryConfig struct {
	ID          string            `yaml:"id" json:"id"`
	Name        string            `yaml:"name" json:"name"`
	Path        string            `yaml:"path" json:"path"`
	RemoteURL   string            `yaml:"remote_url,omitempty" json:"remote_url,omitempty"`
	Credentials CredentialsConfig `yaml:"credentials,omitempty" json:"credentials,omitempty"`
}

// CredentialsConfig holds optional git credentials for a repository.
type CredentialsConfig struct {
	SSHKeyPath string `yaml:"ssh_key_path,omitempty" json:"ssh_key_path,omitempty"`
	Token      string `yaml:"token,omitempty" json:"token,omitempty"`
}

// DefaultPath returns ~/.mesosphere/config.yaml (or MESOSPHERE_HOME/config.yaml).
func DefaultPath() (string, error) {
	if home := os.Getenv("MESOSPHERE_HOME"); home != "" {
		return filepath.Join(home, configFileName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, appDirName, configFileName), nil
}

// DefaultConfig returns a new config with built-in defaults.
func DefaultConfig() *Config {
	return &Config{
		Version: DefaultVersion,
		Settings: Settings{
			Server: ServerSettings{
				Port:            DefaultPort,
				AutoOpenBrowser: true,
			},
			UI: UISettings{
				ActiveTab:        "docs",
				SplitViewEnabled: false,
				Panes:            []PaneState{},
			},
		},
		Repositories: []RepositoryConfig{},
	}
}

// LoadConfig reads YAML from path. If path is empty, DefaultPath is used.
// Missing files produce a default config that is written to disk.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		p, err := DefaultPath()
		if err != nil {
			return nil, err
		}
		path = p
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg := DefaultConfig()
			if saveErr := SaveConfig(path, cfg); saveErr != nil {
				return nil, fmt.Errorf("create default config: %w", saveErr)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	applyDefaults(cfg)
	return cfg, nil
}

// SaveConfig serializes cfg to path using an atomic write.
func SaveConfig(path string, cfg *Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	if path == "" {
		p, err := DefaultPath()
		if err != nil {
			return err
		}
		path = p
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), "config-*.yaml.tmp")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp config: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}
	return nil
}

func applyDefaults(cfg *Config) {
	if cfg.Version == "" {
		cfg.Version = DefaultVersion
	}
	if cfg.Settings.Server.Port == 0 {
		cfg.Settings.Server.Port = DefaultPort
	}
	if cfg.Settings.UI.ActiveTab == "" {
		cfg.Settings.UI.ActiveTab = "docs"
	}
	if cfg.Settings.UI.Panes == nil {
		cfg.Settings.UI.Panes = []PaneState{}
	}
	if cfg.Repositories == nil {
		cfg.Repositories = []RepositoryConfig{}
	}
}
