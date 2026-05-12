# Code Review: TP-011 Step 1 — Map the wellness payload

**Verdict: Approve.**

The revised Step 1 mapping addresses the prior blocking issues: `feel` is now included in the subjective-scale inventory with the OpenAPI gap documented, and the stale bridge wording now consistently treats `fetched_at` older than the wellness row/reference time by `>24h` as stale while keeping exactly `24h` fresh.

## Findings

No blocking findings.

## Non-blocking notes

- The OpenAPI static `Wellness` field list in `STATUS.md:120-166` matches the current public intervals.icu schema I checked via `https://intervals.icu/api/v1/docs`.
- The native provider field names for Garmin/Oura remain explicitly marked as fixture/probe assumptions (`STATUS.md:211-212`), which is appropriate for Step 1 given live authenticated probing was unavailable.
- Later implementation steps should keep the `updated` timestamp caveat intact (`STATUS.md:213`) and avoid using row `updated` as provider `fetched_at` unless a fixture/probe provides bridge-specific evidence.

## What I checked

- Ran `git diff 8943fa7df40229fa9b131ff98c880b8453b827ca..HEAD --name-only` and the full diff.
- Read `PROMPT.md` and the changed `STATUS.md`.
- Cross-checked the current public intervals.icu OpenAPI `Wellness` schema field list.
