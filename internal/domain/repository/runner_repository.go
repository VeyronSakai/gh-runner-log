package repository

import (
	"context"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
)

// RunnerRepository defines the interface for accessing runner data
type RunnerRepository interface {
	// FetchRunners retrieves all runners for a repository or organization
	FetchRunners(ctx context.Context, owner, repo, org string) ([]*entity.Runner, error)
	
	// FetchRunnerByName retrieves a specific runner by name
	FetchRunnerByName(ctx context.Context, owner, repo, org, name string) (*entity.Runner, error)
}
