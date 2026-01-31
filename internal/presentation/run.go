package presentation

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nezdemkovski/folio212/internal/domain/run"
	"github.com/nezdemkovski/folio212/internal/shared/ui"
)

type runStep int

const (
	runStepWorking runStep = iota
	runStepComplete
)

type RunModel struct {
	step    runStep
	spinner spinner.Model

	manager *run.Manager

	result *run.Result
	err    error
}

type runResultMsg struct {
	result *run.Result
}

type runErrorMsg struct {
	err error
}

func NewRunModel(manager *run.Manager) RunModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = ui.SpinnerStyle

	return RunModel{
		step:    runStepWorking,
		spinner: s,
		manager: manager,
	}
}

func (m RunModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.run())
}

func (m RunModel) run() tea.Cmd {
	return func() tea.Msg {
		result, err := m.manager.Run(context.Background())
		if err != nil {
			return runErrorMsg{err: err}
		}
		return runResultMsg{result: result}
	}
}

func (m RunModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	case runErrorMsg:
		m.err = msg.err
		return m, tea.Quit
	case runResultMsg:
		m.result = msg.result
		m.step = runStepComplete
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m RunModel) View() string {
	if m.err != nil || m.step == runStepComplete {
		return ""
	}

	var s strings.Builder
	s.WriteString(ui.Title.Render("Running"))
	s.WriteString("\n\n")
	s.WriteString(m.spinner.View() + " " + ui.Value.Render("Executing operation..."))
	return ui.Container.Render(s.String())
}

func (m RunModel) Error() error {
	return m.err
}

func (m RunModel) Result() *run.Result {
	return m.result
}

func RenderRunCompletion(result *run.Result) string {
	var s strings.Builder

	s.WriteString(ui.SuccessStyle.Render(ui.SymbolDone) + " " + ui.Title.Render("Run Complete"))
	s.WriteString("\n\n")

	if result != nil {
		if result.Environment != "" {
			s.WriteString(ui.SectionHeader("Context"))
			s.WriteString("\n")
			s.WriteString(ui.Bullet("environment: " + result.Environment))
			s.WriteString("\n\n")
		}

		s.WriteString(ui.SectionHeader("Completed"))
		s.WriteString("\n")
		for _, item := range result.Completed {
			s.WriteString(ui.Bullet(item))
			s.WriteString("\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(ui.Meta.Render("Commands:") + "\n")
	s.WriteString(ui.Bullet("folio212 run   - Run again") + "\n")

	return ui.Container.Render(s.String())
}
