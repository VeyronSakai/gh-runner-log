package presentation

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

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
	// Create model in loading state
	m := newLoadingModel(c.runnerLogger, runnerName, limit)

	// Run TUI
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running interactive UI: %w", err)
	}

	// Handle selection
	if finalModel, ok := finalModel.(*Model); ok {
		if choice := finalModel.GetChoice(); choice != nil {
			return openBrowser(choice.HtmlUrl)
		}
	}

	return nil
}

// newLoadingModel creates a model in loading state that will fetch data
func newLoadingModel(runnerLogger *usecase.RunnerLogger, runnerName string, limit int) *Model {
	m := NewModel(nil) // nil history means loading
	m.runnerLogger = runnerLogger
	m.runnerName = runnerName
	m.limit = limit
	return m
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
