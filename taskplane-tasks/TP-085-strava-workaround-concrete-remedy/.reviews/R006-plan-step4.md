# Plan Review — TP-085 Step 4

**Verdict:** APPROVE

The Step 4 plan is sufficient for the remaining docs and verification work. It targets the two required user-facing documentation surfaces and includes the full quality gate required by the task completion criteria.

## What is covered

- Troubleshooting docs will be updated with the same concrete intervals.icu Connections-page remedy used in tool responses.
- `CHANGELOG.md` will record the user-visible behavior change under `[Unreleased]`.
- The full quality gate is included before delivery, which should catch regressions outside the targeted Step 3 test run.

## Implementation notes

- Add the troubleshooting guidance in `web/content/guides/troubleshooting.md` as a clear symptom/fix row for Strava-imported activities whose streams/splits/messages/metrics are unavailable. Use the exact provider-neutral remedy from STATUS for the general docs case, and only include provider-specific wording as an example if it stays identical to the code wording pattern.
- Put the changelog entry under `[Unreleased]` -> `Changed` because this changes existing user-facing response text rather than adding a new feature.
- Treat “full quality gate” as at least:
  - `make test`
  - `make build`
  - `make lint`
- If a gate fails, fix it before marking the step complete, or document a clearly unrelated pre-existing failure in `STATUS.md` as required by the task prompt.
- Since Step 5 duplicates the quality-gate checklist, record the actual command results in `STATUS.md` so the later verification step can be checked off without ambiguity.

No plan changes are required before implementing Step 4.
