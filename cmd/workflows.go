package cmd

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var workflowsCmd = &cobra.Command{
	Use:   "workflows",
	Short: "Render and upload workflow files in the default branch",
	RunE: func(*cobra.Command, []string) error {
		t, err := newTemplater()
		if err != nil {
			return err
		}

		ctx := context.Background()
		logrus.WithFields(logrus.Fields{
			"repos": config.RepositoryCount(),
		}).Info("workflows called")
		return t.RenderWorkflows(ctx)
	},
}

func init() {
	rootCmd.AddCommand(workflowsCmd)
}
