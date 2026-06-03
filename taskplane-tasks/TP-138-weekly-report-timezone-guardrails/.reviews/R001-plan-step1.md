# Plan Review: Step 1 — Audit weekly/date-window safeguards

Result: Approved with one scope note.

The Step 1 plan is appropriate for an audit-only step: it reads the relevant prompt/golden-test surface, checks whether wellness and `_meta.as_of` guidance prevents stale/current-day leakage, records discoveries in `STATUS.md`, and runs the targeted package tests. I ran `go test ./internal/prompts ./internal/tools`; both packages pass from cache.

Scope note: include `internal/tools/get_today_test.go` in the Step 1 audit. It is listed in the task file scope and covers daily `_meta.as_of` behavior, but it is not listed under the Step 1 artifacts. Since the mission explicitly mentions adjacent daily-report wrong-date bugs, the audit should confirm `get_today` as-of/date behavior alongside `as_of_test.go` and `get_wellness_data_test.go`.

Expected audit findings to record before Step 2:
- Existing weekly/plan-health prompts already mention athlete-local dates/timezone and stale/missing wellness caveats.
- They do not yet explicitly forbid using wellness rows after the requested report window, nor do they clearly label current-day `_meta.as_of` data as partial-day context only for prior-week reports.
- Existing tests check prompt stale strings and tool as-of metadata presence/omission, but prompt regression coverage for the new after-window/partial-day caveats still needs to be added in Step 2.
