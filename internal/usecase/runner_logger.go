package usecase

import (
	"context"
	"fmt"
	"sort"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	"github.com/VeyronSakai/gh-runner-log/internal/domain/repository"
)

const (
	fetchLimitMultiplier = 10
	maxFetchLimit        = 1000
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

	// Fetch job history
	// Fetch more jobs than requested (fetchLimitMultiplier) so filtering by runner ID still returns enough rows.
	// Cap the upstream requests at maxFetchLimit to avoid excessive API calls.
	fetchLimit := limit * fetchLimitMultiplier
	if fetchLimit > maxFetchLimit {
		fetchLimit = maxFetchLimit
	}

	allJobs, err := r.jobRepo.FetchJobHistory(ctx, fetchLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job history: %w", err)
	}

	// Filter jobs by runner ID
	runnerJobs := make([]*entity.Job, 0)
	for _, job := range allJobs {
		if job.IsAssignedToRunner(runner.ID) {
			runnerJobs = append(runnerJobs, job)
			if len(runnerJobs) >= limit {
				break
			}
		}
	}

	// Sort jobs by start time (most recent first)
	sort.Slice(runnerJobs, func(i, j int) bool {
		if runnerJobs[i].StartedAt == nil {
			return false
		}
		if runnerJobs[j].StartedAt == nil {
			return true
		}
		return runnerJobs[i].StartedAt.After(*runnerJobs[j].StartedAt)
	})

	return &RunnerJobHistory{
		Runner: runner,
		Jobs:   runnerJobs,
	}, nil
}

// ...existing code...
