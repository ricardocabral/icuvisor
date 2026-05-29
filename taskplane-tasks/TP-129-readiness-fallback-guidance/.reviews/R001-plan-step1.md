# Plan Review: Step 1 — Audit wellness readiness semantics

Verdict: Approved

The Step 1 plan is appropriate for the task. It covers the key audit surfaces needed before changing prompts/docs: wellness shaping, provenance metadata, `_native` provider sidecars, recovery/weekly prompt guidance, null-field metadata, discovery logging, and targeted tests.

I also verified the targeted test command currently passes:

```sh
go test ./internal/tools ./internal/prompts
```

Notes for execution:
- When recording discoveries, explicitly call out that terse shaping records null upstream fields in `_meta.missing_fields`; confirm whether `readiness` is included when upstream sends `readiness: null`.
- Include fallback evidence fields in the discovery notes: HRV/rMSSD, HRV SDNN, resting HR, sleep seconds/quality/score, avg sleeping HR, fatigue/soreness/stress/feel/mood/motivation, SpO2/respiration/steps/VO2max/Baevsky where present, and provider `_native` fields such as Garmin Body Battery.
- Keep the non-goal explicit: do not synthesize or backfill an Intervals readiness score from native or supporting fields.
