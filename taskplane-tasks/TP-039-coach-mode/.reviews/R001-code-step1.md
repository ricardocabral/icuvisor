# Code Review — TP-039 Step 1

## Findings

### [P1] Do not mark the coach-roster endpoint probe complete without an authenticated coach-key probe

`docs/threat-models/coach-mode.md` explicitly says the probe was limited to public OpenAPI docs and unauthenticated HTTP responses, and that no real coach key was available (lines 86-88). However `STATUS.md` marks “Coach-roster endpoint probed” complete (lines 17-19) and the doc turns that absence of credentials into a v0.5 implementation decision to use config-only roster data (lines 119-123).

The Step 1 prompt required a black-box probe “with a real coach key” and, only if the endpoint is not exposed for free-tier coach accounts, documenting that gap. A missing local credential is not the same as confirming the endpoint is unavailable for a coach/free-tier account, and it leaves the path/auth/response/pagination behavior unvalidated. Downstream Step 4 could now implement `list_athletes` from config only based on an incomplete probe rather than an authenticated result.

Please either:

- run the authenticated probe with a real coach-scoped key and document status, auth header, response shape, and pagination behavior; or
- mark this part of Step 1 as blocked/incomplete in `STATUS.md` and phrase the config-roster implementation as a temporary fallback pending authenticated validation, not as a completed endpoint-probe conclusion.

## Notes

- I did not run the Go test suite because this step only changes documentation/status files.
