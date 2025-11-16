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
	basePath   string
}

// NewJobRepository creates a new instance of JobRepositoryImpl
func NewJobRepository(basePath string) (domainrepo.JobRepository, error) {
	restClient, err := api.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %w\nPlease run 'gh auth login' to authenticate with GitHub", err)
	}

	return &JobRepositoryImpl{
		restClient: restClient,
		basePath:   basePath,
	}, nil
}

// FetchJobHistory retrieves job history for a repository or organization
// If runnerID is provided (> 0), only jobs assigned to that runner are returned
func (j *JobRepositoryImpl) FetchJobHistory(ctx context.Context, runnerID int64, limit int) ([]*entity.Job, error) {
	var allJobs []*entity.Job
	var skippedRuns int
	var lastJobErr error

	path := j.getWorkflowRunsPath()
	const perPage = 30
	page := 1

	// Fetch workflow runs page by page until we have enough jobs
	for {
		runs, err := j.fetchWorkflowRuns(path, perPage, page)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch workflow runs (page %d): %w", page, err)
		}

		// If no more runs, stop
		if len(runs.WorkflowRuns) == 0 {
			break
		}

		// Fetch jobs for each run in parallel
		type result struct {
			jobs []*entity.Job
			err  error
		}

		results := make(chan result, len(runs.WorkflowRuns))

		for _, run := range runs.WorkflowRuns {
			go func(r workflowRun) {
				jobs, err := j.getJobsForRun(r)
				results <- result{jobs: jobs, err: err}
			}(run)
		}

		// Collect results and filter by runner ID if specified
		for i := 0; i < len(runs.WorkflowRuns); i++ {
			res := <-results
			if res.err != nil {
				skippedRuns++
				lastJobErr = res.err
				continue
			}

			// Filter by runner ID if specified
			if runnerID > 0 {
				for _, job := range res.jobs {
					if job.IsAssignedToRunner(runnerID) {
						allJobs = append(allJobs, job)
					}
				}
			} else {
				allJobs = append(allJobs, res.jobs...)
			}
		}

		// If we have enough jobs, stop fetching
		if len(allJobs) >= limit {
			break
		}

		// If we got less than requested, we've reached the end
		if len(runs.WorkflowRuns) < perPage {
			break
		}

		page++
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
func (j *JobRepositoryImpl) getWorkflowRunsPath() string {
	return j.basePath + "/runs"
}

// fetchWorkflowRuns fetches workflow runs from GitHub API with pagination
func (j *JobRepositoryImpl) fetchWorkflowRuns(path string, perPage, page int) (*workflowRunsResponse, error) {
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
func (j *JobRepositoryImpl) getJobsForRun(run workflowRun) ([]*entity.Job, error) {
	// Extract owner and repo from the run's repository information
	if run.Repository.FullName == "" {
		return nil, fmt.Errorf("workflow run %d has no repository information", run.ID)
	}

	parts := strings.Split(run.Repository.FullName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository full name format: %s", run.Repository.FullName)
	}

	runOwner := parts[0]
	runRepo := parts[1]

	path := fmt.Sprintf("%s/runs/%d/jobs", getRepoActionsBasePath(runOwner, runRepo), run.ID)
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
