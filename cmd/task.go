package cmd

import (
	"encoding/json"

	"github.com/mesosphere/mesosphere/pkg/core/tasks"
	"github.com/spf13/cobra"
)

var listTasks = tasks.ListTasks
var getTask = tasks.GetTask

func newTaskCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "task",
		Short: "List and read repository tasks",
	}
	c.AddCommand(newTaskListCommand(), newTaskGetCommand())
	return c
}

func newTaskListCommand() *cobra.Command {
	var repoFlag string
	c := &cobra.Command{
		Use:   "list",
		Short: "List tasks in the working repository (JSON)",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}
			list, err := listTasks(r.Path, r.ID)
			if err != nil {
				return err
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(list)
		},
	}
	c.Flags().StringVar(&repoFlag, "repo", "", "repository id or path")
	return c
}

func newTaskGetCommand() *cobra.Command {
	var repoFlag string
	c := &cobra.Command{
		Use:   "get <task-id>",
		Short: "Get a task by id (JSON)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}
			t, err := getTask(r.Path, r.ID, args[0])
			if err != nil {
				return err
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(t)
		},
	}
	c.Flags().StringVar(&repoFlag, "repo", "", "repository id or path")
	return c
}
