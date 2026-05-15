# Upstream gap: athlete-level periodization parameters

## Summary

TP-014 probed intervals.icu's public OpenAPI docs for athlete-profile and training-plan endpoints on 2026-05-11. The API documents training-plan assignment metadata, plan folders, planned events, workout targets, and fitness/wellness metrics, but it does not expose athlete-level planning parameters for:

- ramp-rate percentage;
- recovery-week cadence;
- taper percentage drop;
- intensity-distribution preference.

icuvisor should not compute these values client-side. If intervals.icu exposes explicit fields later, `get_planning_parameters` can ship with only the fields confirmed available, following the PRD availability rule that missing upstream fields are omitted rather than zero-filled.

## Evidence summary

Public documentation sources checked:

- `https://intervals.icu/api-docs.html`
- `https://intervals.icu/api/v1/docs`

Checked endpoint/schema areas:

| Area | Documented fields relevant to planning | Gap verdict |
| --- | --- | --- |
| `GET /api/v1/athlete/{id}` | `training_plan_id`, `training_plan_start_date`, device workout-upload range preferences | No athlete-level periodization parameters. |
| `GET /api/v1/athlete/{id}/profile` | `AthleteProfile.athlete` wrapper around the athlete schema | No athlete-level periodization parameters. |
| `GET /api/v1/athlete/{id}/training-plan` | `athlete_id`, `training_plan_id`, `training_plan_start_date`, `timezone`, `training_plan_last_applied`, `training_plan`, `training_plan_alias` | Training-plan assignment only; no planning-parameter fields. |
| `PUT /api/v1/athlete-plans`, `PUT /api/v1/athlete/{id}/training-plan` | `id`, `training_plan_id`, `training_plan_start_date`, `training_plan_alias` | No writable planning-parameter hints; write probing was out of scope. |
| `GET /api/v1/athlete/{id}/folders` | `rollout_weeks`, `duration_weeks`, `hours_per_week_min`, `hours_per_week_max`, `workout_targets`, `starting_ctl`, `starting_atl` | Plan metadata only; no ramp, recovery cadence, taper, or distribution preference. |
| `GET /api/v1/athlete/{id}/events{format}` | `icu_training_load`, `load_target`, `time_target`, `distance_target`, `target`, `icu_intensity`, plan linkage fields | Per-event values only, not athlete-level periodization settings. |
| `GET /api/v1/athlete/{id}/fitness-model-events` | Fitness-model event categories such as `FITNESS_DAYS`, `SET_FITNESS`, `SET_EFTP` | No planning-parameter fields. |
| `GET /api/v1/athlete/{id}/workouts`, `GET /api/v1/athlete/{id}/workouts/{workoutId}` | `icu_training_load`, `target`, `targets`, `plan_applied`, `for_week`, `icu_intensity` | Per-workout values only, not an athlete-level intensity-distribution preference. |

Near-match fields found elsewhere:

- `Wellness.rampRate`
- `SummaryWithCats.rampRate`

Those near-matches are observed fitness/wellness metrics, not configurable athlete- or coach-level periodization parameters.

No authenticated live probe was run because this worktree did not contain a `.env` file. No raw API responses or athlete identifiers were committed.

## Requested fields and use cases

### Ramp-rate percentage

A coach or self-trained athlete may want the API to expose the configured weekly load progression, such as a 5-8% weekly ramp. Forum post #28 in thread 123739 describes this as an explicit setting coaches may want to review or adjust: "Weekly progression: Just a standard exponential ramp (around 5-8% a week)" and later calls ramp rates the kind of explicit setting a coach could set differently.

Requested upstream shape: an explicit numeric percentage field, with units and scope documented, for example `ramp_rate_percent` on an athlete planning-parameters resource.

### Recovery-week cadence

A plan generator or coach workflow needs to know how often lower-load recovery weeks are scheduled, instead of inferring cadence from generated workouts. Forum post #28 references "baked-in recovery weeks" as part of the load-planning logic, and the requested API field would make that cadence inspectable.

Requested upstream shape: an explicit cadence field such as `recovery_week_cadence` with a documented unit, for example every N weeks or an enumerated strategy.

### Taper percentage drop

