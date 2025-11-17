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
func (c *Controller) Run(ctx context.Context, runnerName string, maxCount int) error {
	// Create model in loading state
	m := newLoadingModel(c.runnerLogger, runnerName, maxCount)

	// Run TUI
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running interactive UI: %w", err)
	}

	// Check if the model has an error from data loading
	if model, ok := finalModel.(*Model); ok && model.err != nil {
		return model.err
	}

	return nil
}

// newLoadingModel creates a model in loading state that will fetch data
func newLoadingModel(runnerLogger *usecase.RunnerLogger, runnerName string, maxCount int) *Model {
	m := NewModel(nil) // nil history means loading
	m.runnerLogger = runnerLogger
	m.runnerName = runnerName
	m.maxCount = maxCount
	return m
}
