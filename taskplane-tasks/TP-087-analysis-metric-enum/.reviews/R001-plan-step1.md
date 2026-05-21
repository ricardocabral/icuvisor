# Plan Review — TP-087 Step 1: Design the enum and aliases

## Verdict: Needs changes before implementation

`STATUS.md` currently only repeats the Step 1 checklist from the prompt. It does not yet contain the actual enum inventory, canonical-name decisions, alias policy, or unknown-metric hint strategy that this design step is meant to produce. Because the later implementation will create a shared analyzer contract, I would not proceed to Step 2 until those decisions are recorded explicitly.

## Required plan fixes

1. **Record the supported metric inventory from existing read-tool fields.**
   The plan should name the first-pass enum entries and cite their source read surfaces. At minimum it should cover the roadmap/PRD examples and the currently exposed fields from:

   - `get_fitness`: `ctl`, `atl`, `tsb` (and whether `ramp` maps to wellness `rampRate` or remains excluded until a read tool exposes a canonical analyzer source).
   - `get_training_summary`: `training_load` / TSS-equivalent naming, `weekly_tss`, `weekly_hours`, `time_seconds`, `moving_time_seconds`, `elapsed_time_seconds`, `distance_*`, `elevation_gain_m`, `calories_burned`, `session_rpe`, zone-time totals.
   - `get_wellness_data`: `rhr`/`resting_hr`, `hrv`, `hrv_sdnn`, `weight`, `sleep_secs`, `sleep_score`, `sleep_quality`, `feel`, `fatigue`, `soreness`, `stress`, `mood`, `motivation`, nutrition fields, readiness, SpO2, respiration, etc., with scale-sensitive fields called out.
   - Activity/detail fields: duration, distance, pace, speed, elevation, heart rate, cadence, calories, training load.
   - `get_extended_metrics`: `if`, `vi`, `np` only if there is an actual existing source, `pw_hr`, decoupling/drift fields, zone distributions, joules above FTP, polarization index, TRIMP, load variants, L/R balance, RPE/feel/compliance.
   - Curve/best-effort surfaces: power/HR duration efforts and pace distance efforts should either be represented as structured analyzer arguments later or explicitly excluded from `analysis_metric` with a hint to `analyze_efforts_delta`.

   This inventory is the guardrail against accidentally accepting raw upstream keys, raw `include_full` fields, or aspirational metrics not yet exposed by current read tools.

2. **Define canonical naming conventions.**
   The plan should state whether canonical enum values are always `snake_case`, whether units are embedded in names only when needed (`sleep_secs`, `distance_km` vs unit-normalized `distance`), and how abbreviations are handled (`rhr`, `hrv`, `if`, `vi`, `np`, `pw_hr`). It should also decide how to handle collisions such as wellness `fatigue` vs fitness ATL/fatigue, and activity/session `training_load` vs weekly/load aggregates.

3. **Make the alias policy explicit and conservative.**
   Safe aliases should be listed, not left to ad hoc normalization. For example, `resting_hr -> rhr`, `restingHR -> rhr`, `sleepSecs -> sleep_secs`, `sleepQuality -> sleep_quality`, `intensity_factor -> if`, and `variability_index -> vi` may be reasonable if documented. The plan should also say which near-misses are intentionally rejected rather than guessed.

4. **Specify free-form arithmetic rejection.**
   The plan should call out that strings containing operators or expression syntax such as `ctl/atl`, `ctl - atl`, `tss_per_hour`, `power:weight`, `np/ftp`, `weekly_tss/weekly_hours`, parentheses, or formulas are invalid even if their component fields are known. This is central to the task and should be tested in Step 3.

5. **Document the hint strategy for unknown metrics.**
   Step 1 should define short, deterministic hint categories, for example:

   - Best-effort duration/distance requests: `try analyze_efforts_delta for best-effort durations/distances`.
   - Zone distribution/load-balance requests: `try compute_zone_time` or `compute_load_balance`.
   - Segment/stream statistics: `try compute_activity_segment_stats`.
   - Compliance/adherence: `try compute_compliance_rate` if not represented as a scalar metric.
   - General unknown names: include a concise supported-metrics sample or nearest safe canonical names.

   The hint text should be one line, actionable, and avoid internal implementation details.

6. **Decide package/API shape before Step 2.**
   Even though implementation is Step 2, the design should state the target boundary: preferably an SDK-free reusable package such as `internal/analysis` with a typed `Metric` enum, `ParseMetric`, `Metric.Values()`/schema-enum helper, alias metadata, and invalid-argument error/hint helpers usable by future analyzer tools. This avoids coupling the enum to one tool or to MCP SDK details.

7. **Call out schema and docs implications.**
   The plan should require schema descriptions/enums to enumerate canonical metric values and should note that `CHANGELOG.md` is updated when behavior lands. If generated tool docs are not affected in this task because analyzers are not registered yet, record that explicitly.

## What looks good

- The task scope is appropriately narrow: a reusable enum/validation contract, not analyzer implementation.
- The prompt correctly forbids arbitrary math expressions and points to existing read-tool fields as the source of truth.
- `internal/analysis` is a reasonable new package because the contract will be shared by multiple future analyzer tools.

## Recommendation

Update `STATUS.md` with the concrete metric inventory, canonical names, aliases, rejected expression examples, and hint categories before starting Step 2. Once those design decisions are recorded, the implementation should be straightforward and testable.
