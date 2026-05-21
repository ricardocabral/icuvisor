# Plan Review — TP-085 Step 2

**Verdict:** APPROVE

The Step 2 plan is sufficient to implement the user-facing wording change. It incorporates the Step 1 audit findings and addresses the main risk from the prior review: inconsistent Strava-unavailable text across different activity read paths.

## What is covered

- The plan updates all identified Strava/import-blocked construction paths, including the non-obvious paths outside the original prompt scope:
  - list/detail rows via `activityRow`
  - messages via `stravaUnavailableMessagesResponse`
  - direct interval DTO empty/stub path via `stravaUnavailableIntervalsResponse`
  - streams/splits fallback via `detectActivityUnavailable`
  - extended metrics via `stravaUnavailableExtendedMetricsResponse`
- It uses a shared workaround builder, which should prevent text drift between tools.
- It keeps provider inference conservative by limiting provider-aware wording to allowlisted native providers from explicit raw payload evidence.
- It preserves the stable structured `unavailable.reason` values (`strava_tos` and `strava_blocked`) while changing only the actionable `workaround` text.
- It explicitly includes both provider-aware and provider-neutral fallback wording, matching the mission requirements.

## Implementation notes

- Prefer a helper that can be called from both typed `intervals.Activity` paths and raw `map[string]any` paths so `activityRow`, direct interval DTO fallback, messages, fallback helper, and extended metrics all share exactly the same wording logic.
- Do not infer a provider from `source: Strava`, `_note`, generic sync-chain markers, or unallowlisted prefixes such as MyWhoosh/TrainerRoad. Unknown cases should always use the provider-neutral text.
- Keep the exact Step 1 strings centralized; Step 3 can then assert exact output without duplicating independent wording decisions.
- Ensure `include_full` behavior remains unchanged: full raw payloads are still only returned when requested.

No plan changes are required before implementing Step 2.
