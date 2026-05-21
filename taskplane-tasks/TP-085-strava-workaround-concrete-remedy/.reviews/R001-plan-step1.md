# Plan Review — TP-085 Step 1

**Verdict:** Changes requested

The Step 1 intent is correct, but the current plan is too sparse for a wording change that must be consistent across all Strava-unavailable responses. Please tighten the audit plan before implementing Step 2.

## Required plan adjustments

1. **Audit all current construction paths, not only the prompt file scope.**
   A repository search shows additional Strava unavailable marker constructors outside the listed file scope:
   - `internal/tools/get_activities_row.go` — list/details rows via `activityRow`, reason `strava_tos`
   - `internal/tools/activity_unavailable.go` — fallback helper used by intervals/streams/splits, reason `strava_blocked`
   - `internal/tools/get_activity_details.go` — direct interval DTO empty/stub path via `stravaUnavailableIntervalsResponse`
   - `internal/tools/get_activity_streams.go` — streams/splits through the shared fallback helper
   - `internal/tools/get_activity_messages.go` — `stravaUnavailableMessagesResponse`, reason `strava_tos`
   - `internal/tools/get_extended_metrics.go` — `stravaUnavailableExtendedMetricsResponse`, reason `strava_blocked`

   Step 1 should explicitly state that the audit is driven by `stravaWorkaround`, `strava_tos`, `strava_blocked`, and `StravaImported` searches, and should either expand the file scope or log why any path is intentionally excluded. Otherwise Step 2 can easily leave inconsistent user-facing text.

2. **Define safe provider inference rules before writing code.**
   The plan should distinguish reliable native-provider evidence from ambiguous Strava/sync-chain evidence. Current fixtures include `external_id` prefixes like `wahoo-synthetic-*`, `mywhoosh-synthetic-*`, and `trainerroad-synthetic-*`; only some of these are obviously native device providers. Avoid wording that implies a native provider was identified from `source: Strava`, `_note`, or opaque IDs unless the audit documents the evidence and an allowlist/normalization rule.

3. **Record exact target workaround strings as the Step 1 deliverable.**
   Step 1 should finish with exact provider-aware and provider-unknown strings that Step 2 and docs/tests will reuse. The strings need to include the concrete action from the task: intervals.icu Connections page → provider → **Download old data**, and must not suggest bypassing Strava restrictions.

## Suggested Step 1 acceptance notes

Before marking Step 1 complete, update `STATUS.md` discoveries with:

- The complete list of unavailable-marker constructors and reason codes found.
- Whether each path can receive raw activity payloads sufficient for provider inference.
- The exact provider-aware and unknown-provider workaround text.
- Any file-scope expansion needed for Step 2/3, especially `get_activity_messages.go` and `get_extended_metrics.go` if they remain in scope for consistent behavior.

No security concerns found in the plan; the main risk is incomplete/inconsistent user-facing behavior.
