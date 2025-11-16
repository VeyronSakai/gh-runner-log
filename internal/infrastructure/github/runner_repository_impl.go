package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

	response, err := r.restClient.Request(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch runners: %w", err)
	}
	defer response.Body.Close()

	var runnersResp runnersResponse
	if err := json.NewDecoder(response.Body).Decode(&runnersResp); err != nil {
		return nil, fmt.Errorf("failed to decode runners response: %w", err)
	}

	for _, r := range runnersResp.Runners {
		if r.Name == name {
			labels := make([]string, 0, len(r.Labels))
			for _, l := range r.Labels {
				labels = append(labels, l.Name)
			}

			return &entity.Runner{
				ID:     r.ID,
				Name:   r.Name,
				OS:     r.OS,
				Status: r.Status,
				Labels: labels,
			}, nil
		}
	}

	return nil, fmt.Errorf("runner '%s' not found", name)
}

// getRunnersPath constructs the API path for fetching runners
func (r *RunnerRepositoryImpl) getRunnersPath() string {
	return r.basePath + "/runners?per_page=100"
}
