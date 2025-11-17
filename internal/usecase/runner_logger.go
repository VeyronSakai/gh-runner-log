package usecase

import (
	"context"
	"fmt"
	"sort"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	"github.com/VeyronSakai/gh-runner-log/internal/domain/repository"
)

// RunnerLogger is a use case for fetching job history for a specific runner
type RunnerLogger struct {
	jobRepo    repository.JobRepository
	runnerRepo repository.RunnerRepository
}

// NewRunnerLogger creates a new RunnerLogger use case
func NewRunnerLogger(jobRepo repository.JobRepository, runnerRepo repository.RunnerRepository) *RunnerLogger {
	return &RunnerLogger{
		jobRepo:    jobRepo,
		runnerRepo: runnerRepo,
	}
}

// RunnerJobHistory represents the job history for a specific runner
type RunnerJobHistory struct {
	Runner *entity.Runner
	Jobs   []*entity.Job
}

// FetchRunnerJobHistory fetches job history for a specific runner
func (r *RunnerLogger) FetchRunnerJobHistory(ctx context.Context, runnerName string, limit int) (*RunnerJobHistory, error) {
	// First, fetch the runner to get its ID
	runner, err := r.runnerRepo.FetchRunnerByName(ctx, runnerName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch runner: %w", err)
	}

	// Fetch job history filtered by runner ID
	// The repository will paginate and filter until it gets enough jobs for this runner
	jobs, err := r.jobRepo.FetchJobHistory(ctx, runner.ID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job history: %w", err)
	}

	// Sort jobs by start time (most recent first)
	sort.Slice(jobs, func(i, j int) bool {
		if jobs[i].StartedAt == nil {
			return false
		}
		if jobs[j].StartedAt == nil {
			return true
		}
		return jobs[i].StartedAt.After(*jobs[j].StartedAt)
	})

	return &RunnerJobHistory{
		Runner: runner,
		Jobs:   jobs,
	}, nil
}
