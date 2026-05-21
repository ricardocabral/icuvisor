# R001 plan review — Step 3: Invited athletes (2–3) read-only

Verdict: **APPROVE**

The Step 3 plan is safe and sufficiently specific for invited-athlete read-only validation. The latest hardening addresses the prior comparability gap by requiring each invite to name the exact release artifact or git revision participants should run.

## Findings

No blocking findings.

## What looks good

- The plan keeps external athlete data local and explicitly forbids collecting API keys, athlete IDs, raw MCP/client logs, transcripts, screenshots with values, exact dates, exact load/threshold values, activity names, and private notes/messages.
- Inclusion targets match the task: 2–3 invited athletes, with best-effort coverage for an imperial/miles user and a non-Garmin wellness/provenance source.
- Participants are directed to run the same canonical `docs/dogfood/v0.2-prompts.md` set with allowed clients only, fresh sessions, and comparable binary/source revisions.
- The redacted findings template captures pass/fail/blocked, tool names, broad units/provenance, qualitative scale/unit/Strava observations, and optional local response-size maxima without requesting raw data.
- Blocked-prompt handling is appropriate for account-specific missing data, and the plan focuses Step 3 on evidence collection while leaving issue classification, launch-blocking calls, and fixes to Step 4.
- The current blocked status is valid: remaining recruitment, participant runs, and result collection require maintainer-supplied external participants/results and should not be fabricated or attempted by the worker via unauthorized outreach.

Proceed when the maintainer supplies participants or redacted invited-athlete results under the documented protocol.
