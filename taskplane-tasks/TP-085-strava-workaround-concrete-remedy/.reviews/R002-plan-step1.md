# Plan Review — TP-085 Step 1

**Verdict:** APPROVE

The revised Step 1 plan addresses the gaps called out in R001 and is sufficient for the audit phase before implementation.

## What is covered

- It expands the audit beyond the original prompt file scope by explicitly searching for `stravaWorkaround`, `strava_tos`, `strava_blocked`, and `StravaImported`.
- It requires the worker to identify all Strava/import-blocked unavailable marker construction paths, including paths discovered outside the initial file list or to document intentional exclusions.
- It adds a provider-inference decision point and correctly calls out the need for safe rules that avoid inferring a native provider from ambiguous Strava/sync-chain evidence.
- It makes the exact provider-aware and provider-unknown workaround strings a Step 1 deliverable, which should keep Step 2 implementation, tests, and docs aligned.
- It requires Step 1 acceptance notes in `STATUS.md` covering constructor list/reason codes, provider inference availability, exact strings, and file-scope expansion.

## Notes for execution

When completing Step 1, make sure the discovery notes include every constructor found by the search, especially the known non-obvious paths from R001 such as `get_activity_messages.go` and `get_extended_metrics.go` if they remain relevant. The notes should also distinguish payloads that contain raw activity data from paths that only receive an activity ID or already-shaped unavailable data, because that determines whether provider-aware wording can be used safely.

No further plan changes are required before proceeding with the Step 1 audit.
