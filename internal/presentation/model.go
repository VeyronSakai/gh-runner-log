package presentation

import (
	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	"github.com/VeyronSakai/gh-runner-log/internal/usecase"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the application state for the TUI
type Model struct {
	table    table.Model
	history  *usecase.RunnerJobHistory
	choice   *entity.Job
	quitting bool
}

// NewModel creates a new Model with the given history
func NewModel(history *usecase.RunnerJobHistory) *Model {
	columns := []table.Column{
		{Title: "ID", Width: 10},
		{Title: "Workflow", Width: 20},
		{Title: "Status", Width: 12},
		{Title: "Conclusion", Width: 12},
		{Title: "Started At", Width: 25},
		{Title: "Duration", Width: 15},
	}

	rows := buildRows(history.Jobs)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	t.Focus()

	return &Model{
		table:   t,
		history: history,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			// Get selected job
			selectedIdx := m.table.Cursor()
			if selectedIdx >= 0 && selectedIdx < len(m.history.Jobs) {
				m.choice = m.history.Jobs[selectedIdx]
				m.quitting = true
				return m, tea.Quit
			}
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// GetChoice returns the selected job, if any
func (m *Model) GetChoice() *entity.Job {
	return m.choice
}


