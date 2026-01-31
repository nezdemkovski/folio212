package cmd

import (
	"fmt"

	"github.com/nezdemkovski/folio212/internal/infrastructure/config"
	"github.com/nezdemkovski/folio212/internal/shared/ui"
	"github.com/spf13/cobra"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "folio212",
	Short: "Trading212 portfolio checker",
	Long:  "Connects to Trading212 and checks your portfolio from the terminal.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Commands that must work without prior setup / config file.
		if cmd.Name() == "init" || cmd.Name() == "skill" {
			return nil
		}

		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("configuration not found. Please run 'folio212 init' first: %w", err)
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
	rootCmd.AddCommand(portfolioCmd)
	rootCmd.AddCommand(skillCmd)
}

func GetConfig() *config.Config {
	return cfg
}
