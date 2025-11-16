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

func (s *StubRunnerRepository) FetchRunners(context.Context) ([]*entity.Runner, error) {
	if s.Runner == nil {
		return []*entity.Runner{}, nil
	}
	return []*entity.Runner{s.Runner}, nil
}

func (s *StubRunnerRepository) FetchRunnerByName(context.Context, string) (*entity.Runner, error) {
	return s.Runner, s.Err
}

// FailingRunnerRepository always returns the configured error.
type FailingRunnerRepository struct {
	Err error
}

var _ repository.RunnerRepository = (*FailingRunnerRepository)(nil)

func (f *FailingRunnerRepository) FetchRunners(context.Context) ([]*entity.Runner, error) {
	return nil, f.Err
}

func (f *FailingRunnerRepository) FetchRunnerByName(context.Context, string) (*entity.Runner, error) {
	return nil, f.Err
}
