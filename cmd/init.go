package cmd

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/nezdemkovski/folio212/internal/presentation"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long:  "Interactive setup that writes a small config file to your home directory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := tea.NewProgram(
			presentation.NewInitModel(),
		)

		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		if m, ok := finalModel.(*presentation.InitModel); ok {
			if m.Error() != nil {
				return m.Error()
			}
			fmt.Println(presentation.RenderInitCompletion(m.Config(), m.AccountSummary(), m.ValidationWarning(), m.SecretSource()))
		}

		return nil
	},
}
