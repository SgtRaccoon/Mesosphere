package server

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestFetchServiceTickNonBlocking(t *testing.T) {
	var n atomic.Int32
	started := make(chan struct{})
	fs := &FetchService{
		Interval: 20 * time.Millisecond,
		ListPaths: func() []string {
			return []string{"/tmp/a", "/tmp/b"}
		},
		Fetch: func(path string) error {
			n.Add(1)
			select {
			case <-started:
			default:
				close(started)
			}
			time.Sleep(50 * time.Millisecond)
			return nil
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go fs.Run(ctx)

	done := make(chan struct{})
	go func() {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Error("fetch never started")
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP-equivalent wait blocked too long")
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if n.Load() >= 2 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("fetches = %d, want >= 2", n.Load())
}

func TestFetchServiceTickInvokesAllRepos(t *testing.T) {
	var paths []string
	fs := &FetchService{
		ListPaths: func() []string { return []string{"r1", "r2"} },
		Fetch: func(p string) error {
			paths = append(paths, p)
			return nil
		},
	}
	fs.Tick()
	if len(paths) != 2 {
		t.Fatalf("paths = %v", paths)
	}
}
