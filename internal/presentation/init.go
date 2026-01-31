package presentation

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/nezdemkovski/folio212/internal/infrastructure/config"
	"github.com/nezdemkovski/folio212/internal/infrastructure/secrets"
	"github.com/nezdemkovski/folio212/internal/shared/ui"
	"github.com/nezdemkovski/folio212/internal/shared/validation"
)

type InitModel struct {
	form      *huh.Form
	env       string
	workspace string
	apiToken  string
	cancelled bool

	width         int
	height        int
	err           error
	cfg           *config.Config
	layout        ui.Layout
	tokenSource   secrets.Source
	tokenInsecure bool
}

func NewInitModel() *InitModel {
	m := &InitModel{
		env:    "local",
		layout: ui.NewLayout(80, 24),
	}

	confirm := true

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Environment").
				Description("A simple example config field").
				Value(&m.env).
				Placeholder("local"),
			huh.NewInput().
				Title("Workspace (optional)").
				Description("Any path or label you want to persist").
				Value(&m.workspace).
				Placeholder(""),
			huh.NewInput().
				Title("API Token (optional)").
				Description("Sensitive data stored securely in OS keyring").
				Value(&m.apiToken).
				EchoMode(huh.EchoModePassword).
				Placeholder(""),
			huh.NewConfirm().
				Title("Proceed?").
				Value(&confirm).
				Affirmative("OK").
				Negative("Cancel").
				Validate(func(v bool) error {
					if !v {
						m.cancelled = true
						return nil
					}
					return validation.ValidateNonEmpty("environment", m.env)
				}),
		),
	).WithTheme(huh.ThemeBase()).
		WithShowHelp(true)

	return m
}

func (m *InitModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m *InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout.UpdateDimensions(msg.Width, msg.Height)
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateCompleted {
		if m.cancelled {
			m.err = fmt.Errorf("init cancelled by user")
			return m, tea.Quit
		}

		c := config.Default()
		c.Environment = strings.TrimSpace(m.env)
		c.Workspace = strings.TrimSpace(m.workspace)
		if err := config.Save(c); err != nil {
			m.err = err
			return m, tea.Quit
		}

		if token := strings.TrimSpace(m.apiToken); token != "" {
			source, insecure, err := secrets.Set(secrets.KeyAPIToken, token)
			if err != nil {
				m.err = fmt.Errorf("failed to save API token: %w", err)
				return m, tea.Quit
			}
			m.tokenSource = source
			m.tokenInsecure = insecure
		}

		m.cfg = c
		return m, tea.Quit
	}

	return m, cmd
}

func (m *InitModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	sections := []string{
		m.layout.RenderLogo(),
		m.layout.RenderSubtitle("Template setup"),
		m.layout.RenderBody(m.form.View()),
	}

	return m.layout.RenderCentered(sections...)
}

func (m *InitModel) Error() error {
	return m.err
}

func (m *InitModel) Config() *config.Config {
	return m.cfg
}

func RenderInitCompletion(cfg *config.Config) string {
	var s strings.Builder

	s.WriteString(ui.SuccessStyle.Render(ui.SymbolDone) + " " + ui.Title.Render("Initialization Complete"))
	s.WriteString("\n\n")

	if cfg != nil {
		s.WriteString(ui.SectionHeader("Config"))
		s.WriteString("\n")
		s.WriteString(ui.Bullet(fmt.Sprintf("environment: %s", cfg.Environment)))
		s.WriteString("\n")
		if cfg.Workspace != "" {
			s.WriteString(ui.Bullet(fmt.Sprintf("workspace: %s", cfg.Workspace)))
			s.WriteString("\n")
		}
	}

	token, source, _ := secrets.Get(secrets.KeyAPIToken)
	if token != "" {
		s.WriteString("\n")
		s.WriteString(ui.SectionHeader("Secrets"))
		s.WriteString("\n")
		switch source {
		case secrets.SourceKeyring:
			s.WriteString(ui.Bullet("API token stored securely in OS keyring"))
		case secrets.SourceFile:
			s.WriteString(ui.WarningStyle.Render(ui.SymbolWarning) + " " + ui.WarningStyle.Render("API token stored in config file (insecure)"))
			s.WriteString("\n")
			s.WriteString(ui.Meta.Render("  Consider using environment variable for servers:"))
			s.WriteString("\n")
			s.WriteString(ui.Meta.Render("  export APP_API_TOKEN=your-token"))
		case secrets.SourceEnv:
			s.WriteString(ui.Bullet("API token loaded from environment variable"))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(ui.Meta.Render("Next steps:") + "\n")
	s.WriteString(ui.Bullet("folio212 run  - Run a demo operation") + "\n")

	return ui.Container.Render(s.String())
}
