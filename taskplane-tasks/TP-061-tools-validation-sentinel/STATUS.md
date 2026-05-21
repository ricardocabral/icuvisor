# TP-061-tools-validation-sentinel — Status

**Current Step:** Step 4: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-17
**Review Level:** 1
**Review Counter:** 5
**Iteration:** 1
**Size:** S

---

### Step 1: Decide on the sentinel home

**Status:** ✅ Complete

- [x] Read TP-048's plan for `UserError`. Decide: new `var ErrInvalidInput = errors.New("invalid input")` vs an `IsInvalidInput()` method on `UserError`. Record the decision in `STATUS.md`.
- [x] R001: Record the concrete ErrInvalidInput sentinel decision and guardrails in `STATUS.md`.

**Notes:**

- Decision: add a package-level `var ErrInvalidInput = errors.New("invalid input")` in `internal/tools/errors.go` with an exported doc note. Do not add `IsInvalidInput()` or validation categorization to `UserError`; keep `UserError` as the public-message wrapper and rely on `Unwrap()` plus `%w` so `errors.Is(err, ErrInvalidInput)` works through wrapped validation errors.
- Guardrails: scope the sentinel to `internal/tools`; preserve public error-message wording and wire shape; use `%w` at validation sites; do not let wrapped chains leak into LLM-visible strings; keep `PublicErrorMessage` precedence if Step 3 changes the mapper.

### Step 2: Wire the sentinel into the five sites

**Status:** ✅ Complete

- [x] At each cited validation site, switch bare validation `fmt.Errorf` calls to wrap `ErrInvalidInput` with `%w`.
- [x] Update or add tests for the cited validation paths to assert `errors.Is(err, ErrInvalidInput)`.
- [x] R002: Tighten the Step 2 plan with explicit sentinel creation, exact five migration paths, scope decision for `validateFloatMin`, wrapping pattern, and concrete tests.

**Notes:**

- Step 2 starts by creating `internal/tools/errors.go` with exported `ErrInvalidInput = errors.New("invalid input")`; package-local call sites use `ErrInvalidInput`.
- Migrate exactly these five current validation paths: `validateActivitiesTokenArgs` unsupported token version; `decodeDeleteEventsByDateRangeRequest` range exceeds `maxDeleteEventsByDateRangeDays`; `applyTrainingPlan` no workouts with relative-day metadata; `validateIntRange` out-of-range wellness int scales; `validateIntMin` negative wellness minimum fields.
- Scope decision: `validateFloatMin` is intentionally out of scope for TP-061 because it was not one of the five audit-cited sites; do not migrate it unless a later review explicitly expands scope.
- Wrapping pattern: `fmt.Errorf("%w: existing message", ErrInvalidInput)`; keep existing `NewUserError(..., err)` handler wrapping so `errors.Is` traverses through `UserError.Unwrap()` and public messages remain unchanged until Step 3.
- Test plan: add or update targeted tests for unsupported activity token version, too-long delete-events date range, apply-training-plan with no relative-day workout metadata, wellness `feel: 6`, and wellness `restingHR: -1`, all asserting `errors.Is(err, ErrInvalidInput)`.

### Step 3: Route through the LLM-facing mapper

**Status:** ✅ Complete

- [x] At the registry/handler boundary, return a short LLM-facing string for `ErrInvalidInput` without exposing the wrapped error chain.
- [x] Log the full wrapped invalid-input error via `slog.Warn` for the operator without API keys or athlete IDs.
- [x] R004: Tighten the Step 3 plan with the exact `internal/mcp/server.go` mapper/logging boundary, public-message precedence, logging detail handling, and targeted tests.

**Notes:**

- Change `internal/mcp/server.go`: `publicToolErrorMessage(err)` is the LLM-facing mapper; `safeRegistrar.AddTool`'s handler closure is the logging boundary because it has logger and tool name.
- Public-message precedence: keep `tools.PublicErrorMessage(err)` first; add a bare `errors.Is(err, tools.ErrInvalidInput)` fallback returning a short constant such as `invalid tool arguments; check the inputs and try again`; keep unrelated errors on `genericToolErrorMessage`.
- Logging plan: in the handler error branch, log invalid-input failures at `Warn` (`tool handler rejected invalid input`) and non-validation failures at `Error`; never log raw request arguments, API keys, session data, or athlete IDs.
- Wrapped detail plan: log the public error as `"error", err`; when `errors.As(err, *tools.UserError)` exposes an internal cause, also log `"cause", userErr.Unwrap()` so validation detail is operator-visible without request payloads.
- Test plan: add mapper tests for UserError precedence, bare ErrInvalidInput fallback redaction, and generic fallback; add buffer-backed slog tests for invalid-input Warn vs generic Error logging and absence of request/athlete payloads.

### Step 4: Verify

**Status:** ✅ Complete

- [x] Update `CHANGELOG.md` under `[Unreleased]` / `Changed` for the invalid-input sentinel behavior.
- [x] `grep -n 'fmt.Errorf("[^%]' internal/tools/` for the five sites confirms each migrated site uses `%w`.
- [x] `grep -rn 'err.Error() ==' internal/tools/` returns zero hits.
- [x] `make build`, `make test`, `make test-race`, and `make lint` pass.
- [x] Commit: `TP-061 introduce ErrInvalidInput sentinel for tool validation`.

| 2026-05-17 01:53 | Task started | Runtime V2 lane-runner execution |
| 2026-05-17 01:53 | Step 1 started | Decide on the sentinel home |
| 2026-05-17 01:56 | Review R001 | plan Step 1: REVISE |
| 2026-05-17 02:01 | Review R002 | plan Step 2: REVISE |
| 2026-05-17 02:03 | Review R003 | plan Step 2: APPROVE |
| 2026-05-17 02:10 | Review R004 | plan Step 3: REVISE |
| 2026-05-17 02:12 | Review R005 | plan Step 3: APPROVE |

| 2026-05-17 02:19 | Worker iter 1 | done in 1504s, tools: 126 |
| 2026-05-17 02:19 | Task complete | .DONE created |