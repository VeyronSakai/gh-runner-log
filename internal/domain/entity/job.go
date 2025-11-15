package entity

import "time"

// Job represents a GitHub Actions workflow job
type Job struct {
	ID           int64
	RunID        int64
	Name         string
	Status       string
	Conclusion   string
	RunnerID     *int64
	RunnerName   *string
	StartedAt    *time.Time
	CompletedAt  *time.Time
	WorkflowName string
	Repository   string
	HtmlUrl      string
}

// IsCompleted returns true if the job has finished execution
func (j *Job) IsCompleted() bool {
	return j.Status == "completed"
}

// IsAssignedToRunner returns true if the job is assigned to a specific runner
func (j *Job) IsAssignedToRunner(runnerID int64) bool {
	return j.RunnerID != nil && *j.RunnerID == runnerID
}

// GetExecutionDuration returns the duration from start to completion
func (j *Job) GetExecutionDuration() time.Duration {
	if j.StartedAt == nil || j.CompletedAt == nil {
		return 0
	}
	return j.CompletedAt.Sub(*j.StartedAt)
}
