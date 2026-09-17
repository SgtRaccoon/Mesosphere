package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Server is the embedded HTTP engine.
type Server struct {
	Addr     string
	Handler  http.Handler
	httpSrv  *http.Server
	listener net.Listener
	started  chan struct{}
	Fetch    *FetchService
}

// New constructs a server. Empty addr uses :8080.
func New(addr string) *Server {
	if addr == "" {
		addr = ":8080"
	}
	return &Server{Addr: addr, Handler: NewRouter(), started: make(chan struct{}), Fetch: defaultFetchService()}
}

// Start listens and serves until ctx is cancelled or Shutdown is called.
func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.listener = ln
	s.Addr = ln.Addr().String()
	s.httpSrv = &http.Server{
		Handler:           s.Handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	close(s.started)
	fetchCtx, fetchCancel := context.WithCancel(context.Background())
	defer fetchCancel()
	go s.Fetch.Run(fetchCtx)
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.httpSrv.Serve(ln)
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.httpSrv.Shutdown(shutdownCtx)
		err := <-errCh
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

// URL is the base URL after Start binds a port.
func (s *Server) URL() string {
	return "http://" + s.Addr
}

// Started is closed once the listener is bound.
func (s *Server) Started() <-chan struct{} {
	return s.started
}
