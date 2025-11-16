# gh-runner-log
- Purpose: GitHub CLI extension that shows job execution history for GitHub Actions self-hosted runners at repo/org scope, using Clean Architecture layers.
- Tech stack: Go 1.25+, Cobra for CLI, GitHub CLI REST client (`github.com/cli/go-gh/v2`).
- Structure: `cmd/` (Cobra root), `internal/domain` (entities & repo interfaces), `internal/infrastructure/github` (REST implementations), `internal/usecase` (runner log logic), `internal/presentation` (table/JSON formatter), `main.go` entry.
- Key dependencies: Cobra, go-gh REST client, standard lib.
- Entry point: `main.go` -> `cmd.Execute()`.
- Testing: `go test ./...` (coverage/verbose variants described in README).
- Build: `go build -o gh-runner-log` (per README) or install via `gh extension install .` for local testing.
- Clean Architecture layering emphasized: domain entities/interfaces → infrastructure → usecase → presentation → command.
