package testhelpers

import (
	"context"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	repository "github.com/VeyronSakai/gh-runner-log/internal/domain/repository"
)

// StubRunnerRepository implements RunnerRepository for tests.
type StubRunnerRepository struct {
	Runner *entity.Runner
	Err    error
}

var _ repository.RunnerRepository = (*StubRunnerRepository)(nil)

func (s *StubRunnerRepository) FetchRunners(context.Context, string, string, string) ([]*entity.Runner, error) {
	if s.Runner == nil {
		return []*entity.Runner{}, nil
	}
	return []*entity.Runner{s.Runner}, nil
}

func (s *StubRunnerRepository) FetchRunnerByName(context.Context, string, string, string, string) (*entity.Runner, error) {
	return s.Runner, s.Err
}

// StubJobRepository implements JobRepository for tests.
type StubJobRepository struct {
	Jobs []*entity.Job
	Err  error
}

var _ repository.JobRepository = (*StubJobRepository)(nil)

func (s *StubJobRepository) FetchJobHistory(context.Context, string, string, string, int) ([]*entity.Job, error) {
	return s.Jobs, s.Err
}

// FailingRunnerRepository always returns the configured error.
type FailingRunnerRepository struct {
	Err error
}

var _ repository.RunnerRepository = (*FailingRunnerRepository)(nil)

func (f *FailingRunnerRepository) FetchRunners(context.Context, string, string, string) ([]*entity.Runner, error) {
	return nil, f.Err
}

func (f *FailingRunnerRepository) FetchRunnerByName(context.Context, string, string, string, string) (*entity.Runner, error) {
	return nil, f.Err
}
