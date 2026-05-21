# Plan Review — Step 2: Mechanical file split

Decision: **Approved**

The Step 2 plan in `STATUS.md` is sufficient to proceed with the mechanical split. Step 1 produced the important safety artifacts: a declaration-by-declaration inventory, a concrete regression gate, and notes about existing adjacent files (`prompts.go`, `catalog_hash.go`) and the later coach/tools import-cycle constraint. The Step 2 checklist follows that inventory and keeps the coach-behavior lift explicitly deferred to Step 3.

## What I checked

- Re-read `PROMPT.md` and `STATUS.md` for the Step 2 scope and acceptance criteria.
- Compared the Step 2 checklist against the current declarations in `internal/mcp/server.go`.
- Verified the plan accounts for the non-obvious existing files:
  - `internal/mcp/prompts.go` already owns the prompt registrar.
  - `internal/mcp/catalog_hash.go` already owns catalog hash helpers and can host `(*Server).CatalogHash`.
- Verified the plan preserves the required sequencing: transport, schema/conversion, tool registrar, resource registrar, recovery helper, then final narrowing of `server.go`.

## Notes to apply during implementation

- Keep this step purely mechanical. Do not change coach ACL behavior, advanced-capabilities rendering, schema text, error strings, or catalog hash semantics while moving declarations.
- Run the recorded Step 1 gate after each concern move:

  ```sh
  go test ./internal/mcp ./internal/tools ./internal/coach ./internal/toolcatalog
  ```

  If time allows, also run `go test ./...` after the larger moves because the prompt says to run checks after each move, but the targeted gate is the minimum regression gate recorded for this refactor.
- Do not create `registrar_prompts.go` unless intentionally renaming `prompts.go`; duplicate prompt registrar files would be a regression.
- Move `(*Server).CatalogHash` into `catalog_hash.go` when narrowing `server.go`, rather than creating another metadata file unless there is a clear need.
- When extracting imports, watch for accidental behavior changes caused by helper relocation. In particular, leave package-private names package-private and keep `withPanicRecovery` behavior byte/semantically identical.
- After the final split, `internal/mcp/server.go` should remain under the acceptance target and contain only the constructor/server surface. Add the package doc note expected by the task either in `server.go` or an explicit package doc file before final verification.

No plan changes are required before starting Step 2.
