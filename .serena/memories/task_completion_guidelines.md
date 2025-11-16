Before delivering changes:
1. `go test ./...` to ensure unit tests pass (add `-v` or `-cover` as needed).
2. For CLI features, optionally `go build -o gh-runner-log` and run sample commands (per README) to smoke test, especially new flags.
3. Keep README/docs updated with new flags/features; follow Clean Architecture layering when adding infra/usecase/presentation logic.
4. If distributing via gh extension, optionally run `gh extension install .` then `gh runner-log ...` to validate integration.
