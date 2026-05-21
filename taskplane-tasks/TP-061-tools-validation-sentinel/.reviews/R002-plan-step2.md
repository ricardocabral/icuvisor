# R002 — Plan review: Step 2

**Verdict: REVISE**

The Step 2 checklist is directionally correct, but it is too underspecified to be safe to execute. In particular, Step 1 recorded only the decision; the repository does not yet contain `internal/tools/errors.go` or `ErrInvalidInput`, so Step 2 must either create the sentinel at its start or explicitly reopen Step 1 to add it before any call sites are changed.

## Required plan revisions

1. **Make the sentinel creation explicit before wiring.**
   - Add `internal/tools/errors.go` with `var ErrInvalidInput = errors.New("invalid input")` and an exported doc note, or add a Step 2 precondition that this file already exists.
   - Inside `internal/tools`, call sites should use `ErrInvalidInput` (not `tools.ErrInvalidInput`).

2. **List the exact five current code paths to migrate.** The prompt line numbers are stale after TP-048, so the plan should name functions/branches, not only line numbers:
   - `validateActivitiesTokenArgs`: unsupported token version (`fmt.Errorf("unsupported token version %d", ...)`).
   - `decodeDeleteEventsByDateRangeRequest`: range exceeds `maxDeleteEventsByDateRangeDays`.
   - `applyTrainingPlan`: plan has no workouts with relative day metadata.
   - `validateIntRange`: out-of-range wellness int scales.
   - `validateIntMin`: negative wellness int minimum fields.

   Also record whether `validateFloatMin` is intentionally out of scope. It is another bare validation `fmt.Errorf` in the same file today; if it is not part of the five audit sites, do not accidentally fold it into this step without noting the scope decision.

3. **Specify the wrapping pattern and preserve error-chain behavior.**
   - Use `fmt.Errorf("%w: existing message", ErrInvalidInput)` at the migrated sites.
   - Keep the existing `NewUserError(..., err)` wrapping at handlers so `errors.Is(err, ErrInvalidInput)` works through `UserError.Unwrap()` while the public message remains unchanged until Step 3.
   - Do not change unrelated `errors.New(...)` validation branches in this step unless the plan is intentionally expanded.

4. **Define concrete tests for each migrated path.** Existing tests only assert “some error” in several places, and some current invalid cases do not hit the cited `fmt.Errorf` branches. The plan should add or update tests such as:
   - `get_activities`: construct or decode a token with an unsupported version and assert `errors.Is(err, ErrInvalidInput)`; the existing mismatched-token test exercises an `errors.New` branch and is not enough for the cited `fmt.Errorf` site.
   - `delete_events_by_date_range`: for the too-long range case, assert `errors.Is(err, ErrInvalidInput)`; keep other invalid-date/order cases as generic errors unless they are migrated.
   - `apply_training_plan`: use a fake plan/workout set with no relative-day metadata and assert `errors.Is(err, ErrInvalidInput)` on the handler error.
   - `update_wellness`: add separate cases for `validateIntRange` (for example `feel: 6`) and `validateIntMin` (for example `restingHR: -1`) and assert `errors.Is(err, ErrInvalidInput)`; do not rely only on the current mixed loop.

## Rationale

The task acceptance criteria depend on `errors.Is`, not just replacing `fmt.Errorf` text. Without a concrete migration/test matrix, it is easy to update the wrong token-validation branch, miss the apply-training-plan runtime validation path, or add sentinel assertions to invalid cases that still return plain `errors.New`. Tightening the Step 2 plan will keep the change small and aligned with the “five cited sites only” scope.
