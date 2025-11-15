package usecase

import (
	"context"
	"fmt"
	"sort"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	"github.com/VeyronSakai/gh-runner-log/internal/domain/repository"
)

// RunnerLog is a use case for fetching job history for a specific runner
type RunnerLog struct {
	jobRepo    repository.JobRepository
	runnerRepo repository.RunnerRepository
}

// NewRunnerLog creates a new RunnerLog use case
func NewRunnerLog(jobRepo repository.JobRepository, runnerRepo repository.RunnerRepository) *RunnerLog {
	return &RunnerLog{
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
func (r *RunnerLog) FetchRunnerJobHistory(ctx context.Context, owner, repo, org, runnerName string, limit int) (*RunnerJobHistory, error) {
	// First, fetch the runner to get its ID
	runner, err := r.runnerRepo.FetchRunnerByName(ctx, owner, repo, org, runnerName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch runner: %w", err)
	}

	// Fetch job history
	// We fetch more jobs than the limit because we need to filter by runner
	fetchLimit := limit * 10 // Fetch more to ensure we get enough jobs for this runner
	if fetchLimit > 1000 {
		fetchLimit = 1000
	}
	
	allJobs, err := r.jobRepo.FetchJobHistory(ctx, owner, repo, org, fetchLimit)
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
