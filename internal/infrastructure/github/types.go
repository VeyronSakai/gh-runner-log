package github

import "time"

// workflowRunsResponse represents the response from GitHub API for workflow runs
type workflowRunsResponse struct {
	TotalCount   int           `json:"total_count"`
	WorkflowRuns []workflowRun `json:"workflow_runs"`
}

// workflowRun represents a single workflow run
type workflowRun struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Conclusion   string    `json:"conclusion"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Repository   repoInfo  `json:"repository"`
	HeadBranch   string    `json:"head_branch"`
	HeadSha      string    `json:"head_sha"`
	RunNumber    int       `json:"run_number"`
	RunAttempt   int       `json:"run_attempt"`
	Event        string    `json:"event"`
	DisplayTitle string    `json:"display_title"`
	HtmlUrl      string    `json:"html_url"`
}

// repoInfo represents repository information in a workflow run
type repoInfo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

// jobsResponse represents the response from GitHub API for jobs
type jobsResponse struct {
	TotalCount int   `json:"total_count"`
	Jobs       []job `json:"jobs"`
}

// job represents a single job in a workflow run
type job struct {
	ID          int64      `json:"id"`
	RunID       int64      `json:"run_id"`
	RunAttempt  int        `json:"run_attempt"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Conclusion  string     `json:"conclusion"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	RunnerID    *int64     `json:"runner_id"`
	RunnerName  *string    `json:"runner_name"`
	HtmlUrl     string     `json:"html_url"`
}

// runnersResponse represents the response from GitHub API for runners
type runnersResponse struct {
	TotalCount int      `json:"total_count"`
	Runners    []runner `json:"runners"`
}

// runner represents a single runner
type runner struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	OS     string  `json:"os"`
	Status string  `json:"status"`
	Labels []label `json:"labels"`
}

// label represents a runner label
type label struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
