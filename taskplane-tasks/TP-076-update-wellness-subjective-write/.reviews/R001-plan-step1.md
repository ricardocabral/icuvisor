# Plan Review: Step 1 live probe

Result: changes requested before running the probe.

The Step 1 plan has the right overall goal (direct live probe, isolate field combinations, capture sanitized fixtures, clean up), but it needs a few safety and fidelity corrections to avoid diagnosing the wrong contract or leaving persistent test-athlete mutations.

## Findings

### 1. Probe must start from the production client shape, not a generic POST

`internal/intervals/wellness.go` currently writes with:

- `PUT`
- path: `/athlete/{athleteID}/wellness/{date}`
- body: sparse field payload only, with no `date` / `start_date_local` field

The plan says to do direct `POST` probes and vary `start_date_local`. That may be useful as a later alternative, but the first baseline probe must exactly mirror the shipped client request. Otherwise the probe can produce a false root cause by testing a different endpoint/method/body shape than `update_wellness` uses.

Required plan amendment:

1. Capture/log a redacted baseline for the current client shape: `PUT /athlete/<redacted>/wellness/YYYY-MM-DD` with `{ "feel": 3 }`.
2. Only if that fails, branch into endpoint/method/date-body variants (`POST`, `start_date_local`, athlete/date in body vs path).
3. Save fixtures for the shape that will actually be implemented in `writeWellnessBody` / `UpdateWellness`.

### 2. Add an explicit pre-state snapshot and restore strategy

The plan says to “clear any unintended subjective values,” but clearing is not equivalent to cleanup if the row already had real values. Since this is a live account, the probe should snapshot the target row before any write and restore it exactly after probing, including `locked`.

Required plan amendment:

- Before the first write, `GET` the target date and save the pre-probe values out of repo/scratch only.
- Prefer a known disposable date/row with no existing wellness data. If using an existing row, restore each touched field to its original value rather than clearing it.
- Probe `locked: true` last and immediately unlock/restore after verifying, because a locked row can interfere with future sync/probes.

### 3. Add a production-account guard before sourcing/running `.env-dev`

Because Step 1 performs live writes, the plan should explicitly verify that the loaded athlete/key are the dedicated test athlete and not the maintainer’s primary account. This was done in dogfood, but should be repeated for this task run.

Required plan amendment:

- Source `.env-dev` without echoing secrets.
- Verify required env vars are present without printing API keys or raw athlete IDs.
- Compare the target athlete against any production/default env target if available, and abort if they match.
- Redact athlete ID/date values in committed notes and fixtures per the task’s “Do NOT” section.

## Suggested revised Step 1 outline

1. Source `.env-dev` safely and run the production-account guard.
2. Choose a disposable probe date; GET/snapshot the row before mutation.
3. Baseline probe using current client shape: `PUT /athlete/{id}/wellness/{YYYY-MM-DD}` with one subjective field.
4. Bisect fields using the same method/path/body shape; re-read after each accepted write.
5. If all singles pass but the bundle fails, bisect combinations, with `locked` tested last.
6. If current shape fails before field-specific evidence, then test alternate date/method/scoping variants.
7. Record exact accepted and rejected payload/response pairs as sanitized fixtures.
8. Restore the pre-probe row, ensure `locked: false` unless it was originally true, and verify with `get_wellness_data`/direct GET.

With these amendments, the Step 1 plan should be safe and targeted enough to isolate the rejection cause.
