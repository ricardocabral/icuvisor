# Plan Review R009 — Step 4

Verdict: APPROVE

The revised Step 4 plan now addresses the blocking items from R008. It names the generated-file guidance concern for both `web/data/tools.json` and `web/data/tool_schemas.json`, includes stale-generation guard coverage or an explicit caveat/follow-up, requires a concrete catalog/schema key comparison, and records the TP-153 dependency boundary.

## Implementation notes

- Prefer updating the actual stale-docs CI guard to diff both generated files rather than recording only a caveat, unless there is a clear scope reason not to touch `.github/workflows/ci.yml`.
- When updating contributor/user guidance, check `README.md`, `web/README.md`, and the `Makefile` help text because they currently describe only `web/data/tools.json` / “tool catalog”.
- Record the internal-only exposure check result in `STATUS.md` after comparing `web/data/tool_schemas.json` keys with `web/data/tools.json` names or the live catalog.
- Keep full `make test`, `make lint`, and `make build` deferred to Step 5 as planned.
