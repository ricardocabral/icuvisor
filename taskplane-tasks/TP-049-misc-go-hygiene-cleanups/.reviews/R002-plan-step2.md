# Plan Review: Step 2 — Propagate `ctx` through toolchecks

Verdict: Approved with implementation clarifications.

The step is correctly scoped to removing the two `context.Background()` calls inside `internal/toolchecks/` and letting the caller supply the context used for `tools.Registry.Register`. This matches the task acceptance criteria and should not require any schema, catalog, or user-visible behavior changes.

Implementation notes to keep the change focused:

- There is no separate `toolchecks.Register` function today. The actual public helpers that need context are:
  - `GenerateSchemaSnapshots` in `internal/toolchecks/schema_stability.go`
  - `GenerateToolCatalog` in `internal/toolchecks/confusable_names.go`
  - `WriteGeneratedSchemaSnapshots` should also accept/pass `ctx` because it calls `GenerateSchemaSnapshots`.
- Use `ctx` directly in the existing calls: `registry.Register(ctx, registrar)`. Do not change the `tools.Registry` interface or registry behavior; it already accepts `context.Context`.
- Update all callers, notably:
  - `scripts/check_schema_stability.go` → `toolchecks.GenerateSchemaSnapshots(context.Background())`
  - `scripts/snapshot_tool_schemas.go` → `toolchecks.WriteGeneratedSchemaSnapshots(context.Background(), *dir)`
  - `scripts/check_confusable_names.go` → `toolchecks.GenerateToolCatalog(context.Background())`
  Program-root `context.Background()` in these `main` packages is acceptable for this task.
- Update tests to use `t.Context()` rather than adding new `context.Background()` calls under `internal/toolchecks/`, because Step 6 explicitly checks `grep -rn "context.Background()" internal/toolchecks/` is empty.
- Preserve existing error wrapping (`fmt.Errorf("registering tools: %w", err)`) so cancellation from a passed context remains traceable without changing output shape.
- If adding `ctx` to `WriteGeneratedSchemaSnapshots`, keep the signature context-first: `func WriteGeneratedSchemaSnapshots(ctx context.Context, dir string) error`.
- Keep this to the cited sites. Do not broaden the change into other test helpers or unrelated `context.Background()` use elsewhere in the repository.

Recommended verification for this step:

- `grep -rn "context.Background()" internal/toolchecks/` returns no matches.
- `go test ./internal/toolchecks ./scripts/...` is not applicable for build-tagged scripts, so rely on the requested `make build`, `make test`, `make test-race`, and `make lint`; if desired, also run `go run ./scripts/check_schema_stability.go` and `go run ./scripts/check_confusable_names.go` as smoke checks.

No blockers found.
