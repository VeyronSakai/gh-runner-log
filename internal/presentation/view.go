package presentation

import (
	"fmt"
	"io"
	"strings"
	"time"

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

// View renders the model
func (m *Model) View() string {
	if m.quitting {
		return quitTextStyle.Render("")
	}

	header := renderHeader(m.history)
	return header + m.list.View()
}

// renderHeader renders the runner information header
func renderHeader(history *usecase.RunnerJobHistory) string {
	return fmt.Sprintf("Runner: %s\nStatus: %s | OS: %s | Labels: %s\n\n",
		history.Runner.Name,
		history.Runner.Status,
		history.Runner.OS,
		strings.Join(history.Runner.Labels, ", "),
	)
}

// itemDelegate handles rendering of individual job items
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
