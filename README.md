# gh-runner-log

A GitHub CLI extension that displays the job run history of any GitHub Actions self-hosted runner.

## Features

- 📜 View job execution history for specific self-hosted runners
- 📊 Display job details including workflow name, status, conclusion, and duration
- ⌨️ Interactive UI with keyboard navigation
- 🌐 Open job run page in browser with Enter key

<img width="831" height="268" alt="スクリーンショット 2025-11-18 1 00 23" src="https://github.com/user-attachments/assets/a0f20cb8-b4d4-497f-bf4b-b2298f021942" />


## Installation

```bash
gh extension install VeyronSakai/gh-runner-log
```

## Usage

### View runner job history for current repository
```bash
gh runner-log my-runner-name
```

### View runner job history for specific repository
```bash
gh runner-log my-runner-name --repo owner/repo
```

### View runner job history for organization
```bash
gh runner-log my-runner-name --org organization-name
```

### Specify number of jobs displayed
```bash
gh runner-log my-runner-name --max-count 10
```

### Filter jobs by time range
```bash
# Show jobs from the last 24 hours (default)
gh runner-log my-runner-name

# Show jobs from the last 7 days
gh runner-log my-runner-name --since 7d

# Show jobs from the last 2 weeks
gh runner-log my-runner-name --since 2w

# Show jobs since a specific date
gh runner-log my-runner-name --since 2025-11-01
```

## Command Line Flags

- `<runner-name>` - Name of the self-hosted runner (required, positional argument)
- `--repo` - Fetch runner logs for a specific repository (format: owner/repo)
- `--org` - Fetch runner logs for an organization
- `-n, --max-count` - Maximum number of jobs to display (default: 5)
- `--since` - Show jobs created since this time (default: 24h)
  - Duration format: `24h`, `2d`, `1w` (hours, days, weeks)
  - Date format: `2025-11-17` (YYYY-MM-DD)
  - RFC3339 format: `2025-11-17T10:00:00Z`
- `--debug` - Load runner/job data from a local JSON file to simulate GitHub API responses

## Interactive UI

The tool displays an interactive list of jobs. Use the following keys:

- `↑/↓` or `j/k` - Navigate through jobs
- `Enter` - Open the selected job's run page in your browser
- `q` or `Ctrl+C` - Quit

## Example Output

```
Runner: my-runner
Status: online
OS: linux
Labels: self-hosted, linux, x64

┌───────────────┬──────────────────────┬───────────┬────────────┬──────────────────────────────┬──────────┐
│ Workflow      │ Job                  │ Status    │ Conclusion │ Started At                   │ Duration │
├───────────────┼──────────────────────┼───────────┼────────────┼──────────────────────────────┼──────────┤
│ CI Pipeline   │ Build                │ completed │ success    │ 2025-11-15 10:30:00 EST      │ 5m 23s   │
│ Build and Test│ Test                 │ completed │ success    │ 2025-11-15 09:15:00 EST      │ 12m 45s  │
│ Deploy Prod   │ Deploy               │ completed │ success    │ 2025-11-15 08:00:00 EST      │ 3m 12s   │
│ Unit Tests    │ Unit Test            │ completed │ failure    │ 2025-11-15 07:45:00 EST      │ 2m 8s    │
│ Linting       │ Lint                 │ completed │ success    │ 2025-11-15 07:30:00 EST      │ 1m 5s    │
└───────────────┴──────────────────────┴───────────┴────────────┴──────────────────────────────┴──────────┘

↑/↓ or j/k: Navigate • Enter: Open in browser • q or Ctrl+C: Quit
```

## Development

### Prerequisites

- Go 1.25 or higher
- GitHub CLI (`gh`) installed and authenticated

### Building from source

```bash
git clone https://github.com/VeyronSakai/gh-runner-log.git
cd gh-runner-log
go build -o gh-runner-log
```

### Testing Locally

#### 1. Build and run directly
```bash
# Build the binary
go build -o gh-runner-log

# Run with help flag to see options
./gh-runner-log --help

# View runner job history
./gh-runner-log my-runner-name
```

#### 2. Install as gh extension from local directory
```bash
# Install from current directory
gh extension install .

# Run as gh extension
gh runner-log my-runner-name

# Uninstall when done testing
gh extension remove runner-log
```

### Running tests

```bash
# Run all tests
go test ./...
```

### Debug JSON format

Create a JSON file containing runners and jobs to validate CLI output without calling the GitHub API. For example, save the following as `debug.json`:

```json
{
  "runners": [
    {
      "id": 123,
      "name": "runner-a",
      "labels": ["self-hosted", "linux"],
      "os": "linux",
      "status": "online"
    }
  ],
  "jobs": [
    {
      "id": 98765,
      "run_id": 54321,
      "run_attempt": 1,
      "name": "Build",
      "status": "completed",
      "conclusion": "success",
      "runner_id": 123,
      "runner_name": "runner-a",
      "started_at": "2025-11-15T10:00:00Z",
      "completed_at": "2025-11-15T10:05:00Z",
      "workflow_name": "CI",
      "repository": "owner/repo",
      "html_url": "https://github.com/owner/repo/actions/runs/54321/job/98765"
    }
  ]
}
```

Run the CLI against this file with:

```bash
./gh-runner-log runner-a --debug ./debug.json
```
