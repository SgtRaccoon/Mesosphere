// Package repo resolves the working repository from config, cwd, and CLI flags.
package repo

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mesosphere/mesosphere/pkg/core/config"
)

var (
	// ErrNoContext is returned when cwd is not inside a configured repo and no flag was given.
	ErrNoContext = errors.New("working context could not be resolved: cwd is not a configured repository")
	// ErrRepoNotFound is returned when repoFlag does not match any configured repository.
	ErrRepoNotFound = errors.New("repository not found")
	// ErrAmbiguousContext is returned when cwd matches more than one repository equally.
	ErrAmbiguousContext = errors.New("working context is ambiguous")
)

// ListRepositories returns configured repositories.
func ListRepositories(cfg *config.Config) []config.RepositoryConfig {
	if cfg == nil || cfg.Repositories == nil {
		return []config.RepositoryConfig{}
	}
	return cfg.Repositories
}

// GetRepositoryContext selects a repository.
// If repoFlag is set, it must match a repository ID or path.
// Otherwise cwd must be the configured path or a subdirectory of it.
func GetRepositoryContext(cfg *config.Config, cwd, repoFlag string) (*config.RepositoryConfig, error) {
	if cfg == nil {
		return nil, ErrNoContext
	}
	repos := cfg.Repositories
	if len(repos) == 0 {
		if strings.TrimSpace(repoFlag) != "" {
			return nil, fmt.Errorf("%w: %s", ErrRepoNotFound, repoFlag)
		}
		return nil, ErrNoContext
	}

	if flag := strings.TrimSpace(repoFlag); flag != "" {
		return matchFlag(repos, flag)
	}

	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return nil, fmt.Errorf("resolve cwd: %w", err)
	}
	absCwd = filepath.Clean(absCwd)

	var matches []config.RepositoryConfig
	bestLen := -1
	for _, r := range repos {
		absRepo, err := absPath(r.Path)
		if err != nil {
			continue
		}
		if containsPath(absRepo, absCwd) {
			n := len(absRepo)
			if n > bestLen {
				bestLen = n
				matches = []config.RepositoryConfig{r}
			} else if n == bestLen {
				matches = append(matches, r)
			}
		}
	}
	if len(matches) == 0 {
		return nil, ErrNoContext
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("%w: cwd matches %d repositories", ErrAmbiguousContext, len(matches))
	}
	r := matches[0]
	return &r, nil
}

func matchFlag(repos []config.RepositoryConfig, flag string) (*config.RepositoryConfig, error) {
	flagAbs, flagAbsErr := absPath(flag)
	for i := range repos {
		r := repos[i]
		if r.ID == flag {
			return &r, nil
		}
		if filepath.Clean(r.Path) == filepath.Clean(flag) {
			return &r, nil
		}
		if flagAbsErr == nil {
			repoAbs, err := absPath(r.Path)
			if err == nil && repoAbs == flagAbs {
				return &r, nil
			}
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrRepoNotFound, flag)
}

func absPath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func containsPath(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if strings.HasPrefix(rel, "..") {
		return false
	}
	return true
}

// ErrRepoExists is returned when adding a repository with a duplicate ID.
var ErrRepoExists = errors.New("repository already exists")

// AddRepository appends a repository to cfg after filling defaults.
func AddRepository(cfg *config.Config, r config.RepositoryConfig) (*config.RepositoryConfig, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	r.Path = strings.TrimSpace(r.Path)
	if r.Path == "" {
		return nil, errors.New("path is required")
	}
	abs, err := absPath(r.Path)
	if err != nil {
		return nil, err
	}
	r.Path = abs
	if r.ID == "" {
		r.ID = slugID(r.Name, filepath.Base(r.Path))
	}
	if r.Name == "" {
		r.Name = r.ID
	}
	for _, existing := range cfg.Repositories {
		if existing.ID == r.ID {
			return nil, fmt.Errorf("%w: %s", ErrRepoExists, r.ID)
		}
	}
	cfg.Repositories = append(cfg.Repositories, r)
	return &r, nil
}

func slugID(parts ...string) string {
	for _, p := range parts {
		s := strings.ToLower(strings.TrimSpace(p))
		s = strings.Map(func(r rune) rune {
			switch {
			case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
				return r
			case r == '-' || r == '_' || r == ' ':
				return '-'
			default:
				return -1
			}
		}, s)
		s = strings.Trim(s, "-")
		if s != "" {
			return s
		}
	}
	return "repo"
}
