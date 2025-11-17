package repository

import (
	"context"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
)

// RunnerRepository defines the interface for accessing runner data
type RunnerRepository interface {
	// FetchRunnerByName retrieves a specific runner by name
	FetchRunnerByName(ctx context.Context, name string) (*entity.Runner, error)
}
