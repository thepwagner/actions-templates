package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/getsops/sops/v3/decrypt"
	"github.com/google/go-github/v62/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/thepwagner/actions-templates/templater"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

var cfgFile string
var config templater.Config

var rootCmd = &cobra.Command{
	Use:   "actions-templates",
	Short: "Template GitHub Actions",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Fatal("error")
	}
}

func init() {
	cobra.OnInitialize(func() {
		if err := initConfig(); err != nil {
			logrus.WithError(err).Fatal("loading config")
		}
	})
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}

func initConfig() error {
	if cfgFile == "" {
		return fmt.Errorf("config file required")
	}
	b, err := decrypt.File(cfgFile, "yaml")
	if err != nil {
		return fmt.Errorf("decrypting configuration: %w", err)
	}
	if err := yaml.Unmarshal(b, &config); err != nil {
		return fmt.Errorf("unmarshalling configuration: %w", err)
	}
	return nil
}

func newTemplater() (*templater.Templater, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.Auth.GitHub})
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	gh := github.NewClient(oauth2.NewClient(ctx, ts))
	return templater.NewTemplater(gh, &config), nil
}
