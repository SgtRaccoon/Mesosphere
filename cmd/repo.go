package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/mesosphere/mesosphere/pkg/core/config"
	"github.com/mesosphere/mesosphere/pkg/core/repo"
	"github.com/spf13/cobra"
)

func newRepoCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "repo",
		Short: "Manage configured repositories",
	}
	c.AddCommand(newRepoListCommand())
	c.AddCommand(newRepoAddCommand())
	return c
}

func newRepoAddCommand() *cobra.Command {
	var id, name, path, remote string
	c := &cobra.Command{
		Use:   "add",
		Short: "Add a repository to the config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			added, err := repo.AddRepository(cfg, config.RepositoryConfig{
				ID:        id,
				Name:      name,
				Path:      path,
				RemoteURL: remote,
			})
			if err != nil {
				return err
			}
			if err := saveConfig(cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "added %s (%s)\n", added.ID, added.Path)
			return nil
		},
	}
	c.Flags().StringVar(&id, "id", "", "repository id")
	c.Flags().StringVar(&name, "name", "", "display name")
	c.Flags().StringVar(&path, "path", "", "filesystem path (required)")
	c.Flags().StringVar(&remote, "remote", "", "remote url")
	_ = c.MarkFlagRequired("path")
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
