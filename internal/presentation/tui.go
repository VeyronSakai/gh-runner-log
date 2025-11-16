package presentation

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	"github.com/VeyronSakai/gh-runner-log/internal/usecase"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type jobItem struct {
	job *entity.Job
}

func (i jobItem) FilterValue() string {
	return i.job.Name
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(jobItem)
	if !ok {
		return
	}

	job := i.job
	startedAt := "-"
	if job.StartedAt != nil {
		startedAt = job.StartedAt.Local().Format("2006-01-02 15:04:05 MST")
	}

	duration := "-"
	if job.CompletedAt != nil && job.StartedAt != nil {
		d := job.GetExecutionDuration()
		duration = formatDuration(d)
	} else if job.StartedAt != nil && job.Status == "in_progress" {
		d := time.Since(*job.StartedAt)
		duration = formatDuration(d) + " (running)"
	}

	conclusion := job.Conclusion
	if conclusion == "" {
		conclusion = "-"
	}

	str := fmt.Sprintf("%d | %s | %s | %s | %s | %s",
		job.ID,
		truncate(job.WorkflowName, 20),
		job.Status,
		conclusion,
		startedAt,
		duration,
	)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("→ " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list     list.Model
	history  *usecase.RunnerJobHistory
	choice   *entity.Job
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m model) View() string {
	if m.quitting {
		return quitTextStyle.Render("")
	}

	header := fmt.Sprintf("Runner: %s (ID: %d)\nStatus: %s | OS: %s | Labels: %s\n\n",
		m.history.Runner.Name,
		m.history.Runner.ID,
		m.history.Runner.Status,
		m.history.Runner.OS,
		strings.Join(m.history.Runner.Labels, ", "),
	)

	return header + m.list.View()
}

// TUI represents the terminal UI for displaying runner job history
type TUI struct {
	runnerLogger *usecase.RunnerLogger
}

// NewTUI creates a new TUI with the given usecase
func NewTUI(runnerLogger *usecase.RunnerLogger) *TUI {
	return &TUI{
		runnerLogger: runnerLogger,
	}
}

// Run fetches runner job history and displays the interactive UI
func (t *TUI) Run(ctx context.Context, owner, repo, org, runnerName string, limit int) error {
	// Fetch runner job history
	history, err := t.runnerLogger.FetchRunnerJobHistory(ctx, owner, repo, org, runnerName, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch runner job history: %w", err)
	}

	// Display interactive UI
	items := make([]list.Item, len(history.Jobs))
	for i, job := range history.Jobs {
		items[i] = jobItem{job: job}
	}

	const defaultWidth = 120
	const listHeight = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = fmt.Sprintf("Job History (%d jobs)", len(history.Jobs))
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{list: l, history: history}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running interactive UI: %w", err)
	}

	if m, ok := finalModel.(model); ok && m.choice != nil {
		return openBrowser(m.choice.HtmlUrl)
	}

	return nil
}

func openBrowser(url string) error {
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
		return fmt.Errorf("unsupported platform")
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// truncate truncates a string to the specified length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatDuration formats a duration in a human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
