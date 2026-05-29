# Plan Review R001 — Step 1

**Verdict:** Approved with minor additions before completion.

The Step 1 plan matches the task: it audits the existing `compute_activity_segment_stats` tool, existing tests/evals, records whether a higher-level helper is warranted, and runs the targeted `go test ./internal/tools` gate. I ran the targeted test command and it passes (`ok`, cached).

## Additions to include in the Step 1 audit

Before marking Step 1 complete, please explicitly record these points in `STATUS.md` discoveries/disposition:

1. **“Last 10 km” requires bound translation.** The tool accepts explicit `start_distance_m`/`end_distance_m`; it does not accept a relative “last N km” request. The audit should confirm what upstream call provides total activity distance so the workflow can compute `total_distance_m - 10000` to `total_distance_m`, or conclude a helper is warranted.
2. **Description/schema mismatch.** The current tool description mentions “maximum” and “zone-time”, but the stat enum is `mean`, `median`, `p90`, `decoupling`, `drift`, `np`, `if`. This is relevant to activation hardening and should be dispositioned for Step 2.
3. **Pace wording.** The schema exposes `velocity_smooth` with velocity units, not a pace-formatted metric. Since the task mentions pace comparisons, record whether final-answer conversion is sufficient or whether tool/prompt text needs tightening.
4. **Terse/full coverage.** Existing tests cover terse omission of `series` for one scalar distance case and full audit for decoupling, but first-vs-last distance segment coverage appears to belong in Step 3. Record that gap explicitly.

No blocker to proceeding with Step 1 as planned.
