package cmd

import (
	"fmt"

	"github.com/nezdemkovski/cli-tool-template/internal/infrastructure/config"
	"github.com/nezdemkovski/cli-tool-template/internal/shared/ui"
	"github.com/spf13/cobra"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "CLI Tool Template",
	Long:  "A minimal, clean-architecture CLI template (Cobra + Bubble Tea) designed to be easy to expand.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "init" {
			return nil
		}

		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("configuration not found. Please run 'app init' first: %w", err)
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ui.ExitWithError("Command failed", err)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(runCmd)
}

func GetConfig() *config.Config {
	return cfg
}
