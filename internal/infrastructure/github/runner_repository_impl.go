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
	owner      string
	repo       string
	org        string
}

// NewRunnerRepository creates a new instance of RunnerRepositoryImpl
func NewRunnerRepository(owner, repo, org string) (domainrepo.RunnerRepository, error) {
	restClient, err := api.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %w\nPlease run 'gh auth login' to authenticate with GitHub", err)
	}

	return &RunnerRepositoryImpl{
		restClient: restClient,
		owner:      owner,
		repo:       repo,
		org:        org,
	}, nil
}

// FetchRunners retrieves all runners for a repository or organization
func (r *RunnerRepositoryImpl) FetchRunners(ctx context.Context) ([]*entity.Runner, error) {
	path := r.getRunnersPath(r.owner, r.repo, r.org)

	response, err := r.restClient.Request(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch runners: %w", err)
	}
	defer response.Body.Close()

	var runnersResp runnersResponse
	if err := json.NewDecoder(response.Body).Decode(&runnersResp); err != nil {
		return nil, fmt.Errorf("failed to decode runners response: %w", err)
	}

	runners := make([]*entity.Runner, 0, len(runnersResp.Runners))
	for _, r := range runnersResp.Runners {
		labels := make([]string, 0, len(r.Labels))
		for _, l := range r.Labels {
			labels = append(labels, l.Name)
		}

		runners = append(runners, &entity.Runner{
			ID:     r.ID,
			Name:   r.Name,
			OS:     r.OS,
			Status: r.Status,
			Labels: labels,
		})
	}

	return runners, nil
}

// FetchRunnerByName retrieves a specific runner by name
func (r *RunnerRepositoryImpl) FetchRunnerByName(ctx context.Context, name string) (*entity.Runner, error) {
	runners, err := r.FetchRunners(ctx)
	if err != nil {
		return nil, err
	}

	for _, runner := range runners {
		if runner.Name == name {
			return runner, nil
		}
	}

	return nil, fmt.Errorf("runner '%s' not found", name)
}

// getRunnersPath constructs the API path for fetching runners
func (r *RunnerRepositoryImpl) getRunnersPath(owner, repo, org string) string {
	if org != "" {
		return fmt.Sprintf("orgs/%s/actions/runners?per_page=100", org)
	}
	return fmt.Sprintf("repos/%s/%s/actions/runners?per_page=100", owner, repo)
}
