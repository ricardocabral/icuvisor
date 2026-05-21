# Plan Review: Step 4 build + lint + race + live re-validation

Result: approve.

The revised Step 4 plan addresses the blocker from R009. It now separates local verification from live validation, explicitly validates the new public behavior for unsupported `feel`, and adds the key safety constraint for the live account: do not create another locked row. Using either a fresh row for only the six overwrite-able non-lock subjective fields, or the already-contaminated Step 1 row if `locked` must be exercised, is an acceptable mitigation given the known API cleanup limitation.

A few execution notes to keep the step safe and auditable:

- Run `make build`, `make test`, `make test-race`, and `make lint` first, and abort live mutation if any check fails.
- Source `.env-dev` without echoing secrets or raw athlete IDs.
- Snapshot the target row before any Step 4 write, including before the `feel` rejection probe if it targets an existing row.
- For accepted-path validation, prefer avoiding `locked:true` on fresh rows. If `locked` is included, use only the already-locked Step 1 probe row and restore all overwrite-able fields afterward.
- Re-read after each live validation and record in `STATUS.md` exactly what was verified and what remains blocked. The only acceptable unresolved live-account state after Step 4 is the pre-existing Step 1 locked-row blocker.

With those notes followed, the plan is sufficient for Step 4.
