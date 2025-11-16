package presentation

import (
	"context"
	"fmt"

	"github.com/VeyronSakai/gh-runner-log/internal/usecase"
	tea "github.com/charmbracelet/bubbletea"
)

// Controller handles the presentation logic and coordinates between model and view
type Controller struct {
	runnerLogger *usecase.RunnerLogger
}

// NewController creates a new Controller with the given usecase
func NewController(runnerLogger *usecase.RunnerLogger) *Controller {
	return &Controller{
		runnerLogger: runnerLogger,
	}
}

// Run fetches runner job history and displays the interactive UI
func (c *Controller) Run(ctx context.Context, runnerName string, limit int) error {
	// Fetch runner job history
	history, err := c.runnerLogger.FetchRunnerJobHistory(ctx, runnerName, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch runner job history: %w", err)
	}

	// Create model
	m := NewModel(history)

	// Run TUI
	p := tea.NewProgram(m)
	_, err = p.Run()
	if err != nil {
		return fmt.Errorf("error running interactive UI: %w", err)
	}

	return nil
}
