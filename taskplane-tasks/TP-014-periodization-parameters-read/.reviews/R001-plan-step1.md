# Review R001 — Plan review for Step 1

## Verdict: approved

The revised Step 1 plan is sufficient to proceed. It addresses the prior plan-review blockers by documenting the PRD/ROADMAP mismatch, making the probe endpoint matrix explicit, requiring evidence that supports negative findings, and adding credential/redaction constraints appropriate for live black-box probing.

## What looks good

- The plan keeps the task scoped to upstream availability and does not propose deriving periodization parameters client-side.
- The PRD/ROADMAP mismatch is captured in `STATUS.md`, with a clear decision to proceed under the task prompt unless maintainer guidance changes.
- The probe matrix covers the relevant documented athlete/profile and training-plan surfaces, including read-only adjacent endpoints that may expose plan metadata or planned-load hints.
- Write-capable endpoints are marked docs-only, which avoids accidental mutation during a read-path discovery task.
- The availability table is structured around the required yes/no/partial verdicts and now calls for endpoints checked, JSON paths or explicit absence, evidence source, and near-match notes.
- Redaction guidance is adequate: record endpoint paths with `{athlete_id}` placeholders and field/key availability only, not raw values or raw responses.

## Non-blocking recommendations while executing

- Treat the current probe matrix as a minimum: if the public OpenAPI docs reveal additional athlete-profile or training-plan variants, add them to `STATUS.md` before recording the final verdict.
- For docs and live JSON, inspect raw keys/paths rather than existing Go structs or typed fixtures, since local structs can hide upstream fields.
- Use broad near-match searches for names around load progression, ramp/rate, recovery/rest weeks, taper/deload, and intensity distribution/polarization, then explain why each near-match does or does not satisfy the requested athlete-level parameter.
- If a field appears only as an event/workout consequence or derived projection rather than an athlete-level planning parameter/preference, classify it as `partial` or `no` with that caveat rather than treating it as directly exposed.
- Keep any downloaded specs or live response samples outside the repo, and discard or redact them after extracting field/path availability.

No plan changes are required before starting the probe.