Race-week and event-specific planning need to distinguish normal training-load changes from an intentional taper. Forum post #28 describes tapering as "A progressive ~40% volume drop based on Bosquet," while forum post #30 discusses using ATL and CTL to manage the taper. An explicit taper setting would let icuvisor report the configured taper policy without inventing it from calendar data.

Requested upstream shape: an explicit `taper_drop_percent` or equivalent field, with the taper window/scope documented.

### Intensity-distribution preference

An LLM planning assistant should not infer whether an athlete or coach prefers polarized, pyramidal, or another distribution from workouts after the fact. Forum post #28 describes "Either Seiler's polarized model (80/20) or pyramidal" depending on context. Exposing this preference would let icuvisor report the selected strategy and its scale without reverse-engineering planned workouts.

Requested upstream shape: an enum such as `intensity_distribution_preference` with documented values, for example `polarized`, `pyramidal`, or `custom`, plus optional distribution percentages if upstream stores them.

## Feature request status

Status: not filed — maintainer-authenticated forum filing required.

Public feature-request URL: pending maintainer-authenticated forum filing.

Target public channel URL: `https://forum.intervals.icu/`

Suggested context links for the maintainer to include:

- `https://forum.intervals.icu/t/any-reviews-or-comparisons-of-the-wealth-of-ai-tools-for-intervals/123739/28`
- `https://forum.intervals.icu/t/any-reviews-or-comparisons-of-the-wealth-of-ai-tools-for-intervals/123739/30`

On 2026-05-11, TP-014 attempted to post the request to the public intervals.icu forum endpoint, but the anonymous request was rejected with `403` / `BAD CSRF`; an authenticated forum session or maintainer action is required. Per maintainer steering, no further anonymous or authenticated POSTs or external account actions should be attempted by task workers. Until that external action happens, the URL placeholder above is the canonical link target for TP-014.

### Copy-paste-ready feature-request draft

Title: Expose athlete-level periodization planning parameters in the public API

```text
Hi intervals.icu team,

Could the public API expose athlete-level periodization/planning parameters when those settings are stored upstream?

Requested fields:

- ramp-rate percentage, e.g. the configured weekly load progression;
- recovery-week cadence, e.g. every N weeks or the configured recovery-week strategy;
- taper percentage drop and taper-window scope;
- intensity-distribution preference, e.g. polarized, pyramidal, custom, and optional distribution percentages if stored.

Use case: integrations such as local MCP/LLM assistants need to report an athlete or coach's configured planning assumptions without deriving them from activities, planned workouts, or generated calendar data. Client-side derivation would risk misrepresenting coach intent. Per-event fields such as `icu_intensity`, workout targets, folder metadata, and observed metric fields such as `rampRate` are useful, but they do not identify the athlete-level planning preferences themselves.

Evidence checked in the current public OpenAPI docs:

- `GET /api/v1/athlete/{id}` exposes training-plan assignment metadata, but no periodization parameters.
- `GET /api/v1/athlete/{id}/profile` wraps the athlete schema and does not expose these settings.
- `GET /api/v1/athlete/{id}/training-plan` exposes assignment metadata such as `training_plan_id`, `training_plan_start_date`, timezone, last-applied state, and aliases, but no ramp/recovery/taper/distribution settings.
- Folder, event, fitness-model-event, and workout endpoints expose plan metadata or per-event/per-workout targets, not athlete-level periodization preferences.
- `Wellness.rampRate` and `SummaryWithCats.rampRate` appear to be observed metrics rather than configurable planning assumptions.

A small read-only resource such as `GET /api/v1/athlete/{id}/planning-parameters` would be enough for this use case, but any documented schema/path is fine. Missing values could simply be omitted or returned as null according to intervals.icu's existing API conventions.

Thanks!
```

Once filed, replace the placeholder above with the forum/support URL and update the TP-014 status file.

## References

- Forum thread 123739, post #28: `https://forum.intervals.icu/t/any-reviews-or-comparisons-of-the-wealth-of-ai-tools-for-intervals/123739/28`
- Forum thread 123739, post #30: `https://forum.intervals.icu/t/any-reviews-or-comparisons-of-the-wealth-of-ai-tools-for-intervals/123739/30`
