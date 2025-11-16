package presentation

import (
	"fmt"
	"strings"
	"time"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	"github.com/VeyronSakai/gh-runner-log/internal/usecase"
	"github.com/charmbracelet/bubbles/table"
)

// View renders the model
func (m *Model) View() string {
	if m.quitting {
		if m.err != nil {
			return fmt.Sprintf("\nError: %v\n", m.err)
		}
		return ""
	}

	if m.loading {
		return fmt.Sprintf("\n%s Loading runner job history...\n", m.spinner.View())
	}

	header := renderHeader(m.history)
	return header + "\n" + m.table.View()
}

// renderHeader renders the runner information header
func renderHeader(history *usecase.RunnerJobHistory) string {
	return fmt.Sprintf("Runner: %s\nStatus: %s\nOS: %s\nLabels: %s\n",
		history.Runner.Name,
		history.Runner.Status,
		history.Runner.OS,
		strings.Join(history.Runner.Labels, ", "),
	)
}

// buildRows converts jobs to table rows
func buildRows(jobs []*entity.Job) []table.Row {
	rows := make([]table.Row, len(jobs))
	for i, job := range jobs {
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

		rows[i] = table.Row{
			truncate(job.Name, 25),
			truncate(job.WorkflowName, 20),
			job.Status,
			conclusion,
			startedAt,
			duration,
		}
	}
	return rows
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
