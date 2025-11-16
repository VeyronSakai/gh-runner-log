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

func (s *StubJobRepository) FetchJobHistory(context.Context, int) ([]*entity.Job, error) {
	return s.Jobs, s.Err
}
