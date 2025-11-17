package debug

import (
	"context"
	"fmt"
	"strings"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	domainrepo "github.com/VeyronSakai/gh-runner-log/internal/domain/repository"
)

var _ domainrepo.RunnerRepository = (*RunnerRepositoryImpl)(nil)

// RunnerRepositoryImpl serves runner data from the loaded dataset.
type RunnerRepositoryImpl struct {
	ds    *dataset
	scope string
}

func NewRunnerRepository(ds *dataset, scope string) domainrepo.RunnerRepository {
	return &RunnerRepositoryImpl{
		ds:    ds,
		scope: scope,
	}
}

// FetchRunnerByName retrieves a runner by its name.
func (r *RunnerRepositoryImpl) FetchRunnerByName(_ context.Context, name string) (*entity.Runner, error) {
	for _, runner := range r.ds.runners {
		if strings.EqualFold(runner.Name, name) {
			return runner, nil
		}
	}
	return nil, fmt.Errorf("runner '%s' not found in debug dataset", name)
}
