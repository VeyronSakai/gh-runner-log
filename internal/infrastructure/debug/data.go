package debug

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	domainrepo "github.com/VeyronSakai/gh-runner-log/internal/domain/repository"
)

// LoadRepositories loads job and runner repositories from a debug file.
func LoadRepositories(path string) (domainrepo.JobRepository, domainrepo.RunnerRepository, error) {
	ds, err := loadDataset(path)
	if err != nil {
		return nil, nil, err
	}
	return NewJobRepository(ds), NewRunnerRepository(ds), nil
}

// dataFile mirrors the JSON schema used by the --debug flag.
type dataFile struct {
	Runners []runnerRecord `json:"runners"`
	Jobs    []jobRecord    `json:"jobs"`
}

type runnerRecord struct {
	ID     int64    `json:"id"`
	Name   string   `json:"name"`
	Labels []string `json:"labels"`
	OS     string   `json:"os"`
	Status string   `json:"status"`
}

type jobRecord struct {
	ID           int64      `json:"id"`
	RunID        int64      `json:"run_id"`
	Name         string     `json:"name"`
	Status       string     `json:"status"`
	Conclusion   string     `json:"conclusion"`
	RunnerID     *int64     `json:"runner_id"`
	RunnerName   *string    `json:"runner_name"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	WorkflowName string     `json:"workflow_name"`
	Repository   string     `json:"repository"`
	HtmlURL      string     `json:"html_url"`
}

// dataset keeps parsed entities ready for repositories.
type dataset struct {
	runners []*entity.Runner
	jobs    []*entity.Job
}

func loadDataset(path string) (*dataset, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read debug file: %w", err)
	}

	var raw dataFile
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse debug file: %w", err)
	}

	if len(raw.Runners) == 0 {
		return nil, fmt.Errorf("debug file does not contain any runners")
	}

	ds := &dataset{
		runners: make([]*entity.Runner, 0, len(raw.Runners)),
		jobs:    make([]*entity.Job, 0, len(raw.Jobs)),
	}

	for _, r := range raw.Runners {
		ds.runners = append(ds.runners, &entity.Runner{
			ID:     r.ID,
			Name:   r.Name,
			Labels: append([]string(nil), r.Labels...),
			OS:     r.OS,
			Status: r.Status,
		})
	}

	for _, j := range raw.Jobs {
		job := &entity.Job{
			ID:           j.ID,
			RunID:        j.RunID,
			Name:         j.Name,
			Status:       j.Status,
			Conclusion:   j.Conclusion,
			RunnerID:     j.RunnerID,
			RunnerName:   j.RunnerName,
			StartedAt:    j.StartedAt,
			CompletedAt:  j.CompletedAt,
			WorkflowName: j.WorkflowName,
			Repository:   j.Repository,
			HtmlUrl:      j.HtmlURL,
		}
		ds.jobs = append(ds.jobs, job)
	}

	return ds, nil
}
