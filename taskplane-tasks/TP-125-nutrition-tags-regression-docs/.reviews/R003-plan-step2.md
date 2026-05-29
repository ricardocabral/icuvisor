# Plan Review: Step 2 — Fill regression or docs gaps

Verdict: Needs changes

The revised regression-test scope addresses the prior blocking issue: Step 2 now explicitly calls for default/no-`include_full` tag preservation regressions for both present and empty tags in `get_activities` and `get_activity_details`. The plan also correctly avoids behavior changes and keeps the targeted `go test ./internal/tools` gate.

Remaining blocking plan gap:
- `CHANGELOG.md` is a Must Update file in the task prompt, and Step 2 is expected to change tests/docs. The current Step 2 checklist only mentions user-facing docs/cookbook text, while Step 4 still has a generic “Must Update docs modified” item. Add an explicit checkbox or handoff note to update `CHANGELOG.md` under `[Unreleased]` before execution completes, so the documentation requirement is not missed.

Once that explicit changelog item is added, the Step 2 plan is acceptable.
