package repository

import (
	"context"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
)

// JobRepository defines the interface for accessing job data
type JobRepository interface {
	// FetchJobHistory retrieves job history for a repository or organization
	// Returns all jobs that have been completed or are in progress
	FetchJobHistory(ctx context.Context, owner, repo, org string, limit int) ([]*entity.Job, error)
}
