# Upstream gap: precomputed per-activity zone-time coverage

## Summary

TP-099 audited the repository's v0.2 fixture corpus on 2026-05-20 to determine whether `compute_zone_time` and `compute_load_balance` can rely on upstream precomputed zone-time fields for power, heart-rate, and pace. The current analyzer implementation is intentionally precomputed-only: it does not reduce raw streams for these tools, and it reports missing/partial coverage when upstream zone-time fields are absent.

The fixture audit found no valid precomputed zone-time arrays in the eligible activity-like fixtures. The fixtures are also too sparse to classify deterministic stream-math fallback opportunities without overstating missing coverage, so every metric-family opportunity in the current fixture set is marked `unknown`. This is an upstream/fixture coverage risk: the analyzer contract depends on fields that are not represented in the checked-in v0.2 fixtures.

## Reproduction

Run from the repository root:

```sh
go run scripts/audit_zone_time_coverage.go
```

The script scans only these task-scoped fixture roots:

- `internal/intervals/testdata/**/*.json`
- `internal/tools/testdata/**/*.json`

It applies path/type exclusions before object-shape detection so wellness, events, gear, workout-library, custom-item, activity-message, activity-interval, analyzer-golden, schema-snapshot, and non-JSON fixtures cannot be counted as activity-zone coverage opportunities.

## Precomputed field criteria

A field counts as precomputed zone time only when it is a non-empty numeric array with positive total seconds.

Training-summary rows count as power coverage when both of these are present and positive/useful:

- `timeInZones`
- `timeInZonesTot > 0`

Activity or extended-metrics rows count by metric family using the keys currently recognized by `compute_zone_time` and `compute_load_balance`:

| Metric family | Precomputed fields |
| --- | --- |
| power | `icu_zone_times`, `power_zone_distribution_seconds`, `power_zone_times` |
| heart_rate | `hr_zone_times`, `heartrate_zone_times`, `heart_rate_zone_times`, `hr_time_in_zones` |
| pace | `gap_zone_times`, `pace_zone_times`, `pace_zone_time_seconds` |

## Audit result

Command output summary:

| Metric family | Precomputed count | Fallback count | Unknown count | Known coverage |
| --- | ---: | ---: | ---: | ---: |
| power | 0 | 0 | 6 | 0.0% |
| heart_rate | 0 | 0 | 6 | 0.0% |
| pace | 0 | 0 | 6 | 0.0% |

Eligible fixture units:

| Path | Eligible objects | Result |
| --- | ---: | --- |
| `internal/intervals/testdata/activities/strava_sync_chain_empty_stubs.json` | 3 | Sparse Strava-sync empty stubs; no precomputed zone arrays or family-specific signals. |
| `internal/intervals/testdata/activity_detail_with_gear.json` | 1 | Gear/detail fixture; no precomputed zone arrays or family-specific signals. |
| `internal/intervals/testdata/activity_list_with_gear.json` | 2 | Gear/list fixture; no precomputed zone arrays or family-specific signals. |

Skipped excluded or non-eligible objects/files: 36.

## Interpretation

No agreed fallback threshold is documented in `ROADMAP.md` or `docs/prd/PRD-icuvisor.md`; threshold selection remains an operator decision. For this audit, any missing-precomputed opportunity would be risky evidence rather than an automatic pass/fail threshold.

The measured result is more conservative: the checked-in v0.2 fixtures do not prove either adequate upstream precomputed coverage or actual fallback frequency. They prove that the repository lacks fixture coverage for the precomputed zone fields the analyzers require. Until live/dogfood or upstream-provided fixtures include those fields, `compute_zone_time` and `compute_load_balance` should keep their explicit missing-precomputed behavior and should not silently switch to LLM-side or raw-stream reduction.

## Requested upstream/API evidence

If intervals.icu exposes per-activity or summary precomputed zone-time fields, icuvisor needs stable public documentation and representative responses for:

- power zone seconds per activity and/or day;
- heart-rate zone seconds per activity and/or day;
- pace/GAP zone seconds per activity and/or day;
- field units and zone ordering;
- whether zero-filled arrays mean no data, no time in zone, or unavailable source data.

## Feature request status

Status: not filed — maintainer-authenticated forum or support filing required.

Public feature-request URL: pending maintainer-authenticated filing.

Target public channel URL: `https://forum.intervals.icu/`

### Copy-paste-ready feature-request draft

Title: Document and expose precomputed activity zone-time fields in the public API

```text
Hi intervals.icu team,

Could the public API document and consistently expose precomputed time-in-zone arrays for activities and/or daily summaries?

Use case: integrations such as icuvisor's local MCP analyzers need to answer questions like "how much time did I spend in power/HR/pace zones this block?" and classify low/moderate/high load balance. The safe implementation is to aggregate upstream-computed zone seconds, because intervals.icu already owns the athlete's configured zones, sport context, and GAP/pace semantics. Recomputing from raw streams in each integration risks definition drift, higher API volume, and inconsistent answers.

Requested coverage:

- power zone seconds per activity and/or summary day;
- heart-rate zone seconds per activity and/or summary day;
- pace/GAP zone seconds per activity and/or summary day;
- documented field names, units, zone ordering, and zero/null semantics;
- representative examples in the public API docs.

icuvisor currently looks for these observed/expected field names when present:

- power: `icu_zone_times`, `power_zone_distribution_seconds`, `power_zone_times`, and summary `timeInZones` with `timeInZonesTot`;
- heart rate: `hr_zone_times`, `heartrate_zone_times`, `heart_rate_zone_times`, `hr_time_in_zones`;
- pace: `gap_zone_times`, `pace_zone_times`, `pace_zone_time_seconds`.

If a different canonical schema exists, documenting that schema would let clients rely on it and avoid raw stream reduction.

Thanks!
```

Once filed, replace the placeholder above with the forum/support URL.
