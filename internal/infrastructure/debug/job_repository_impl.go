package debug

import (
	"context"
	"strings"
	"time"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	domainrepo "github.com/VeyronSakai/gh-runner-log/internal/domain/repository"
)

var _ domainrepo.JobRepository = (*JobRepositoryImpl)(nil)

// JobRepositoryImpl serves job data from the loaded dataset.
type JobRepositoryImpl struct {
	ds           *dataset
	scope        string
	createdAfter time.Time
}

func NewJobRepository(ds *dataset, scope string, createdAfter time.Time) domainrepo.JobRepository {
	return &JobRepositoryImpl{
		ds:           ds,
		scope:        scope,
		createdAfter: createdAfter,
	}
}

func (j *JobRepositoryImpl) FetchJobHistory(_ context.Context, runnerID int64) ([]*entity.Job, error) {
	filtered := make([]*entity.Job, 0, len(j.ds.jobs))
	for _, job := range j.ds.jobs {
		if !j.matchScope(job.Repository) {
			continue
		}

		// Filter by runner ID if specified
		if runnerID > 0 && !job.IsAssignedToRunner(runnerID) {
			continue
		}

		// Filter by created time if specified
		// Note: In production, GitHub API filters by workflow run created_at time.
		// In debug mode, we use job started_at as a proxy since debug data doesn't
		// include workflow run metadata. This is acceptable for testing purposes.
		if !j.createdAfter.IsZero() {
			// Skip jobs without start time when time filter is active
			if job.StartedAt == nil || job.StartedAt.Before(j.createdAfter) {
				continue
			}
		}

		filtered = append(filtered, job)
	}

	return filtered, nil
}

// matchScope verifies that the repository string should be included for the given scope filter.
func (j *JobRepositoryImpl) matchScope(repository string) bool {
	if j.scope == "" {
		return true
	}
	// If scope contains "/", it's owner/repo format, otherwise it's org format
	if strings.Contains(j.scope, "/") {
		return repository == j.scope
	}
	// org format: check if repository starts with org/
	return strings.HasPrefix(repository, j.scope+"/")
}
