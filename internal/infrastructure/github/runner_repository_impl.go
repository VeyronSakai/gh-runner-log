package github

import (
	"context"
	"fmt"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	domainrepo "github.com/VeyronSakai/gh-runner-log/internal/domain/repository"
	"github.com/cli/go-gh/v2/pkg/api"
)

// RunnerRepositoryImpl implements the RunnerRepository interface using GitHub API
type RunnerRepositoryImpl struct {
	restClient *api.RESTClient
	basePath   string
}

// NewRunnerRepository creates a new instance of RunnerRepositoryImpl
func NewRunnerRepository(basePath string) (domainrepo.RunnerRepository, error) {
	restClient, err := api.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %w\nPlease run 'gh auth login' to authenticate with GitHub", err)
	}

	return &RunnerRepositoryImpl{
		restClient: restClient,
		basePath:   basePath,
	}, nil
}

// FetchRunnerByName retrieves a specific runner by name
func (r *RunnerRepositoryImpl) FetchRunnerByName(ctx context.Context, name string) (*entity.Runner, error) {
	path := r.getRunnersPath()

	var runnersResp runnersResponse
	if err := r.restClient.Get(path, &runnersResp); err != nil {
		return nil, fmt.Errorf("failed to fetch runners: %w", err)
	}

	for _, runner := range runnersResp.Runners {
		if runner.Name == name {
			labels := make([]string, 0, len(runner.Labels))
			for _, l := range runner.Labels {
				labels = append(labels, l.Name)
			}

			return &entity.Runner{
				ID:     runner.ID,
				Name:   runner.Name,
				OS:     runner.OS,
				Status: runner.Status,
				Labels: labels,
			}, nil
		}
	}

	return nil, fmt.Errorf("runner '%s' not found in the repository", name)
}

// getRunnersPath constructs the API path for fetching runners
func (r *RunnerRepositoryImpl) getRunnersPath() string {
	return r.basePath + "/runners?per_page=100"
}
