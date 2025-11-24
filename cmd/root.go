package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/repository"
	debuginfra "github.com/VeyronSakai/gh-runner-log/internal/infrastructure/debug"
	"github.com/VeyronSakai/gh-runner-log/internal/infrastructure/github"
	"github.com/VeyronSakai/gh-runner-log/internal/presentation"
	"github.com/VeyronSakai/gh-runner-log/internal/usecase"
	ghrepo "github.com/cli/go-gh/v2/pkg/repository"
	"github.com/spf13/cobra"
)

var (
	org       string
	repo      string
	maxCount  int
	debugFile string
	since     string
)

var rootCmd = &cobra.Command{
	Use:   "gh-runner-log <runner-name>",
	Short: "View job execution history for GitHub Actions self-hosted runners",
	Long: `GitHub Actions Runner Log is a CLI tool that displays the job execution 
history for a specific self-hosted runner. It shows completed and in-progress 
jobs with details like workflow name, status, duration, and more.`,
	Args: cobra.ExactArgs(1),
	RunE: runCommand,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&org, "org", "", "Fetch runner logs for an organization")
	rootCmd.Flags().StringVar(&repo, "repo", "", "Fetch runner logs for a specific repository (owner/repo)")
	rootCmd.Flags().IntVarP(&maxCount, "max-count", "n", 5, "Maximum number of jobs to display")
	rootCmd.Flags().StringVar(&debugFile, "debug", "", "Path to debug JSON file (bypasses GitHub API)")
	rootCmd.Flags().StringVar(&since, "since", "24h", "Show jobs created since this time (e.g., '24h', '2d', '1w', or RFC3339 format)")
}

func runCommand(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	runnerName := args[0]

	owner, repoName, orgName, err := determineScope(debugFile != "", org, repo)
	if err != nil {
		return err
	}

	// Parse since parameter
	createdAfter, err := parseSince(since)
	if err != nil {
		return fmt.Errorf("invalid --since value: %w", err)
	}

	jobRepo, runnerRepo, err := resolveRepositories(debugFile, owner, repoName, orgName, createdAfter)
	if err != nil {
		return err
	}

	// Create use case
	runnerLogger := usecase.NewRunnerLogger(jobRepo, runnerRepo)

	// Create and run controller
	controller := presentation.NewController(runnerLogger)
	return controller.Run(ctx, runnerName, maxCount)
}

func resolveRepositories(debugPath, owner, repo, org string, createdAfter time.Time) (repository.JobRepository, repository.RunnerRepository, error) {
	if debugPath != "" {
		jobRepo, runnerRepo, err := debuginfra.LoadRepositories(debugPath, owner, repo, org, createdAfter)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load debug data: %w", err)
		}
		return jobRepo, runnerRepo, nil
	}

	basePath := github.GetActionsBasePath(owner, repo, org)

	runnerRepo, err := github.NewRunnerRepository(basePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	jobRepo, err := github.NewJobRepository(basePath, createdAfter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GitHub job client: %w", err)
	}

	return jobRepo, runnerRepo, nil
}

func determineScope(debugEnabled bool, orgFlag, repoFlag string) (string, string, string, error) {
	var owner, repoName, orgName string

	if orgFlag != "" {
		orgName = orgFlag
		return owner, repoName, orgName, nil
	}

	if repoFlag != "" {
		parts := strings.Split(repoFlag, "/")
		if len(parts) != 2 {
			return "", "", "", fmt.Errorf("invalid repository format. Use owner/repo")
		}
		return parts[0], parts[1], "", nil
	}

	if debugEnabled {
		return "", "", "", nil
	}

	currentRepo, err := ghrepo.Current()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to detect current repository context. Please specify either --repo owner/repo or --org organization-name")
	}

	return currentRepo.Owner, currentRepo.Name, "", nil
}

// parseSince parses the --since flag value and returns the corresponding time
// Supports formats like "24h", "2d", "1w" or RFC3339 timestamps
func parseSince(since string) (time.Time, error) {
	if since == "" {
		// Default to 24 hours ago
		return time.Now().Add(-24 * time.Hour), nil
	}

	// Try parsing as duration (e.g., "24h", "2d", "1w")
	if duration, err := parseDuration(since); err == nil {
		return time.Now().Add(-duration), nil
	}

	// Try parsing as RFC3339 timestamp
	if t, err := time.Parse(time.RFC3339, since); err == nil {
		return t, nil
	}

	// Try parsing as date only (YYYY-MM-DD)
	if t, err := time.Parse("2006-01-02", since); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s (expected format: duration like '24h', '2d', '1w' or date like '2024-01-01')", since)
}

// parseDuration extends time.ParseDuration to support days (d) and weeks (w)
func parseDuration(s string) (time.Duration, error) {
	// Try standard Go duration first
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// Handle days (d) and weeks (w)
	var value int
	var unit string
	if n, err := fmt.Sscanf(s, "%d%s", &value, &unit); err == nil && n == 2 {
		switch unit {
		case "d":
			return time.Duration(value) * 24 * time.Hour, nil
		case "w":
			return time.Duration(value) * 7 * 24 * time.Hour, nil
		}
	}

	return 0, fmt.Errorf("invalid duration format: %s", s)
}
