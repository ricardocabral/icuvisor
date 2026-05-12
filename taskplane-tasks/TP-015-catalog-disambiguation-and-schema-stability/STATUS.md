# TP-015 — Status

**Issue:** v0.2 — read path
**Status:** ✅ Complete
**Review Level:** 2
**Iteration:** 3
**Current Step:** Step 5: Tests and verify
**Last Updated:** 2026-05-12
**State:** Complete

_Task scaffolded from PROMPT.md; hydrated for execution._

## Execution Log

| Time             | Event                 | Notes                                                        |
| ---------------- | --------------------- | ------------------------------------------------------------ |
| 2026-05-12 03:46 | Task started          | Runtime V2 lane-runner execution                             |
| 2026-05-12 03:46 | Step 1 started        | Audit confusable clusters                                    |
| 2026-05-12 03:46 | Hydrated              | Expanded STATUS.md with task steps and checkboxes            |
| 2026-05-12 03:46 | Step 1 plan review    | R001 requested concrete cluster inventory and audit evidence |
| 2026-05-12 03:46 | Step 1 plan review    | Approved after plan hydration                                |
| 2026-05-12 03:46 | Step 1 targeted tests | `go test ./internal/tools` passed after description rewrites |
| 2026-05-12 04:42 | Exit intercept close  | Supervisor directed session close: "close"                   |
| 2026-05-12 04:42 | Worker iter 1         | done in 3398s, tools: 187                                    |
| 2026-05-12 04:42 | No progress           | Iteration 1: 0 new checkboxes (1/3 stall limit)              |
| 2026-05-12 04:42 | Step 1 started        | Audit confusable clusters                                    |
| 2026-05-12 04:44 | Exit intercept close | Supervisor directed session close: "close" |
| 2026-05-12 04:44 | Worker iter 2 | done in 91s, tools: 10 |
| 2026-05-12 04:44 | No progress | Iteration 2: 0 new checkboxes (2/3 stall limit) |
| 2026-05-12 04:44 | Step 1 started | Audit confusable clusters |

## Discoveries

| Time | Area | Finding |
| ---- | ---- | ------- |

## Blockers

| Time | Blocker | Attempts |
| ---- | ------- | -------- |

## Notes

- Review Level 2 requires plan and code reviews for implementation steps before marking step status complete.
- Step 2 canonical schema snapshots are live-registry `Tool.InputSchema` JSON files, one per tool, encoded with Go `encoding/json.MarshalIndent(value, "", "  ")`; string map keys are sorted by the encoder and every file has a trailing newline.
- Step 4 confusable-name threshold is token Jaccard >= 0.58 on normalized first description sentences within a domain cluster; current v0.2 catalog passes after Step 1 rewrites.

## Step 1: Audit confusable clusters

**Status:** ✅ Complete

### Plan details

- Concrete clusters to audit from `internal/tools/registry.go` and description constants:
  - Activity domain: `get_activities`, `get_activity_details`, `get_activity_intervals`, `get_activity_streams`, `get_activity_splits`, `get_activity_messages`, `get_extended_metrics`
  - Event/calendar domain: `get_events`, `get_event_by_id`, plus relationship note for `get_training_plan` because training-plan wording can overlap calendar/workout-plan language
  - Workout library domain: `get_workout_library`, `get_workouts_in_folder`
  - Custom items domain: `get_custom_items`, `get_custom_item_by_id`
  - Fitness/performance domain: `get_fitness`, `get_training_summary`, `get_best_efforts`, `get_power_curves`
  - Wellness domain: `get_wellness_data` singleton/no-op entry so the audit records that no cluster mate exists today
- First-sentence extraction method: inspect the Go description constants and extract through the first sentence boundary manually, preserving `intervals.icu` rather than splitting on every period.
- Evidence method: record every audited tool in the table below with cluster, file, before sentence, after sentence, and action/reason; unchanged tools must say why the first sentence already distinguishes access pattern and payload.
- Scope: edit only first description sentences in `internal/tools/*.go`, run targeted `go test ./internal/tools` if descriptions change, and defer schema snapshots/CI helper work to later steps.

### Step 1 audit evidence

