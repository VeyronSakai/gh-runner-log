package presentation

import (
	"context"
	"os/exec"
	"runtime"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	"github.com/VeyronSakai/gh-runner-log/internal/usecase"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Column width constants
const (
	minWorkflowWidth    = 20
	minJobWidth         = 20
	statusWidth         = 12
	conclusionWidth     = 12
	startedAtWidth      = 25
	durationWidth       = 15
	borderPadding       = 10
	headerFooterHeight  = 8
	defaultTableHeight  = 20
	defaultTerminalWidth = 120

	// Proportions for distributing extra width (only for Workflow and Job)
	ratioWorkflow  = 0.5
	ratioJob       = 0.5
)

// Model represents the application state for the TUI
type Model struct {
	table        table.Model
	spinner      spinner.Model
	history      *usecase.RunnerJobHistory
	choice       *entity.Job
	loading      bool
	quitting     bool
	runnerLogger *usecase.RunnerLogger
	runnerName   string
	limit        int
	width        int
	height       int
	err          error
}

// historyLoadedMsg is sent when history is loaded
type historyLoadedMsg struct {
	history *usecase.RunnerJobHistory
	err     error
}

// NewModel creates a new Model with the given history (can be nil during loading)
func NewModel(history *usecase.RunnerJobHistory) *Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	var t table.Model
	loading := history == nil

	if !loading {
		columns := []table.Column{
			{Title: "ID", Width: 10},
			{Title: "Workflow", Width: 20},
			{Title: "Status", Width: 12},
			{Title: "Conclusion", Width: 12},
			{Title: "Started At", Width: 25},
			{Title: "Duration", Width: 15},
		}

		rows := buildRows(history.Jobs)

		t = table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(20),
		)

		ts := table.DefaultStyles()
		ts.Header = ts.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		ts.Selected = ts.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		t.SetStyles(ts)
		t.Focus()
	}

	return &Model{
		table:   t,
		spinner: s,
		history: history,
		loading: loading,
		width:   defaultTerminalWidth,
		height:  defaultTableHeight + headerFooterHeight,
	}
}

// getCalculatedColumnWidths calculates column widths based on available terminal width
func getCalculatedColumnWidths(terminalWidth int) []table.Column {
	if terminalWidth == 0 {
		terminalWidth = defaultTerminalWidth
	}

	availableWidth := terminalWidth - borderPadding
	fixedWidth := statusWidth + conclusionWidth + startedAtWidth + durationWidth
	totalMinWidth := minWorkflowWidth + minJobWidth + fixedWidth

	if availableWidth < totalMinWidth {
		// Terminal is too small, use minimum widths
		return []table.Column{
			{Title: "Workflow", Width: minWorkflowWidth},
			{Title: "Job", Width: minJobWidth},
			{Title: "Status", Width: statusWidth},
			{Title: "Conclusion", Width: conclusionWidth},
			{Title: "Started At", Width: startedAtWidth},
			{Title: "Duration", Width: durationWidth},
		}
	}

	remainingWidth := availableWidth - totalMinWidth

	// Distribute remaining width proportionally between Workflow and Job
	workflowExtra := int(float64(remainingWidth) * ratioWorkflow)
	jobExtra := int(float64(remainingWidth) * ratioJob)

	return []table.Column{
		{Title: "Workflow", Width: minWorkflowWidth + workflowExtra},
		{Title: "Job", Width: minJobWidth + jobExtra},
		{Title: "Status", Width: statusWidth},
		{Title: "Conclusion", Width: conclusionWidth},
		{Title: "Started At", Width: startedAtWidth},
		{Title: "Duration", Width: durationWidth},
	}
}

// getCalculatedTableHeight calculates table height based on terminal height
func getCalculatedTableHeight(terminalHeight int) int {
	if terminalHeight == 0 {
		return defaultTableHeight
	}
	height := terminalHeight - headerFooterHeight
	if height < 5 {
		return 5
	}
	return height
}

// updateTableDimensions updates the table dimensions based on current terminal size
func (m *Model) updateTableDimensions() {
	columns := getCalculatedColumnWidths(m.width)
	m.table.SetColumns(columns)
	
	tableHeight := getCalculatedTableHeight(m.height)
	m.table.SetHeight(tableHeight)
}

func (m *Model) Init() tea.Cmd {
	if m.loading {
		return tea.Batch(
			m.spinner.Tick,
			m.fetchHistory(),
		)
	}
	return nil
}

// fetchHistory fetches the runner job history
func (m *Model) fetchHistory() tea.Cmd {
	return func() tea.Msg {
		history, err := m.runnerLogger.FetchRunnerJobHistory(context.Background(), m.runnerName, m.limit)
		return historyLoadedMsg{history: history, err: err}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.loading {
			m.updateTableDimensions()
		}
		return m, nil

	case historyLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.loading = false
			return m, tea.Quit
		}
		m.history = msg.history
		m.loading = false
		
		// Build table now that we have data
		columns := getCalculatedColumnWidths(m.width)
		rows := buildRows(m.history.Jobs)

		tableHeight := getCalculatedTableHeight(m.height)
		m.table = table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(tableHeight),
		)

		ts := table.DefaultStyles()
		ts.Header = ts.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		ts.Selected = ts.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		m.table.SetStyles(ts)
		m.table.Focus()
		return m, nil
		
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if !m.loading {
				// Get selected job
				selectedIdx := m.table.Cursor()
				if selectedIdx >= 0 && selectedIdx < len(m.history.Jobs) {
					m.choice = m.history.Jobs[selectedIdx]
					// Open browser but don't quit
					go openBrowserAsync(m.choice.HtmlUrl)
				}
			}
		}
	}
	
	if m.loading {
		m.spinner, cmd = m.spinner.Update(msg)
	} else {
		m.table, cmd = m.table.Update(msg)
	}
	return m, cmd
}

// GetChoice returns the selected job, if any
func (m *Model) GetChoice() *entity.Job {
	return m.choice
}

// openBrowserAsync opens a URL in the browser asynchronously
func openBrowserAsync(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		return
	}

	args = append(args, url)
	exec.Command(cmd, args...).Start()
}


