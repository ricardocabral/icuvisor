# Plan Review — TP-096 Step 5

**Verdict:** Approved

The Step 5 plan covers the task's completion criteria: targeted verification, `make test`, `make build`, `make lint`, and a failure policy that requires fixing failures or documenting clearly pre-existing unrelated ones in `STATUS.md`.

A couple of execution notes to keep it tight:

- For the “targeted tests passing” checkbox, reuse the explicit targeted commands that already proved the affected surface in Step 4 rather than leaving it implicit:

  ```sh
  go test ./internal/tools -run 'TestCatalog' -count=1
  go test ./internal/toolchecks -run 'TestCheckConfusableCatalog|TestFirstDescriptionSentence|TestGenerateToolCatalogUsesCallerContext' -count=1
  go test ./cmd/gendocs -run TestRunWritesToolsCatalogGolden -count=1
  go test ./internal/tools -run 'TestComputeBaseline|TestComputeActivitySegmentStats|TestGetActivityHistogram|TestGetFitnessProjection' -count=1
  ```

- Do not substitute the Step 4 `make check` result for the Step 5 required commands. `make check` already gave useful confidence, but the prompt explicitly asks for:

  ```sh
  make test
  make build
  make lint
  ```

- If anything fails, record the exact command and concise output/rationale in `STATUS.md`; failures in analyzer descriptions, catalog generation, or the touched `compute_baseline` path should be treated as in-scope and fixed rather than documented away.

Minor status hygiene: before executing Step 5, update the top-level `STATUS.md` header from `Current Step: Step 4` to Step 5 so the header matches the checklist state.
