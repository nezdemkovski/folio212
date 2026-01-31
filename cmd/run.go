package cmd

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/nezdemkovski/folio212/internal/domain/run"
	"github.com/nezdemkovski/folio212/internal/presentation"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a demo operation",
	Long:  "Connects to Trading212 and checks your portfolio (TUI-ready).",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()

		manager := run.NewManager(cfg)

		if !isatty.IsTerminal(os.Stdin.Fd()) || !isatty.IsTerminal(os.Stdout.Fd()) {
			result, err := manager.Run(context.Background())
			if err != nil {
				return err
			}
			fmt.Println(presentation.RenderRunCompletion(result))
			return nil
		}

		p := tea.NewProgram(presentation.NewRunModel(manager))
		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		if m, ok := finalModel.(presentation.RunModel); ok {
			if m.Error() != nil {
				return m.Error()
			}
			if result := m.Result(); result != nil {
				fmt.Println(presentation.RenderRunCompletion(result))
			}
		}

		return nil
	},
}
