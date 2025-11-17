package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

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
		// Explicitly ignore the error from Fprintln as we're already exiting
		// and there's nothing meaningful we can do if stderr write fails
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&org, "org", "", "Fetch runner logs for an organization")
	rootCmd.Flags().StringVar(&repo, "repo", "", "Fetch runner logs for a specific repository (owner/repo)")
	rootCmd.Flags().IntVarP(&maxCount, "max-count", "n", 5, "Maximum number of history to display")
	rootCmd.Flags().StringVar(&debugFile, "debug", "", "Path to debug JSON file (bypasses GitHub API)")
}

func runCommand(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	runnerName := args[0]

	owner, repoName, orgName, err := determineScope(debugFile != "", org, repo)
	if err != nil {
		return err
	}

	jobRepo, runnerRepo, err := resolveRepositories(debugFile, owner, repoName, orgName)
	if err != nil {
		return err
	}

	// Create use case
	runnerLogger := usecase.NewRunnerLogger(jobRepo, runnerRepo)

	// Create and run controller
	controller := presentation.NewController(runnerLogger)
	return controller.Run(ctx, runnerName, maxCount)
}

func resolveRepositories(debugPath, owner, repo, org string) (repository.JobRepository, repository.RunnerRepository, error) {
	if debugPath != "" {
		jobRepo, runnerRepo, err := debuginfra.LoadRepositories(debugPath, owner, repo, org)
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

	jobRepo, err := github.NewJobRepository(basePath)
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
