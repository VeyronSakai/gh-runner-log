package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/VeyronSakai/gh-runner-log/internal/infrastructure/github"
	"github.com/VeyronSakai/gh-runner-log/internal/presentation"
	"github.com/VeyronSakai/gh-runner-log/internal/usecase"
	ghrepo "github.com/cli/go-gh/v2/pkg/repository"
	"github.com/spf13/cobra"
)

var (
	runnerName string
	org        string
	repo       string
	limit      int
	format     string
)

var rootCmd = &cobra.Command{
	Use:   "gh-runner-log",
	Short: "View job execution history for GitHub Actions self-hosted runners",
	Long: `GitHub Actions Runner Log is a CLI tool that displays the job execution 
history for a specific self-hosted runner. It shows completed and in-progress 
jobs with details like workflow name, status, duration, and more.`,
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
	rootCmd.Flags().StringVarP(&runnerName, "runner", "r", "", "Name of the self-hosted runner (required)")
	rootCmd.Flags().StringVar(&org, "org", "", "Fetch runner logs for an organization")
	rootCmd.Flags().StringVar(&repo, "repo", "", "Fetch runner logs for a specific repository (owner/repo)")
	rootCmd.Flags().IntVarP(&limit, "limit", "l", 10, "Maximum number of jobs to display")
	rootCmd.Flags().StringVarP(&format, "format", "f", "table", "Output format: table or json")

	// Mark runner flag as required
	if err := rootCmd.MarkFlagRequired("runner"); err != nil {
		// This should never fail for a valid flag name
		panic(err)
	}
}

func runCommand(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	// Validate format
	var outputFormat presentation.OutputFormat
	switch strings.ToLower(format) {
	case "table":
		outputFormat = presentation.FormatTable
	case "json":
		outputFormat = presentation.FormatJSON
	default:
		return fmt.Errorf("invalid format: %s (supported: table, json)", format)
	}

	// Create infrastructure layer (GitHub client)
	runnerRepo, err := github.NewRunnerRepository()
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	jobRepo, err := github.NewJobRepository()
	if err != nil {
		return fmt.Errorf("failed to create GitHub job client: %w", err)
	}

	// Determine owner, repo, and org
	var owner, repoName, orgName string

	if org != "" {
		orgName = org
	} else if repo != "" {
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format. Use owner/repo")
		}
		owner = parts[0]
		repoName = parts[1]
	} else {
		// Try to get current repository
		currentRepo, err := ghrepo.Current()
		if err != nil {
			return fmt.Errorf("not in a git repository and no --repo or --org flag specified")
		}
		owner = currentRepo.Owner
		repoName = currentRepo.Name
	}

	// Create use case
	runnerLog := usecase.NewRunnerLog(jobRepo, runnerRepo)

	// Fetch runner job history
	history, err := runnerLog.FetchRunnerJobHistory(ctx, owner, repoName, orgName, runnerName, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch runner job history: %w", err)
	}

	// Format and display output
	formatter := presentation.NewFormatter(outputFormat)
	output, err := formatter.Format(history)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Println(output)
	return nil
}
