# Plan Review — TP-096 Step 3

**Verdict:** Changes requested

The Step 3 checklist has the right intent, but it is not yet executable as a plan in the current tree. Regenerating the public catalog is mandatory after Step 2 changed tool descriptions, and the current catalog generator cannot run until the known `compute_baseline` package build errors are resolved or the step is explicitly marked blocked. I confirmed this with a temp-output generation attempt:

```sh
go run ./cmd/gendocs --out "$tmp"
```

It fails while compiling `internal/tools` with the same `compute_baseline.go` duplicate helper / wrong-signature errors already recorded in `STATUS.md`. Step 3 should not hand-edit generated outputs or mark regeneration complete while that build blocker remains.

## Required plan changes

1. **Handle the generator build blocker first.** The plan must say what happens if `make docs-tools`/`go run ./cmd/gendocs` still fails:
   - either fix/rebase onto the dependency repair before regenerating artifacts, or
   - stop and mark Step 3 blocked in `STATUS.md` with the exact compiler errors.

   Do not manually edit `web/data/tools.json` to work around a failing generator.

2. **Regenerate both generated catalog artifacts.** `make docs-tools` updates `web/data/tools.json`, but `cmd/gendocs/main_test.go` compares generator output to `cmd/gendocs/testdata/tools.golden.json`. Since Step 2 changed descriptions, the golden file will also need to be refreshed from the same generator output path or `make test` will fail once the package builds. Add `cmd/gendocs/testdata/tools.golden.json` to the Step 3 affected-file list.

3. **Inspect the generated diff, not just command success.** After generation, review a path-scoped diff for the analyzer-family descriptions, for example:

```sh
git diff -- web/data/tools.json cmd/gendocs/testdata/tools.golden.json
```

Verify that only the intended summary/description text changed and that all 11 analyzer-family tools from Step 1 are represented where registered.

4. **Make the rendered-docs review concrete.** `web/content/reference/tools.md` is a Hugo page that renders `web/data/tools.json` through `{{< tool-catalog >}}`; it usually should not be hand-edited for this task. The Step 3 plan should either run `make web-build` and inspect the rendered reference page, or document if Hugo is unavailable and fall back to reviewing the JSON plus the static reference page wrapper.

5. **Update `CHANGELOG.md` under `[Unreleased]`.** Public MCP tool descriptions and generated website catalog text are user-visible. Add a concise `Changed` entry for the analyzer-family activation-hint wording rather than relying on the existing analyzer `Added` entries.

With those changes, Step 3 will be appropriately scoped: fix/block on the build issue, regenerate from source of truth, update the generator golden, review the rendered/catalog diff, and record the user-visible documentation change.
