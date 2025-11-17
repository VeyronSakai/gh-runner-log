package testhelpers

import (
	"context"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
)

// StubJobRepository implements JobRepository for tests.
type StubJobRepository struct {
	Jobs []*entity.Job
	Err  error
}

func (s *StubJobRepository) FetchJobHistory(_ context.Context, runnerID int64, limit int) ([]*entity.Job, error) {
	if s.Err != nil {
		return nil, s.Err
	}

	// Filter by runner ID if specified
	filtered := make([]*entity.Job, 0)
	for _, job := range s.Jobs {
		if runnerID > 0 && !job.IsAssignedToRunner(runnerID) {
			continue
		}
		filtered = append(filtered, job)
		if len(filtered) >= limit {
			break
		}
	}

	return filtered, nil
}
