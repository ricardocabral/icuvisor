# Review R004 — Plan Review for Step 2

**Verdict:** Approved

The Step 2 plan can proceed. The approved Step 1 measurement method in `STATUS.md` is specific enough for a reproducible implementation: it fixes fixture roots, exclusion precedence, object eligibility, precomputed-zone fields, per-family opportunity semantics, and threshold policy. Step 2's checklist maps directly to that method: add a small offline scanner/helper, run it against the fixture corpus, and record the command/results.

## Notes for implementation

- If the helper is a Go script under `scripts/`, follow the existing pattern and add `//go:build ignore` so `go test ./...` and normal builds do not pick it up as part of the `scripts` package.
- Make traversal deterministic: sort file paths and sort/report families consistently. Include the exact command used in `STATUS.md` so the audit is reproducible.
- Apply hard path/type exclusions before any shape detection. Excluded events, wellness, workout library, custom items, activity messages, activity intervals, analyzer goldens, and schema snapshots should never re-enter due to broad `id`/date fields.
- Keep family signals value-aware. Present-but-null or zero-duration fields should not create fallback opportunities; this is especially important for sparse Strava/import stub fixtures.
- Report enough detail to audit the aggregate numbers: totals by metric family plus a per-fixture/per-family line or summary of which units were precomputed, fallback, or unknown.
- Non-zero fallback is evidence for Step 3 documentation, not a script failure. The script should fail only on malformed input, implementation errors, or explicitly invalid audit assumptions.

No blocking plan changes are required before implementing/running the audit.
