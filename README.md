# gh-runner-log

A GitHub CLI extension that displays the job execution history for GitHub Actions self-hosted runners.

## Features

- 📜 View job execution history for specific self-hosted runners
- 📊 Display job details including workflow name, status, conclusion, and duration
- 🏢 Support for both repository and organization level runners
- ⌨️ Interactive UI with keyboard navigation
- 🌐 Open job run page in browser with Enter key
- 🔍 Configurable job limit

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

### Limit number of jobs displayed
```bash
gh runner-log my-runner-name --max-count 20
```

### Debug mode with sample JSON data
```bash
gh runner-log my-runner-name --debug ./test/sample_runner_jobs.json
```

## Command Line Flags

- `<runner-name>` - Name of the self-hosted runner (required, positional argument)
- `--repo` - Fetch runner logs for a specific repository (format: owner/repo)
- `--org` - Fetch runner logs for an organization
- `-n, --max-count` - Maximum number of jobs to display (default: 5)
- `--debug` - Load runner/job data from a local JSON file to simulate GitHub API responses

## Interactive UI

The tool displays an interactive list of jobs. Use the following keys:

- `↑/↓` or `j/k` - Navigate through jobs
- `Enter` - Open the selected job's run page in your browser
- `q` or `Ctrl+C` - Quit

## Example Output

```
Runner: my-runner (ID: 123456)
Status: online | OS: linux | Labels: self-hosted, linux, x64

Job History (5 jobs)
→ 987654321 | CI Pipeline          | completed | success | 2025-11-15 10:30:00 EST | 5m 23s
  987654320 | Build and Test       | completed | success | 2025-11-15 09:15:00 EST | 12m 45s
  987654319 | Deploy Production    | completed | success | 2025-11-15 08:00:00 EST | 3m 12s
  987654318 | Unit Tests           | completed | failure | 2025-11-15 07:45:00 EST | 2m 8s
  987654317 | Linting              | completed | success | 2025-11-15 07:30:00 EST | 1m 5s

↑/↓: Navigate • Enter: Open in browser • q: Quit
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

## Architecture

This project follows Clean Architecture principles:

- **Domain Layer**: Core business entities and repository interfaces
  - `entity`: Job and Runner entities
  - `repository`: Interfaces for data access
  
- **Infrastructure Layer**: External service implementations
  - `github`: GitHub API client implementations
  
- **Use Case Layer**: Application business logic
  - `runner_log`: Fetches and filters job history by runner
  
- **Presentation Layer**: Interactive UI
  - `interactive_ui`: Terminal UI for browsing and selecting jobs
  
- **Command Layer**: CLI interface
  - `cmd`: Cobra-based command definitions

## License

MIT