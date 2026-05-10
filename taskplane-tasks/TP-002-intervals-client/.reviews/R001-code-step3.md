# Code Review — Step 3: Implement retries and structured errors

**Verdict:** REVISE

## Findings

### 1. `make lint` fails because the newly added retry/error path remains unused

- **Severity:** Blocking
- **Files/lines:** `internal/intervals/client.go:95`, `internal/intervals/client.go:106`, `internal/intervals/client.go:169`, `internal/intervals/client.go:173`, `internal/intervals/client.go:180`, `internal/intervals/client.go:192`, `internal/intervals/client.go:206`, `internal/intervals/client.go:222`, `internal/intervals/errors.go:46`
- **Details:** `golangci-lint run ./...` reports all of the private client/retry helpers as unused. The project lint configuration enables the `unused` linter, so this change cannot pass CI in its current state.
- **Command output:**
  - `go test ./...` passes.
  - `make lint` fails with 9 `unused` issues for `newRequest`, `doJSON`, retry helpers, `addJitter`, `parseRetryAfter`, and `errorForStatus`.
- **Recommendation:** Either wire the client path into a real production caller in this step, or add focused same-package tests that exercise `doJSON` and the retry/error helpers, or defer these private helpers until the profile retrieval step introduces a caller. The important outcome is that the repository should remain lint-clean after the step.

## Notes

- The structured sentinel error design is generally on the right track: status errors are wrapped with `%w`, and `errors.Is` / `errors.As` should work for the implemented HTTP-status paths.
- Retry handling correctly limits retries to GET and retries only 429/5xx response statuses, while respecting context cancellation during backoff sleeps.
