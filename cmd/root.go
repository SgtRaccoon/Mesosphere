package cmd

import (
	"github.com/spf13/cobra"
)

// Serve starts HTTP server mode. Tests may replace it.
var Serve = func(addr string, openBrowser ...bool) error {
	open := true
	if len(openBrowser) > 0 {
		open = openBrowser[0]
	}
	return ServeWith(addr, open)
}

// NewRootCommand builds the Cobra command tree.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "mesosphere",
		Short: "Git-native product management CLI and local UI",
		Long:  "Mesosphere unifies CLI, REST API, and an embedded web UI over git-backed docs and tasks.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Serve("", true)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newServeCommand(), newDocCommand(), newTaskCommand(), newRepoCommand())
	return root
}

// Execute runs the CLI. Zero extra args (or `serve`) enter server mode.
func Execute() error {
	return NewRootCommand().Execute()
}
