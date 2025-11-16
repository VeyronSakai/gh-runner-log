package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	domainrepo "github.com/VeyronSakai/gh-runner-log/internal/domain/repository"
	"github.com/cli/go-gh/v2/pkg/api"
)

// JobRepositoryImpl implements the JobRepository interface using GitHub API
type JobRepositoryImpl struct {
	restClient *api.RESTClient
}

// NewJobRepository creates a new instance of JobRepositoryImpl
func NewJobRepository() (domainrepo.JobRepository, error) {
	restClient, err := api.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %w\nPlease run 'gh auth login' to authenticate with GitHub", err)
	}

	return &JobRepositoryImpl{
		restClient: restClient,
	}, nil
}

// FetchJobHistory retrieves job history for a repository or organization
func (j *JobRepositoryImpl) FetchJobHistory(ctx context.Context, owner, repo, org string, limit int) ([]*entity.Job, error) {
	// Fetch workflow runs without status filter to get all statuses
	// Use default per_page (30) from GitHub API
	path := j.getWorkflowRunsPath(owner, repo, org)
	runs, err := j.fetchWorkflowRuns(path, 30)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workflow runs: %w", err)
	}

	// Fetch jobs for each run in parallel
	type result struct {
		jobs []*entity.Job
		err  error
	}

	results := make(chan result, len(runs.WorkflowRuns))

	for _, run := range runs.WorkflowRuns {
		go func(r workflowRun) {
			jobs, err := j.getJobsForRun(r, org, owner, repo)
			results <- result{jobs: jobs, err: err}
		}(run)
	}

	// Collect results
	var allJobs []*entity.Job
	var skippedRuns int
	var lastJobErr error

	for i := 0; i < len(runs.WorkflowRuns); i++ {
		res := <-results
		if res.err != nil {
			skippedRuns++
			lastJobErr = res.err
			continue
		}
		allJobs = append(allJobs, res.jobs...)
	}

	if skippedRuns > 0 && lastJobErr != nil && len(allJobs) == 0 {
		return nil, fmt.Errorf("failed to fetch jobs for %d workflow run(s): %w", skippedRuns, lastJobErr)
	}

	// Apply limit after collecting all jobs
	if len(allJobs) > limit {
		allJobs = allJobs[:limit]
	}

	return allJobs, nil
}

// getWorkflowRunsPath constructs the API path for fetching workflow runs
func (j *JobRepositoryImpl) getWorkflowRunsPath(owner, repo, org string) string {
	if org != "" {
		return fmt.Sprintf("orgs/%s/actions/runs", org)
	}
	return fmt.Sprintf("repos/%s/%s/actions/runs", owner, repo)
}

// fetchWorkflowRuns fetches workflow runs from GitHub API (page 1 only, for backward compatibility)
func (j *JobRepositoryImpl) fetchWorkflowRuns(path string, perPage int) (*workflowRunsResponse, error) {
	return j.fetchWorkflowRunsWithPagination(path, perPage, 1)
}

// fetchWorkflowRunsWithPagination fetches workflow runs from GitHub API with pagination support
func (j *JobRepositoryImpl) fetchWorkflowRunsWithPagination(path string, perPage, page int) (*workflowRunsResponse, error) {
	// Determine the separator for query parameters
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}

	currentPath := fmt.Sprintf("%s%sper_page=%d&page=%d", path, separator, perPage, page)
	response, err := j.restClient.Request(http.MethodGet, currentPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request workflow runs: %w", err)
	}
	defer response.Body.Close()

	var runs workflowRunsResponse
	if err := json.NewDecoder(response.Body).Decode(&runs); err != nil {
		return nil, fmt.Errorf("failed to decode workflow runs response: %w", err)
	}

	return &runs, nil
}

// getJobsForRun fetches all jobs for a specific workflow run
func (j *JobRepositoryImpl) getJobsForRun(run workflowRun, org, owner, repo string) ([]*entity.Job, error) {
	// Use the repository from the run if available, otherwise use provided values
	runOwner := owner
	runRepo := repo
	if run.Repository.FullName != "" {
		parts := strings.Split(run.Repository.FullName, "/")
		if len(parts) == 2 {
			runOwner = parts[0]
			runRepo = parts[1]
		}
	}

	path := fmt.Sprintf("repos/%s/%s/actions/runs/%d/jobs", runOwner, runRepo, run.ID)
	response, err := j.restClient.Request(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs for run %d: %w", run.ID, err)
	}
	defer response.Body.Close()

	var jobsResp jobsResponse
	if err := json.NewDecoder(response.Body).Decode(&jobsResp); err != nil {
		return nil, fmt.Errorf("failed to decode jobs response: %w", err)
	}

	jobs := make([]*entity.Job, 0, len(jobsResp.Jobs))
	for _, j := range jobsResp.Jobs {
		jobs = append(jobs, &entity.Job{
			ID:           j.ID,
			RunID:        j.RunID,
			Name:         j.Name,
			Status:       j.Status,
			Conclusion:   j.Conclusion,
			RunnerID:     j.RunnerID,
			RunnerName:   j.RunnerName,
			StartedAt:    j.StartedAt,
			CompletedAt:  j.CompletedAt,
			WorkflowName: run.Name,
			Repository:   run.Repository.FullName,
			HtmlUrl:      j.HtmlUrl,
		})
	}

	return jobs, nil
}