| Cluster                    | Tool                     | File                                       | Before first sentence                                                                                                                          | After first sentence                                                                                                                           | Action / reason                                                                                                |
| -------------------------- | ------------------------ | ------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| Activity domain            | `get_activities`         | `internal/tools/get_activities.go`         | List activities for a date range with terse unit-disambiguated rows, Strava-unavailable detection, and opaque pagination.                      | List activities for a date range with terse unit-disambiguated rows, Strava-unavailable detection, and opaque pagination.                      | Unchanged — clearly lists activity rows and says to use before activity detail/interval/stream tools.          |
| Activity domain            | `get_activity_details`   | `internal/tools/get_activity_details.go`   | Get one activity's terse metadata and metrics by activity_id.                                                                                  | Get one activity's terse metadata and metrics by activity_id.                                                                                  | Unchanged — clearly retrieves one activity's metadata/metrics, not streams, splits, intervals, or messages.    |
| Activity domain            | `get_activity_intervals` | `internal/tools/get_activity_details.go`   | Get analyzed intervals for one activity by activity_id.                                                                                        | Get analyzed intervals for one activity by activity_id.                                                                                        | Unchanged — clearly retrieves analyzed interval rows.                                                          |
| Activity domain            | `get_activity_streams`   | `internal/tools/get_activity_streams.go`   | Get canonical activity stream channels by activity_id.                                                                                         | Get canonical activity stream channels by activity_id.                                                                                         | Unchanged — clearly targets heavy stream channels/samples.                                                     |
| Activity domain            | `get_activity_splits`    | `internal/tools/get_activity_streams.go`   | Get manual or virtual per-km/per-mile activity splits.                                                                                         | Get manual or virtual per-km/per-mile activity splits.                                                                                         | Unchanged — clearly targets split rows derived from intervals/streams.                                         |
| Activity domain            | `get_activity_messages`  | `internal/tools/get_activity_messages.go`  | List comments and notes on one activity by activity_id.                                                                                        | List comments and notes on one activity by activity_id.                                                                                        | Unchanged — clearly targets comments/notes only.                                                               |
| Activity domain            | `get_extended_metrics`   | `internal/tools/get_extended_metrics.go`   | Get upstream-exposed extended activity metrics only.                                                                                           | Get one activity's upstream-exposed extended metrics by activity_id.                                                                           | Rewritten to state one-activity access pattern and distinguish extended metrics from list/detail reads.        |
| Event/calendar domain      | `get_events`             | `internal/tools/get_events.go`             | List intervals.icu calendar events for a bounded athlete-local YYYY-MM-DD date range.                                                          | List calendar events across a bounded athlete-local YYYY-MM-DD date range.                                                                     | Rewritten to emphasize range/list access and avoid first-sentence extraction ambiguity around `intervals.icu`. |
| Event/calendar domain      | `get_event_by_id`        | `internal/tools/get_event_by_id.go`        | Fetch one intervals.icu calendar event by ID.                                                                                                  | Fetch a single calendar event detail by event_id.                                                                                              | Rewritten to emphasize single-detail access by `event_id`.                                                     |
| Event/calendar domain      | `get_training_plan`      | `internal/tools/get_training_plan.go`      | Fetch the active intervals.icu training-plan assignment exposed by the public API.                                                             | Fetch the athlete's active training-plan assignment, not calendar events or workout-library templates.                                         | Rewritten to disambiguate training-plan assignment from calendar events and workout-library templates.         |
| Workout library domain     | `get_workout_library`    | `internal/tools/get_workout_library.go`    | List intervals.icu workout-library folders and plans.                                                                                          | List workout-library folders and plans, not calendar events or the active training-plan assignment.                                            | Rewritten to distinguish library folders/plans from calendar events and active assignment.                     |
| Workout library domain     | `get_workouts_in_folder` | `internal/tools/get_workouts_in_folder.go` | List workout-library templates inside one intervals.icu folder or plan by folder ID.                                                           | List workout-library templates inside one folder or plan by folder_id.                                                                         | Rewritten to emphasize child template list by `folder_id` and avoid `intervals.icu` split ambiguity.           |
| Custom items domain        | `get_custom_items`       | `internal/tools/get_custom_items.go`       | List intervals.icu custom items such as charts, fields, streams, panels, histograms, maps, and zones.                                          | List custom item definitions such as charts, fields, streams, panels, histograms, maps, and zones.                                             | Rewritten to emphasize list/catalog rows and avoid `intervals.icu` split ambiguity.                            |
| Custom items domain        | `get_custom_item_by_id`  | `internal/tools/get_custom_item_by_id.go`  | Fetch one intervals.icu custom item by ID and return the full item including its item_type-specific content payload.                           | Fetch one custom item by item_id and return the full item_type-specific content payload.                                                       | Rewritten to emphasize detail access by `item_id` and full content payload.                                    |
| Fitness/performance domain | `get_fitness`            | `internal/tools/get_fitness.go`            | Get CTL, ATL, and TSB fitness trends for a local date range.                                                                                   | Get CTL, ATL, and TSB fitness trends for a local date range.                                                                                   | Unchanged — clearly distinguishes fitness trend time series.                                                   |
| Fitness/performance domain | `get_training_summary`   | `internal/tools/get_fitness.go`            | Get aggregated training volume, neutral training load, sRPE, and upstream zone-order totals for a local date range.                            | Get aggregated training volume, neutral training load, sRPE, and upstream zone-order totals for a local date range.                            | Unchanged — clearly distinguishes aggregate volume/load summary.                                               |
| Fitness/performance domain | `get_best_efforts`       | `internal/tools/get_fitness.go`            | Get upstream best efforts grouped by sport and default/requested power, heart-rate, and pace buckets.                                          | Get upstream best efforts grouped by sport and default/requested power, heart-rate, and pace buckets.                                          | Unchanged — clearly targets best-effort bucket records.                                                        |
| Fitness/performance domain | `get_power_curves`       | `internal/tools/get_fitness.go`            | Get the upstream-computed mean-maximal power curve for a date range.                                                                           | Get the upstream-computed mean-maximal power curve for a date range.                                                                           | Unchanged — clearly targets mean-maximal power curve buckets.                                                  |
| Wellness domain            | `get_wellness_data`      | `internal/tools/get_wellness_data.go`      | Get daily wellness rows for a local date range with distinct sleepQuality, sleepScore, sleepSecs, custom fields, and native provider sidecars. | Get daily wellness rows for a local date range with distinct sleepQuality, sleepScore, sleepSecs, custom fields, and native provider sidecars. | Singleton/no-op — no sibling read tool exists today, but recorded for audit completeness.                      |

