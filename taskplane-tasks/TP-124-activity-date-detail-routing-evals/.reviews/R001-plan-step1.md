# Review R001 — Plan Step 1

Verdict: Changes requested

The Step 1 plan covers the main list → detail/interval path, but it is not quite sufficient for the task as written.

## Blocking gap

- The task explicitly includes “lap splits or reps” and Step 2 will add a split/reps scenario, but Step 1 only plans to inspect `get_activity_intervals`. It should also inspect `get_activity_splits` in `internal/tools/get_activity_streams.go` and its current description/schema. Otherwise the mapping can miss the routing hint that models need for “splits” prompts.

## Recommended plan adjustments

- Expand the first Step 1 checkbox to inspect:
  - `internal/tools/get_activities.go`
  - `internal/tools/get_activity_details.go`
  - `internal/tools/get_activity_streams.go` (`get_activity_splits`)
  - cookbook activity-retrospective guidance
  - existing cookbook eval scenarios
  - prompt testdata where activity routing could apply
- When recording Discoveries, separate:
  - date resolution hints for relative athlete-local dates like “last Sunday”
  - ID-routing hints from `get_activities` to detail/interval/splits tools
  - any missing or ambiguous split-vs-interval wording
- Keep the targeted test command as planned: `go test ./internal/tools ./internal/prompts`.

Once `get_activity_splits` is included in the mapping, the plan is appropriate for a discovery-only step.
