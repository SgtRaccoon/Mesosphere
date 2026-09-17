package server

import (
	"context"
	"time"

	"github.com/mesosphere/mesosphere/pkg/core/config"
	"github.com/mesosphere/mesosphere/pkg/git"
)

// FetchInterval is the background git fetch period (DESIGN §3.4).
const FetchInterval = 3 * time.Minute

// FetchService periodically fetches remotes for configured repositories.
type FetchService struct {
	Interval  time.Duration
	Fetch     func(repoPath string) error
	ListPaths func() []string
}

func defaultFetchService() *FetchService {
	g := &git.Client{}
	return &FetchService{
		Interval: FetchInterval,
		Fetch:    g.Fetch,
		ListPaths: func() []string {
			cfg, err := config.LoadConfig("")
			if err != nil || cfg == nil {
				return nil
			}
			out := make([]string, 0, len(cfg.Repositories))
			for _, r := range cfg.Repositories {
				if r.Path != "" {
					out = append(out, r.Path)
				}
			}
			return out
		},
	}
}

// Run ticks until ctx is cancelled. Each tick starts fetches in a new goroutine
// so HTTP serving is not blocked.
func (f *FetchService) Run(ctx context.Context) {
	if f == nil {
		return
	}
	interval := f.Interval
	if interval <= 0 {
		interval = FetchInterval
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			go f.Tick()
		}
	}
}

// Tick fetches every configured repo path.
func (f *FetchService) Tick() {
	if f == nil || f.Fetch == nil || f.ListPaths == nil {
		return
	}
	for _, p := range f.ListPaths() {
		_ = f.Fetch(p)
	}
}
