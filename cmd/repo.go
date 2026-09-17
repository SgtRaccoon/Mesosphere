package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/mesosphere/mesosphere/pkg/core/repo"
	"github.com/spf13/cobra"
)

func newRepoCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "repo",
		Short: "List configured repositories",
	}
	c.AddCommand(newRepoListCommand())
	return c
}

func newRepoListCommand() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "list",
		Short: "List configured repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			list := repo.ListRepositories(cfg)
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(list)
			}
			if len(list) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no repositories)")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-24s %s\n", "ID", "NAME", "PATH")
			for _, r := range list {
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-24s %s\n", r.ID, r.Name, r.Path)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}
