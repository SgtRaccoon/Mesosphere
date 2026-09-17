package cmd

import (
	"fmt"
	"os"

	"github.com/mesosphere/mesosphere/pkg/core/config"
	"github.com/mesosphere/mesosphere/pkg/core/docs"
	"github.com/mesosphere/mesosphere/pkg/core/repo"
	"github.com/spf13/cobra"
)

var (
	getwd      = os.Getwd
	loadConfig = func() (*config.Config, error) {
		return config.LoadConfig("")
	}
	saveConfig = func(cfg *config.Config) error {
		return config.SaveConfig("", cfg)
	}
	docsEngine = &docs.Engine{}
)

func newDocCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "doc",
		Short: "List and read repository documents",
	}
	c.AddCommand(newDocListCommand(), newDocVersionsCommand(), newDocGetCommand())
	return c
}

func resolveRepo(repoFlag string) (*config.RepositoryConfig, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	cwd, err := getwd()
	if err != nil {
		return nil, err
	}
	return repo.GetRepositoryContext(cfg, cwd, repoFlag)
}

func newDocListCommand() *cobra.Command {
	var repoFlag string
	c := &cobra.Command{
		Use:   "list",
		Short: "List documents in the working repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}
			list, err := docsEngine.ListDocuments(r.Path)
			if err != nil {
				return err
			}
			if len(list) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no documents)")
				return nil
			}
			for _, d := range list {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", d.Path, d.Title)
			}
			return nil
		},
	}
	c.Flags().StringVar(&repoFlag, "repo", "", "repository id or path")
	return c
}

func newDocVersionsCommand() *cobra.Command {
	var repoFlag string
	c := &cobra.Command{
		Use:   "versions <doc-path>",
		Short: "List git versions of a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}
			vers, err := docsEngine.ListDocumentVersions(r.Path, args[0])
			if err != nil {
				return err
			}
			for _, v := range vers {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", v.Hash, v.Timestamp.UTC().Format("2006-01-02"), v.Message)
			}
			return nil
		},
	}
	c.Flags().StringVar(&repoFlag, "repo", "", "repository id or path")
	return c
}

func newDocGetCommand() *cobra.Command {
	var repoFlag, version string
	c := &cobra.Command{
		Use:   "get <doc-path>",
		Short: "Print document contents (optional historical version)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}
			doc, err := docsEngine.GetDocument(r.Path, args[0], version)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), doc.Content)
			return nil
		},
	}
	c.Flags().StringVar(&repoFlag, "repo", "", "repository id or path")
	c.Flags().StringVar(&version, "version", "", "commit hash (default: working tree)")
	return c
}
