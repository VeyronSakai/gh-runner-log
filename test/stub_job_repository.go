package testhelpers

import (
	"context"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	repository "github.com/VeyronSakai/gh-runner-log/internal/domain/repository"
)

// StubJobRepository implements JobRepository for tests.
type StubJobRepository struct {
	Jobs []*entity.Job
	Err  error
}

var _ repository.JobRepository = (*StubJobRepository)(nil)

func (s *StubJobRepository) FetchJobHistory(context.Context, int) ([]*entity.Job, error) {
	return s.Jobs, s.Err
}
