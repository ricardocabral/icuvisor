# R006 Plan Review — Step 3: Generator binary

**Verdict:** Approve.

I read `PROMPT.md`, `STATUS.md`, and the current Step 2 implementation in `internal/tools/catalog.go`, `internal/tools/registry.go`, and `internal/tools/catalog_test.go`. Step 3 is now narrow enough that the checklist in `STATUS.md` is an acceptable plan: add a small `cmd/gendocs` command, write `tools.Catalog()` deterministically to `web/data/tools.json`, test it against a golden file, and commit the generated output.

## Required implementation details to preserve the acceptance criteria

These are not blockers to starting the step, but they should be followed during implementation:

1. **Keep `main` thin and test a `run` helper.**  
   Implement `main()` as argument plumbing plus `os.Exit`, and put flag parsing / generation in a small `run(args []string) error` or similar helper. The golden test can call that helper directly with `--out <tempdir>/tools.json` instead of shelling out to `go run` or writing into the repo.

2. **Use deterministic JSON with a trailing newline.**  
   `tools.Catalog()` already returns descriptors sorted by `group` then `name`; preserve that order and marshal the top-level array with two-space indentation. Add a final newline so repeated generation is byte-stable and conventional for committed data files.

3. **Make the atomic write real and local to the target directory.**  
   For `--out`, create the parent directory if needed, write to a temp file in the same directory, close it, then rename it over the destination; remove the temp file on error. Avoid stdout redirection as the primary behavior because Step 3 and the later Makefile/CI guard depend on the binary owning the write.

4. **Do not create a second catalog source.**  
   The generator should only call `tools.Catalog()` and marshal the returned descriptors. The golden test may compare the emitted bytes to `cmd/gendocs/testdata/tools.golden.json`, but production code should not read the golden file or hand-maintain any tool metadata.

5. **Generate the committed artifact in this step.**  
   Add `web/data/tools.json` by running the new binary, not by copying expected JSON by hand. Since `web/data/` does not currently exist, this step should create it.

6. **Keep errors actionable and panic-free.**  
   Outside `main`, return wrapped errors (`fmt.Errorf("writing catalog: %w", err)`, etc.). If `main` prints errors, keep them short and avoid stack traces.

## Suggested checks for the step

- `go test ./cmd/gendocs ./internal/tools`
- `go run ./cmd/gendocs --out web/data/tools.json`
- `git diff -- web/data/tools.json cmd/gendocs/testdata/tools.golden.json` to confirm the committed artifact and golden output are intentionally updated

No plan changes required before implementation.
