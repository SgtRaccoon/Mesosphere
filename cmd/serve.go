package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/mesosphere/mesosphere/pkg/browser"
	"github.com/mesosphere/mesosphere/pkg/server"
	"github.com/spf13/cobra"
)

func newServeCommand() *cobra.Command {
	var addr string
	var noBrowser bool
	c := &cobra.Command{
		Use:   "serve",
		Short: "Launch the embedded HTTP server and UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			open := !noBrowser
			return Serve(addr, open)
		},
	}
	c.Flags().StringVar(&addr, "addr", "", "listen address (default from config)")
	c.Flags().BoolVar(&noBrowser, "no-browser", false, "do not open the system browser")
	return c
}

// ServeWith starts the HTTP server and optionally opens the browser.
func ServeWith(addr string, openBrowser bool) error {
	autoOpen := openBrowser
	if addr == "" {
		cfg, err := loadConfig()
		if err == nil && cfg != nil {
			port := cfg.Settings.Server.Port
			if port == 0 {
				port = 8080
			}
			addr = fmt.Sprintf("127.0.0.1:%d", port)
			if autoOpen {
				autoOpen = cfg.Settings.Server.AutoOpenBrowser
			}
		} else {
			addr = "127.0.0.1:8080"
		}
	}
	if strings.HasPrefix(addr, ":") {
		addr = "127.0.0.1" + addr
	}
	srv := server.New(addr)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		<-srv.Started()
		fmt.Fprintf(os.Stdout, "mesosphere: listening on %s\n", srv.URL())
		if autoOpen {
			_ = browser.OpenBrowser(srv.URL())
		}
	}()
	return srv.Start(ctx)
}
