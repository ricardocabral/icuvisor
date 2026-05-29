# Code Review: Step 3 — Update cookbook docs

Verdict: APPROVE

No blocking findings. The cookbook update clearly tells assistants to state missing/null Intervals readiness before falling back to HRV, resting HR, sleep, subjective scales, and `_native` provider data, while preserving scale labels and avoiding a synthetic readiness score.

Verification run:
- `make web-build` — pass, with existing Hugo deprecation warnings only.
