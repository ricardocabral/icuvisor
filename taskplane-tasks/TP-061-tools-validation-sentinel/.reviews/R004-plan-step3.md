# R004 — Plan review: Step 3

**Verdict: REVISE**

The Step 3 checklist repeats the prompt requirements, but it is not yet specific enough to execute safely. The mapper/logging boundary is in `internal/mcp/server.go`, not in `internal/tools/registry.go`, and the current plan does not say exactly how to preserve existing `UserError` public messages while adding `ErrInvalidInput` handling and operator logging.

## Required plan revisions

1. **Name the exact mapper/boundary to change.**
   - Update `internal/mcp/server.go`.
   - `publicToolErrorMessage(err)` is the LLM-facing mapper.
   - The `safeRegistrar.AddTool` handler closure is where logging can happen because it has the logger and tool name.
   - Do not move this into `tools.PublicErrorMessage`; that helper intentionally has no logger/tool context.

2. **Define the public-message precedence.**
   - Keep `tools.PublicErrorMessage(err)` first so existing `NewUserError(invalid...Message, err)` messages and wire shape stay unchanged.
   - Add an `errors.Is(err, tools.ErrInvalidInput)` fallback only for invalid-input errors that are not already wrapped in `UserError`.
   - The fallback must be a short constant such as `invalid tool arguments; check the inputs and try again`, not `err.Error()` and not the wrapped validation detail.
   - Keep unknown/non-validation errors mapped to `genericToolErrorMessage`.

3. **Define invalid-input logging separately from generic handler failures.**
   - In the handler error branch, log `errors.Is(err, tools.ErrInvalidInput)` at `Warn` with a message like `tool handler rejected invalid input`; keep non-validation handler errors at `Error`.
   - Do not log `req.Params.Arguments`, raw JSON, API keys, session data, or athlete IDs.
   - Be explicit about how the wrapped cause is logged. `slog` with `"error", err` on a `*tools.UserError` only logs the public message because `UserError.Error()` returns the public text; if the plan requires operator-visible validation detail, it must log the internal cause separately (for example after `errors.As` to `*tools.UserError`) while still respecting the no-secret/no-athlete-ID rule.

4. **Add targeted mapper/logging tests.**
   - Extend `TestPublicToolErrorMessageSanitizesUnknownErrors` (or add a sibling test) to cover:
     - `tools.NewUserError("try a valid test input", fmt.Errorf("%w: secret detail", tools.ErrInvalidInput))` returns the public message.
     - a bare `fmt.Errorf("%w: secret detail", tools.ErrInvalidInput)` returns the new short fallback and does not expose `secret detail`.
     - an unrelated error still returns `genericToolErrorMessage`.
   - Add a small test for the logging helper/branch, preferably by extracting a helper around the handler-error logging, using a buffer-backed `slog` logger, asserting invalid input is logged at `WARN` and non-validation errors remain `ERROR`. Also assert the log does not include raw request arguments or athlete IDs.

## Rationale

The five migrated validation sites are already wrapped by `NewUserError(...)` at their handlers, so `publicToolErrorMessage` already returns short LLM-facing messages for those paths. The Step 3 value is to make that behavior explicit and safe for any future bare `ErrInvalidInput`, plus to downgrade validation failures from operator `Error` logs to `Warn` while preserving enough internal detail for debugging. Without the precedence rule and tests, it is easy to accidentally leak the wrapped chain to the LLM or replace existing tool-specific invalid-argument messages with a generic one.
