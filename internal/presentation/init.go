package presentation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/nezdemkovski/folio212/internal/infrastructure/config"
	"github.com/nezdemkovski/folio212/internal/infrastructure/secrets"
	"github.com/nezdemkovski/folio212/internal/infrastructure/trading212"
	"github.com/nezdemkovski/folio212/internal/shared/ui"
	"github.com/nezdemkovski/folio212/internal/shared/validation"
)

type InitModel struct {
	form              *huh.Form
	t212Env           string
	t212KeyID         string
	t212Secret        string
	validateNow       bool
	cancelled         bool
	hasSavedSecret    bool
	validationWarning error

	width          int
	height         int
	err            error
	cfg            *config.Config
	accountSummary *trading212.AccountSummary
	layout         ui.Layout
	secretSource   secrets.Source
	secretInsecure bool
}

func NewInitModel() *InitModel {
	m := &InitModel{
		t212Env:     "demo",
		validateNow: true,
		layout:      ui.NewLayout(80, 24),
	}

	// Prefill values if config already exists.
	if cfg, err := config.Load(); err == nil && cfg != nil {
		if strings.TrimSpace(cfg.Trading212Env) != "" {
			m.t212Env = strings.TrimSpace(cfg.Trading212Env)
		}
		if strings.TrimSpace(cfg.Trading212APIKey) != "" {
			m.t212KeyID = strings.TrimSpace(cfg.Trading212APIKey)
		}
	}

	// Check if a secret is already stored so we can allow "leave blank to keep existing".
	if secret, _, _ := secrets.Get(secrets.KeyTrading212APISecret); strings.TrimSpace(secret) != "" {
		m.hasSavedSecret = true
	}

	secretPlaceholder := ""
	if m.hasSavedSecret {
		secretPlaceholder = "leave blank to keep existing"
	}

	confirm := true

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Trading212 environment").
				Description("Use demo first. Switch to live only when you're ready.").
				Options(
					huh.NewOption("Demo (paper trading)", "demo"),
					huh.NewOption("Live (real money)", "live"),
				).
				Value(&m.t212Env),
			huh.NewInput().
				Title("Trading212 API key (required)").
				Value(&m.t212KeyID).
				Placeholder(""),
			huh.NewInput().
				Title("Trading212 API secret (required)").
				Value(&m.t212Secret).
				EchoMode(huh.EchoModePassword).
				Placeholder(secretPlaceholder),
			huh.NewConfirm().
				Title("Validate credentials now?").
				Description("Recommended. We'll call Trading212 to verify your key + secret.").
				Value(&m.validateNow).
				Affirmative("Yes").
				Negative("Skip"),
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
					if err := validation.ValidateNonEmpty("trading212 api key", strings.TrimSpace(m.t212KeyID)); err != nil {
						return err
					}
					if strings.TrimSpace(m.t212Secret) == "" && m.hasSavedSecret {
						return nil
					}
					return validation.ValidateNonEmpty("trading212 api secret", strings.TrimSpace(m.t212Secret))
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
		c.Trading212Env = strings.TrimSpace(strings.ToLower(m.t212Env))
		c.Trading212APIKey = strings.TrimSpace(m.t212KeyID)
		if c.Trading212Env != "demo" && c.Trading212Env != "live" {
			m.err = fmt.Errorf("invalid trading212 environment %q (expected demo or live)", c.Trading212Env)
			return m, tea.Quit
		}
		if err := validation.ValidateNonEmpty("trading212 api key", c.Trading212APIKey); err != nil {
			m.err = err
			return m, tea.Quit
		}

		secret := strings.TrimSpace(m.t212Secret)
		if secret == "" && m.hasSavedSecret {
			// Reuse previously stored secret (we don't prefill the field).
			prev, _, _ := secrets.Get(secrets.KeyTrading212APISecret)
			secret = strings.TrimSpace(prev)
		}
		if err := validation.ValidateNonEmpty("trading212 api secret", secret); err != nil {
			m.err = err
			return m, tea.Quit
		}

		if m.validateNow {
			baseURL := trading212.BaseURLDemo
			if c.Trading212Env == "live" {
				baseURL = trading212.BaseURLLive
			}

			client, err := trading212.NewClient(baseURL, c.Trading212APIKey, secret)
			if err != nil {
				m.err = err
				return m, tea.Quit
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			summary, err := client.GetAccountSummary(ctx)
			if err != nil {
				// Don't block init. Save credentials and show a warning instead.
				m.validationWarning = fmt.Errorf("validation failed: %w", humanizeTrading212AuthError(err))
			} else {
				m.accountSummary = summary
			}
		}

		if err := config.Save(c); err != nil {
			m.err = err
			return m, tea.Quit
		}

		// Only store when user provided a new secret, or when we didn't have one saved already.
		if strings.TrimSpace(m.t212Secret) != "" || !m.hasSavedSecret {
			source, insecure, err := secrets.Set(secrets.KeyTrading212APISecret, secret)
			if err != nil {
				m.err = fmt.Errorf("failed to save Trading212 API secret: %w", err)
				return m, tea.Quit
			}
			m.secretSource = source
			m.secretInsecure = insecure
			m.hasSavedSecret = true
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

	help := strings.Join([]string{
		"How to get Trading212 API keys:",
		"  Trading212 → Settings → API (Beta) → Generate API key",
		"  Permissions: Account data + Portfolio",
		"  Secret key is shown once — store it securely.",
		"  Help: https://helpcentre.trading212.com/hc/en-us/articles/14584770928157-Trading-212-API-key",
	}, "\n")

	sections := []string{
		m.layout.RenderLogo(),
		m.layout.RenderSubtitle("Check your Trading212 holdings from the terminal. AI ready."),
		m.layout.RenderBody(ui.Meta.Render(help) + "\n\n" + m.form.View()),
	}

	return m.layout.RenderCentered(sections...)
}

func (m *InitModel) Error() error {
	return m.err
}

func (m *InitModel) Config() *config.Config {
	return m.cfg
}

func (m *InitModel) AccountSummary() *trading212.AccountSummary {
	return m.accountSummary
}

func (m *InitModel) ValidationWarning() error {
	return m.validationWarning
}

func humanizeTrading212AuthError(err error) error {
	var httpErr *trading212.HTTPError
	if errors.As(err, &httpErr) && httpErr.StatusCode == 403 {
		// For account summary validation we need the "Account data" permission.
		return fmt.Errorf("%w (missing permission: enable \"Account data\" for your Trading212 API key)", err)
	}
	return err
}

func (m *InitModel) SecretSource() secrets.Source {
	return m.secretSource
}

func RenderInitCompletion(cfg *config.Config, summary *trading212.AccountSummary, validationWarning error, secretSource secrets.Source) string {
	var s strings.Builder

	s.WriteString(ui.SuccessStyle.Render(ui.SymbolDone) + " " + ui.Title.Render("Initialization Complete"))
	s.WriteString("\n\n")

	if cfg != nil {
		s.WriteString(ui.SectionHeader("Config"))
		s.WriteString("\n")
		if cfg.Trading212Env != "" {
			s.WriteString(ui.Bullet(fmt.Sprintf("trading212 env: %s", cfg.Trading212Env)))
			s.WriteString("\n")
		}
		if cfg.Trading212APIKey != "" {
			s.WriteString(ui.Bullet(fmt.Sprintf("trading212 api key: %s", cfg.Trading212APIKey)))
			s.WriteString("\n")
		}
	}

	if summary != nil {
		s.WriteString("\n")
		s.WriteString(ui.SectionHeader("Trading212"))
		s.WriteString("\n")
		s.WriteString(ui.Bullet(fmt.Sprintf("account id: %d", summary.ID)))
		s.WriteString("\n")
		if summary.Currency != "" {
			s.WriteString(ui.Bullet(fmt.Sprintf("currency: %s", summary.Currency)))
			s.WriteString("\n")
		}
		s.WriteString(ui.Bullet(fmt.Sprintf("total value: %.2f", summary.TotalValue)))
		s.WriteString("\n")
	}

	if validationWarning != nil {
		s.WriteString("\n")
		s.WriteString(ui.SectionHeader("Validation"))
		s.WriteString("\n")
		s.WriteString(ui.WarningStyle.Render(ui.SymbolWarning) + " " + ui.WarningStyle.Render(validationWarning.Error()))
		s.WriteString("\n")
	}

	if secretSource != secrets.SourceNone {
		s.WriteString("\n")
		s.WriteString(ui.SectionHeader("Secrets"))
		s.WriteString("\n")
		switch secretSource {
		case secrets.SourceKeyring:
			s.WriteString(ui.Bullet("Trading212 API secret stored securely in OS keyring"))
		case secrets.SourceFile:
			s.WriteString(ui.WarningStyle.Render(ui.SymbolWarning) + " " + ui.WarningStyle.Render("Trading212 API secret stored in config file (insecure)"))
			s.WriteString("\n")
			s.WriteString(ui.Meta.Render("  Consider using environment variable for servers:"))
			s.WriteString("\n")
			s.WriteString(ui.Meta.Render("  export FOLIO212_T212_API_SECRET=your-secret"))
		case secrets.SourceEnv:
			s.WriteString(ui.Bullet("Trading212 API secret loaded from environment variable"))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(ui.Meta.Render("Next steps:") + "\n")
	s.WriteString(ui.Bullet("folio212 portfolio  - Show current holdings") + "\n")

	return ui.Container.Render(s.String())
}
