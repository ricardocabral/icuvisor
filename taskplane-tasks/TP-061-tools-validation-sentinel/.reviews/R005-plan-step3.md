# R005 — Plan review: Step 3

**Verdict: APPROVE**

The Step 3 plan now identifies the right boundary and is specific enough to implement safely.

## What looks good

- Correctly scopes the mapper change to `internal/mcp/server.go`:
  - `publicToolErrorMessage(err)` for LLM-facing sanitization.
  - `safeRegistrar.AddTool`'s handler closure for logging because it has logger/tool context.
- Preserves existing `tools.PublicErrorMessage(err)` precedence, which avoids regressing current `NewUserError(...)` public messages and wire shape.
- Adds a bare `errors.Is(err, tools.ErrInvalidInput)` fallback with a short constant rather than exposing `err.Error()`.
- Separates operator logging policy from LLM-facing mapping:
  - invalid input at `Warn`;
  - unrelated handler failures remain `Error`;
  - no raw arguments, API keys, session data, or athlete IDs in logs.
- Includes targeted tests for public-message precedence, fallback redaction, generic fallback, and logging level behavior.

## Implementation reminders

- When extracting/logging the internal cause, use the Go pattern `var userErr *tools.UserError; errors.As(err, &userErr)` and guard `userErr.Unwrap() != nil` before adding a `cause` attribute.
- Keep the invalid-input fallback message as a package constant so mapper tests can assert it without duplicating a string in multiple places.
- Ensure the logging helper/branch does not accept or log request arguments; tests should make that hard to regress.

No blocking revisions required.
