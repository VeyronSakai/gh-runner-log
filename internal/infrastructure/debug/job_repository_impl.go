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
	owner string
	repo  string
	org   string
}

func NewJobRepository(ds *dataset, owner, repo, org string) domainrepo.JobRepository {
	return &JobRepositoryImpl{
		ds:    ds,
		owner: owner,
		repo:  repo,
		org:   org,
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

// matchScope verifies that the repository string should be included for the given owner/repo/org filters.
func (j *JobRepositoryImpl) matchScope(repository string) bool {
	if j.org != "" {
		return strings.HasPrefix(repository, j.org+"/")
	}
	if j.owner == "" || j.repo == "" {
		return true
	}
	return repository == j.owner+"/"+j.repo
}
