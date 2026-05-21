# Plan Review: TP-001 Step 2

Verdict: **Approved for implementation**

The revised Step 2 plan in `STATUS.md` addresses the previous concerns and is aligned with the prompt:

- `main` stays thin and delegates to an `internal/app` entrypoint such as `app.Run(ctx, Options{...})`.
- `icuvisor version` remains supported and is planned for practical coverage at the app layer with injected args and writers.
- The default invocation delegates to an internal startup path without pulling in Step 3 config loading/validation.
- Build-time version is injected from `main` into app/runtime startup state instead of being imported from `main` by lower layers.
- Internal packages return errors; final stderr/exit-code handling remains in `main`.

## Implementation notes

- Keep Step 2 narrowly scoped: no env/JSON config precedence, `.env` parsing, intervals client, or MCP SDK wiring unless needed as a compile-time stub.
- Make the starter function context-aware now, since it will become blocking I/O when stdio MCP startup is implemented.
- Prefer `io.Writer` fields for stdout/stderr in `app.Options` so tests can use buffers.
- If unknown-command handling is introduced, keep the returned error short and actionable, and cover it in the app tests as planned.
- `CHANGELOG.md` can remain for Step 5 unless this step is committed independently with user-visible CLI changes.

No blocking changes are required before coding Step 2.
