# Code Review R022 — Step 6: Trim inline tool descriptions

**Verdict: REVISE**

I reviewed `git diff c704353..HEAD`, read the changed tool/schema/README/status files, and ran:

- `go test ./...` — passes
- `make lint` — passes
- `go run ./scripts/check_confusable_names.go` — passes
- `go run ./scripts/check_schema_stability.go` — passes for snapshot freshness only
- `go run ./scripts/check_schema_stability.go -baseline-dir <c704353 schema snapshots>` — **fails**

## Blocking findings

1. **Schema-stability CI guard fails because existing input-property schemas changed.**

   The task explicitly requires TP-015 schema-stability to stay green, and CI runs `check_schema_stability.go` with a baseline directory. Reproducing that against `c704353` fails with `property-changed` for existing arguments:

   - `internal/tools/add_or_update_event.go:170` (`category`)
   - `internal/tools/add_or_update_event.go:174` (`workout_doc`)
   - `internal/tools/create_workout.go:126` (`workout_doc`)
   - `internal/tools/get_events.go:250` (`category`)
   - `internal/tools/update_workout.go:195` (`workout_doc`)

   Command used:

   ```sh
   tmp=$(mktemp -d)
   git archive c704353 internal/tools/schema_snapshot | tar -x -C "$tmp"
   go run ./scripts/check_schema_stability.go \
     -baseline-dir "$tmp/internal/tools/schema_snapshot"
   ```

   Result:

   ```text
   additive-only stability: FAIL (5 issue(s))
   ... kind=property-changed property=category ... stable argument schemas are additive-only
   ... kind=property-changed property=workout_doc ... stable argument schemas are additive-only
   ```

   Please restore the existing input-schema property descriptions (and corresponding snapshots) or add an intentional compatibility mechanism/allowance before merging. The Step 6 trim can still happen in top-level `Tool.Description` strings and response metadata without mutating stable argument schemas.

## Non-blocking note

- `STATUS.md` records R021 as `APPROVE`, but the added `.reviews/R021-plan-step6.md` content says `Verdict: REVISE`. Please reconcile the task bookkeeping so the review log is accurate.
