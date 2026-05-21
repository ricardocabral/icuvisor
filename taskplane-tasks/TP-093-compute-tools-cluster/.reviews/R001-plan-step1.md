# Plan Review — TP-093 Step 1: Design deterministic contracts

**Verdict:** Changes requested

The current Step 1 plan in `STATUS.md` is only the high-level task checklist. It does not yet define the deterministic contracts needed before implementation, so it is not reviewable/approvable as a Step 1 design.

## Blocking findings

1. **No concrete request schemas are specified.**
   Step 1 must pin the public inputs for all four tools, including exact field names, required vs optional arguments, date-window semantics, sport/category filters, zone metric enum (`power` / `heart_rate` / `pace` or equivalent), baseline/current window shape, limits, and `include_full` behavior. Without this, Step 2 can accidentally invent incompatible schemas.

2. **Source-tool priority is not defined.**
   The task explicitly requires deciding source priority before any fallback. The plan needs a per-tool ordered source list, especially for `compute_zone_time` and `compute_load_balance`, e.g. which precomputed fields from `get_training_summary`, `get_extended_metrics`, and/or `get_activity_intervals` are authoritative, and whether `get_activity_streams` is prohibited or only a reported unavailable fallback. This matters because the PRD says analyzer tools should not recompute from streams when upstream zone time exists, and `compute_activity_segment_stats` is the raw-stream exception.

3. **Analyzer `_meta` and formula refs are not mapped per tool.**
   Step 1 should specify each tool's `_meta.method`, `_meta.source_tools`, `_meta.n`, `_meta.missing_days`, `_meta.missing_action`, `_meta.insufficient_sample`, assumptions/boundaries, and formula ref. At minimum, baseline should reference `icuvisor://analysis-formulas#z_score`, and zone/load-balance outputs using polarization should reference `#polarization_index`. If `compute_compliance_rate` has no canonical formula ref, the plan should say so explicitly and define the method string instead.

4. **Missing/insufficient-data behavior is not contracted.**
   The mission requires explicit missing/insufficient signals and no silent imputation. The plan should define how missing days, activities without precomputed zone arrays, events without targets, zero baseline variance, and insufficient baseline samples are represented in both result and `_meta`.

5. **`compute_compliance_rate` pairing semantics are undefined.**
   The plan must define what counts as scheduled vs completed, which event fields provide targets/actuals, how `link_activity_to_event` pairings are detected/reused, how manual vs auto-pairing is surfaced, and how TP-090 auto-lap caution propagates when interval-consuming compliance behavior is involved.

6. **Response shapes are not pinned.**
   The design should state terse vs `include_full:true` payload contents for each tool, including aggregate rows, audit series, missing-source summaries, units, classification labels (`polarized` / `pyramidal` / `threshold`), and stable `_meta` contents. These are part of the public MCP contract and need to be decided before implementation.

## Recommended Step 1 deliverable

Add a concise contract section to `STATUS.md` or a scoped design note under the task folder that includes, for each tool:

- request schema fields and enum values;
- ordered source-tool/data-field priority;
- deterministic calculation method and formula ref, if any;
- minimum sample and missing-data rules;
- terse and full response shape;
- user-facing unavailable/insufficient states;
- tests implied by the contract.

Once that exists, the implementation plan can be reviewed against the PRD/ROADMAP constraints before code is written.
