# Plan Review — Step 3: Extract cursor codec

Result: **Approved with required adjustments**

The Step 3 plan is mostly aligned with TP-066: moving the activities page-token type, encode/decode path, token-argument validation, and cursor progression helpers into `internal/tools/get_activities_cursor.go` can be behavior-preserving and should materially reduce `get_activities.go` without changing the tool schema or wire shape.

## Required adjustments

1. **Match the prompt's verification scope or explicitly document the deferral.** `PROMPT.md` says Step 3 should “Run all checks,” while `STATUS.md` currently says only “Run targeted tool tests for cursor pagination and token invariants.” Either update the Step 3 plan to run the requested broader check (`make test`, or the repo's agreed equivalent plus targeted `go test ./internal/tools`), or add a clear `STATUS.md` note that full-suite/build/lint verification is intentionally deferred to Step 5.

2. **Keep hard-coded golden tokens, do not regenerate them.** The byte-identical requirement is only meaningful if the existing pre-refactor token constants remain authoritative. Moving tests is fine, but avoid replacing them with self-referential encode/decode tests that would pass after an accidental token-format change.

3. **Define the extraction boundary narrowly.** `get_activities_cursor.go` should own token structs, parse/encode, token-argument validation, and cursor/boundary advancement helpers. Leave handler glue, response shaping, schema literals, and upstream fetch orchestration (`fetchActivitiesPage` / `iteratePages`, unless there is a very specific reason) in `get_activities.go` for this step. This keeps Step 3 a codec/cursor extraction rather than an unplanned pagination-driver rewrite.

## Implementation notes to keep the plan safe

- Preserve the exact `activitiesPageToken` JSON field order/tags and `base64.RawURLEncoding` behavior. Reordering fields in the struct can change the emitted JSON bytes and therefore the opaque token string.
- Keep `normalizeActivitiesPageSize`, `activitiesTokenArgs`, `validateActivitiesTokenArgs`, `parseActivitiesPageToken`, and `encodeActivitiesPageToken` package-private; no exported API is needed.
- Keep athlete binding behavior intact: generated tokens must still include `athlete_id` when a target athlete is resolved, and handler validation must still reject a token whose athlete does not match the resolved athlete.
- Preserve boundary behavior around same-timestamp filtered rows: `SkipIDsAtBoundary`, `BeforeID`, `BeforeStartDateLocal`, `justBeforeActivityTimestamp`, and the max-fetch boundary error are part of the pagination contract even if they are internal.
- Move or add focused tests in `get_activities_cursor_test.go` for parse/encode invariants, unsupported token versions, mismatched supplied args, athlete binding, and the existing full-page golden token. Keep at least one handler/response integration assertion that `_meta.next_page_token` is byte-identical so the wire shape remains covered.
- After the move, run `gofmt`/`goimports` and clean imports in `get_activities.go` (`bytes`, `base64`, `slices`, and `time` are likely to move with the cursor code; `fmt`, `math`, `sort`, and `strings` may still be needed depending on what remains).

With those adjustments, the plan is suitably small and behavior-preserving for Step 3.
