# Code Review — TP-039 Step 1 (R002)

## Findings

### [P1] Do not keep the endpoint-probe checklist item checked while the authenticated probe is blocked

`STATUS.md` now correctly records a blocker saying the authenticated coach-key roster probe could not be completed because no real coach-scoped key was available (lines 70-72), and the threat-model doc also says the authenticated probe is incomplete/not passed (`docs/threat-models/coach-mode.md` lines 119-123). However the Step 1 checklist still marks “Coach-roster endpoint probed; path/auth/shape documented OR gap documented” as complete (`STATUS.md` line 18).

That is still the same acceptance problem from R001: the task required a black-box probe with a real coach key, or a confirmed free-tier/endpoint exposure gap. The current work only documents public OpenAPI plus unauthenticated responses and a missing local credential, so downstream steps can incorrectly treat the roster endpoint as validated. Please change this checklist item to incomplete/blocked (for example `- [ ] ... (blocked: no real coach key available)`) and keep Step 1 incomplete until an authenticated coach-key probe is performed or a real plan/endpoint gap is confirmed.

## Notes

- I did not run the Go test suite because this step only changes documentation/status files.
