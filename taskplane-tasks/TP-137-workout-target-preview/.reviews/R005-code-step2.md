# Code Review — Step 2

Result: Approved.

No blocking findings. The implementation adds compact `workout_doc_summary.target_previews` through the shared event/workout row helpers, reuses the existing profile fetch, omits unsupported targets, and preserves terse/full raw-payload behavior in the covered paths.

Verification run:

- `go test ./internal/tools ./internal/workoutdoc`
- `go test ./...`
- `golangci-lint run ./...`
- `git diff --check 9ab6472..HEAD`