- [x] Enumerate read-tool clusters that share a prefix or domain (`get_activity_*`, `get_wellness_*`, `get_workout_*`, `get_custom_item*`, `get_event*`)
- [x] For each cluster, read every member's first description sentence; rewrite where the access pattern is not obvious from name alone (the §7.2.E rule)
- [x] Record before/after first sentences in `STATUS.md`

## Step 2: Snapshot every tool's argument schema

**Status:** ✅ Complete

### Plan details

- Add a small Go snapshot generator under `scripts/` that calls `tools.NewRegistryWithOptions` with a local fake catalog client implementing every current read-tool client interface, then captures `Tool.Name` and `Tool.InputSchema` through a `tools.Registrar` implementation. This keeps snapshots sourced from the live registry instead of hand-authored constants.
- Write one canonical JSON file per registered tool to `internal/tools/schema_snapshot/<tool_name>.json`; use `encoding/json` with two-space indentation and sorted map keys (the standard encoder sorts string map keys) plus a trailing newline so diffs are stable and reviewable.
- Include all currently registered read tools, including singleton/no-argument schemas, so later write tools automatically gain snapshots when registered.
- Document in `CONTRIBUTING.md` that maintainers regenerate snapshots with the script after additive optional arguments, and that stable-tool argument removals/renames require a new tool name rather than editing the existing schema.

- [x] Generate a JSON-Schema snapshot per tool from the live registry; commit under `internal/tools/schema_snapshot/<tool_name>.json`
- [x] Define the canonical serialization (key ordering, indentation) so diffs are reviewable
- [x] Document the snapshot-update workflow in `CONTRIBUTING.md`: snapshots may grow (new optional arguments) but cannot shrink or rename existing arguments on a stable tool; the only way to "rename" is to ship a new tool name
- [x] Add descriptions/defaults to activity stream and split input schema properties, then regenerate affected snapshots
- [x] Make the snapshot generator remove or report stale JSON files so the snapshot directory exactly matches the live registry

## Step 3: Implement the CI schema-stability check

**Status:** ✅ Complete

### Plan details

- Add reusable schema snapshot/check logic in an internal Go package so the Step 2 generator, the Step 3 CLI, and Step 5 tests use the same live-registry snapshot source and canonical JSON serialization.
- Implement `scripts/check_schema_stability.go` as a build-ignore command with two explicit comparisons: (1) a freshness/canonicalization check that generated current live-registry schemas exactly match the snapshot files committed in the current tree, and (2) a stability check that generated current live-registry schemas are additive-only relative to a pre-PR baseline snapshot directory from the PR base/merge-base or an explicit `-baseline-dir` supplied by CI/tests.
- Enforce additive-only at the argument-property level for stable tools: every baseline tool must still exist in the current registry (missing baseline tools fail as removals/renames), every baseline property must still exist with the same property schema, existing `required` semantics/root invariants may not become more restrictive, and new properties are allowed only when optional. A new current tool with no baseline is accepted only as an addition and must still have a committed current snapshot via the freshness check.
- Emit GitHub Actions annotations (`::error file=...`) plus a Markdown summary to `$GITHUB_STEP_SUMMARY` when available; also print the failing-tool list and actionable hints to stderr/stdout locally, including the baseline file, current snapshot/generated file, tool, and property when available.
- Add compile-time interface assertions around the all-tools fake catalog client so future registry interface changes cannot silently omit tools from snapshot generation/checking.

