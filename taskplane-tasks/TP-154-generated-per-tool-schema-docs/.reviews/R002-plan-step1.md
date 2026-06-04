# Plan Review: Step 1 — Design generated docs data shape

**Verdict:** Approved

The updated `STATUS.md` now contains concrete Step 1 design notes and addresses the prior R001 blockers: it identifies `cmd/gendocs` as the implementation target, chooses a separate `web/data/tool_schemas.json`, defines a concise projection, describes nested/object-heavy handling, sets an examples policy, and records deterministic/no-secret boundaries.

## Notes for implementation

- Prefer a single canonical nested field name in the generated contract (`children` or `properties`, not both). The plan currently says `properties`/`children`; choose one in code/tests so Hugo templates and golden output stay stable.
- Keep `tools.json` summary-only as planned; update `make docs-tools`/`cmd/gendocs` so the current command still writes both generated files without requiring a second manual step.
- Add golden/determinism coverage for `tool_schemas.json`, especially that write tools include capped `input_examples` and large object fields (`workout_doc`, custom item `content`, sport zones) are summarized rather than raw-recursed.
- The placeholder-ID policy is acceptable as long as generated examples do not contain `athlete_id`, API keys, local paths, or `i\d{4,}`-style real-athlete-looking values.

No further plan review is needed before Step 2.
