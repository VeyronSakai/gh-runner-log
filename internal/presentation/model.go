package presentation

import (
	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	"github.com/VeyronSakai/gh-runner-log/internal/usecase"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the application state for the TUI
type Model struct {
	list     list.Model
	history  *usecase.RunnerJobHistory
	choice   *entity.Job
	quitting bool
}

// NewModel creates a new Model with the given history
func NewModel(history *usecase.RunnerJobHistory) *Model {
	items := make([]list.Item, len(history.Jobs))
	for i, job := range history.Jobs {
		items[i] = jobItem{job: job}
	}

	const defaultWidth = 120
	const listHeight = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Job History"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return &Model{
		list:    l,
		history: history,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(jobItem)
			if ok {
				m.choice = i.job
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// GetChoice returns the selected job, if any
func (m *Model) GetChoice() *entity.Job {
	return m.choice
}

// jobItem represents a job in the list
type jobItem struct {
	job *entity.Job
}

func (i jobItem) FilterValue() string {
	return i.job.Name
}
