# Review R004 — Plan review for Step 2

Verdict: **approved**

The revised Step 2 plan addresses the blocking concerns from R003 and is scoped appropriately for TP-056.

## Why this is ready

- The plan targets the actual remaining leak risk in `internal/intervals/events.go`: the success-path `defer resp.Body.Close()` inside the `doJSONBody` retry loop.
- It explicitly uses `decodeErr` / `closeErr` handling, which preserves closure on every terminal 2xx path, including invalid-JSON decode errors.
- It avoids `readBody` / `io.LimitReader`, keeping the task within the prompt's “Do NOT” constraints and avoiding a response-size behavior change.
- It preserves the existing non-2xx drain-and-close retry branch and does not change retry eligibility, `Retry-After`, backoff, or ignored close errors on retry.
- The added invalid-JSON regression case is a good guard against the most likely unsafe implementation of removing the defer.
- The planned checks and `CHANGELOG.md` update match the task acceptance criteria.

## Implementation guardrails

- Keep the final 2xx path ordered as: decode, close, then return the decode error first if present, otherwise close error.
- Do not touch retry policy or introduce broader helper refactors in this task.
- After the change, verify `grep -n "defer resp.Body.Close" internal/intervals/` does not show a defer inside a retry loop.
