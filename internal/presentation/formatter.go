package presentation

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/VeyronSakai/gh-runner-log/internal/usecase"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	// FormatTable outputs data in a human-readable table format
	FormatTable OutputFormat = "table"
	// FormatJSON outputs data in JSON format
	FormatJSON OutputFormat = "json"
)

// Formatter handles formatting of runner job history
type Formatter struct {
	format OutputFormat
}

// NewFormatter creates a new Formatter
func NewFormatter(format OutputFormat) *Formatter {
	return &Formatter{format: format}
}

// Format formats the runner job history based on the configured format
func (f *Formatter) Format(history *usecase.RunnerJobHistory) (string, error) {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(history)
	case FormatTable:
		return f.formatTable(history)
	default:
		return "", fmt.Errorf("unknown output format: %s", f.format)
	}
}

// formatJSON formats the output as JSON
func (f *Formatter) formatJSON(history *usecase.RunnerJobHistory) (string, error) {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// formatTable formats the output as a human-readable table
func (f *Formatter) formatTable(history *usecase.RunnerJobHistory) (string, error) {
	var sb strings.Builder

	// Runner information
	sb.WriteString(fmt.Sprintf("Runner: %s (ID: %d)\n", history.Runner.Name, history.Runner.ID))
	sb.WriteString(fmt.Sprintf("Status: %s\n", history.Runner.Status))
	sb.WriteString(fmt.Sprintf("OS: %s\n", history.Runner.OS))
	sb.WriteString(fmt.Sprintf("Labels: %s\n", strings.Join(history.Runner.Labels, ", ")))
	sb.WriteString("\n")

	// Job history
	if len(history.Jobs) == 0 {
		sb.WriteString("No jobs found for this runner.\n")
		return sb.String(), nil
	}

	sb.WriteString(fmt.Sprintf("Job History (%d jobs):\n", len(history.Jobs)))
	sb.WriteString(strings.Repeat("=", 120) + "\n")
	sb.WriteString(fmt.Sprintf("%-12s %-20s %-15s %-15s %-20s %-30s\n",
		"JOB ID", "WORKFLOW", "STATUS", "CONCLUSION", "STARTED AT", "DURATION"))
	sb.WriteString(strings.Repeat("-", 120) + "\n")

	for _, job := range history.Jobs {
		jobID := fmt.Sprintf("%d", job.ID)
		workflow := truncate(job.WorkflowName, 20)
		status := job.Status
		conclusion := job.Conclusion
		if conclusion == "" {
			conclusion = "-"
		}

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

		sb.WriteString(fmt.Sprintf("%-12s %-20s %-15s %-15s %-20s %-30s\n",
			jobID, workflow, status, conclusion, startedAt, duration))
	}

	sb.WriteString(strings.Repeat("=", 120) + "\n")
	return sb.String(), nil
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
