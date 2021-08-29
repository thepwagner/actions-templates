package cmd

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Synchronize SOPS secrets to every repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		t, err := newTemplater()
		if err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"secrets": len(config.Secrets),
			"repos":   config.RepositoryCount(),
		}).Info("secrets called")

		return t.SetRepoSecrets(ctx)
	},
}

func init() {
	rootCmd.AddCommand(secretsCmd)
	secretsCmd.Flags().String("secrets", "", "Path to the secrets file")
}
