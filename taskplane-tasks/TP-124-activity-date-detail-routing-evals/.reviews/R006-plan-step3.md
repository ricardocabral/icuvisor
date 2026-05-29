# Review R006 — Plan Step 3

Verdict: Revise

The Step 3 plan is directionally correct, but it is not specific enough for the gaps already recorded in Step 1/2. The current checklist/artifact set focuses on `get_activities`, `get_activity_details`, and the cookbook, while the discoveries and new evals also depend on `get_activity_intervals` and `get_activity_splits` routing. `get_activity_splits` lives in `internal/tools/get_activity_streams.go`, which is currently omitted from the Step 3 artifact/checklist scope.

Blocking plan updates:

- Add `internal/tools/get_activity_streams.go` to Step 3 scope if hardening split/stream activation hints. The split/reps eval expects `get_activity_splits`, so the plan should explicitly cover its description (or explicitly justify why cookbook-only guidance is sufficient).
- Make the intended hinting precise and concise: downstream tools that require `activity_id` should say to resolve described/date-based activities with `get_activities` first, using the athlete-local date window, then pass the returned `activity_id`.
- If any tool description/catalog text changes, include the generated docs follow-up in the plan (`make docs-tools`, with the resulting generated file committed as applicable) or explicitly defer it to Step 5. The project guidance says generated tool reference data must stay in sync with the registry.
- Keep the existing targeted validation: `go test ./internal/tools ./internal/prompts` and `make eval-validate`.

Once those scope/details are added, the plan should be ready to implement.
