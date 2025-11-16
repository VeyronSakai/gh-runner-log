package debug

import (
	"context"
	"strings"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	domainrepo "github.com/VeyronSakai/gh-runner-log/internal/domain/repository"
)

var _ domainrepo.JobRepository = (*JobRepositoryImpl)(nil)

// JobRepositoryImpl serves job data from the loaded dataset.
type JobRepositoryImpl struct {
	ds    *dataset
	scope string
}

func NewJobRepository(ds *dataset, scope string) domainrepo.JobRepository {
	return &JobRepositoryImpl{
		ds:    ds,
		scope: scope,
	}
}

func (j *JobRepositoryImpl) FetchJobHistory(_ context.Context, runnerID int64, limit int) ([]*entity.Job, error) {
	if limit <= 0 {
		return []*entity.Job{}, nil
	}

	filtered := make([]*entity.Job, 0, len(j.ds.jobs))
	for _, job := range j.ds.jobs {
		if !j.matchScope(job.Repository) {
			continue
		}

		// Filter by runner ID if specified
		if runnerID > 0 && !job.IsAssignedToRunner(runnerID) {
			continue
		}

		filtered = append(filtered, job)
		if len(filtered) >= limit {
			break
		}
	}

	if len(filtered) > limit {
		filtered = filtered[:limit]
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
