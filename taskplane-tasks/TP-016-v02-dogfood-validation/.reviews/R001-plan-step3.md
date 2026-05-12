# R001 plan review — Step 3: Invited athletes (2–3) read-only

Verdict: **APPROVE**

The revised Step 3 plan is now reviewable and safe to execute. The invited-athlete section in `docs/dogfood/v0.2-findings.md` covers the privacy-sensitive parts that were missing from the prior review: anonymous participant labels, inclusion targets, local-only account execution, explicit forbidden data, consent language, client setup paths, fresh-session requirements, blocked-prompt handling, local response-size measurement, artifact cleanup, and a clear boundary that triage/fixes wait for Step 4.

## What looks good

- The participant protocol avoids collecting API keys, athlete IDs, screenshots with values, raw transcripts, raw tool payloads, exact dates, exact load/threshold values, activity names, and private comments/messages.
- Inclusion targets match the task: 2–3 invited athletes, with best-effort coverage for an imperial/miles account and a non-Garmin wellness/provenance source.
- Run instructions point to the expected manual setup docs and the canonical `docs/dogfood/v0.2-prompts.md` prompt set.
- The plan requires fresh Claude/Codex sessions after configuration or rebuilds, reducing schema-cache contamination risk.
- The redacted findings template captures the right validation signals without requesting raw athlete data.
- Coverage/fallback rules are explicit, including `blocked` for missing account data and specific follow-up focus from the solo run: activity interval/stream/split/extended-metrics failures, `get_workouts_in_folder` verbosity, and dual sleep-scale coverage.
- Step 4 boundaries are respected: record failures now, defer issue classification, launch-blocking calls, and fixes.

## Non-blocking suggestions for execution

- When inviting participants, mention the exact git revision or release artifact they should build/run so results are comparable across accounts.
- The referenced client docs are still labeled v0.1 in places; participants should treat them as setup recipes and use the current TP-016 prompt set/tool catalog for validation.
- For participants who cannot measure response bytes, keep the optional size column blank rather than asking for logs or transcripts.

Proceed with recruitment and collection under the documented protocol.
