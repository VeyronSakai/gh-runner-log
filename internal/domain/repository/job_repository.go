package repository

import (
	"context"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
)

// JobRepository defines the interface for accessing job data
type JobRepository interface {
	// FetchJobHistory retrieves job history for a repository or organization
	// If runnerID is provided (> 0), only jobs assigned to that runner are returned
	FetchJobHistory(ctx context.Context, runnerID int64) ([]*entity.Job, error)
}
