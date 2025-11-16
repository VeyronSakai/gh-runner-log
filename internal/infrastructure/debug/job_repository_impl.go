package debug

import (
	"context"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	domainrepo "github.com/VeyronSakai/gh-runner-log/internal/domain/repository"
)

var _ domainrepo.JobRepository = (*JobRepositoryImpl)(nil)

// JobRepositoryImpl serves job data from the loaded dataset.
type JobRepositoryImpl struct {
	ds *dataset
}

func NewJobRepository(ds *dataset) domainrepo.JobRepository {
	return &JobRepositoryImpl{ds: ds}
}

func (j *JobRepositoryImpl) FetchJobHistory(_ context.Context, owner, repo, org string, limit int) ([]*entity.Job, error) {
	if limit <= 0 {
		return []*entity.Job{}, nil
	}

	filtered := make([]*entity.Job, 0, len(j.ds.jobs))
	for _, job := range j.ds.jobs {
		if matchScope(job.Repository, owner, repo, org) {
			filtered = append(filtered, job)
		}
		if len(filtered) >= limit {
			break
		}
	}

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}