- [x] On every PR, regenerate snapshots and diff against the checked-in versions
- [x] **Additive-only** rule: new properties are allowed; removed or renamed properties fail the check
- [x] An override mechanism exists for genuine new-tool introductions: a new tool name produces a new snapshot file with no prior version, which the check accepts
- [x] Output the failing-tool list as an annotated CI summary, not just a non-zero exit
- [x] Fail loudly when `-require-baseline` stability mode points at a missing or empty baseline snapshot directory
- [x] Fix `gosec` file-permission lint in `WriteGeneratedSchemaSnapshots`
- [x] Include current/generated snapshot paths in stability failures and annotations

## Step 4: Implement the confusable-names check

**Status:** ✅ Complete

### Plan details

- Reuse the live registry catalog source from `internal/toolchecks` to collect tool names and descriptions, then extract first sentences with a helper that treats `intervals.icu` and other dotted tokens as non-boundaries.
- Build prefix/domain clusters from the current read catalog: `get_activity*`, event/calendar plus `get_training_plan`, workout-library, custom-items, fitness/performance, and wellness singleton/no-op. The checker should skip singleton clusters and compare only within multi-tool clusters.
- Use token Jaccard similarity on normalized first-sentence tokens; fail pairs at or above a documented threshold after Step 1 rewrites. Emit actionable messages that name both tools and suggest rewriting the first sentence to emphasize access pattern/payload.
- Add `scripts/check_confusable_names.go` as a build-ignore CLI with GitHub annotations and summary output matching the schema checker style.

- [x] Compute a similarity metric (token Jaccard or Levenshtein on the first description sentence) within each prefix cluster
- [x] Threshold tuned so existing v0.2 clusters pass after Step 1's rewrites; new PRs that push two tools above threshold fail with a suggested-rewrite hint
- [x] Document the threshold in `CONTRIBUTING.md`
- [x] Include future `get_event*` tools in the event/calendar confusability cluster while preserving `get_training_plan` as a cross-domain alias

## Step 5: Tests and verify

**Status:** ✅ Complete

### Plan details

- Add table-driven unit tests in `internal/toolchecks` for schema freshness/stability and confusable-name behavior using temporary snapshot directories and synthetic catalogs.
- Wire CI to prepare a baseline snapshot directory from the merge base/base ref before running `check_schema_stability.go -baseline-dir ... -require-baseline`, then run `check_confusable_names.go`.
- Update `CHANGELOG.md` for the user-visible catalog wording and CI guard additions, then run full verification commands and local guard checks.

- [x] Unit tests for both CI helpers covering: clean diff (pass); added argument (pass); removed argument (fail); renamed argument (fail); new tool (pass); confusable pair (fail); rewritten pair (pass)
- [x] Wire both checks into `.github/workflows/ci.yml`
- [x] Update `CHANGELOG.md` under `[Unreleased]`
- [x] `make test`, `make build`, `make lint` pass
- [x] Run the checks locally against the v0.2 catalog; commit the resulting snapshots
      | 2026-05-12 03:49 | Review R001 | plan Step 1: REVISE |
      | 2026-05-12 03:52 | Review R001 | plan Step 1: APPROVE |
      | 2026-05-12 03:58 | Review R001 | code Step 1: APPROVE |
      | 2026-05-12 04:01 | Review R001 | plan Step 2: APPROVE |
      | 2026-05-12 04:06 | Review R001 | code Step 2: REVISE |
      | 2026-05-12 04:10 | Review R001 | code Step 2: APPROVE |
      | 2026-05-12 04:13 | Review R001 | plan Step 3: REVISE |
      | 2026-05-12 04:15 | Review R001 | plan Step 3: APPROVE |
      | 2026-05-12 04:20 | Review R001 | code Step 3: REVISE |
      | 2026-05-12 04:25 | Review R001 | code Step 3: APPROVE |
      | 2026-05-12 04:28 | Review R001 | plan Step 4: APPROVE |
      | 2026-05-12 04:33 | Review R001 | code Step 4: REVISE |
      | 2026-05-12 04:38 | Review R001 | code Step 4: APPROVE |
