# Plan Review: TP-043 Step 3 Tests

Verdict: Approved with test-scope watch-outs.

The Step 3 plan matches the prompt's explicit requirements: run the existing tests without the removed `SetDeleteMode` / `SetToolset` helpers and add a regression test proving independent `response.Options` values can produce divergent `_meta` delete-mode values. That is the right guard for the process-global state removal.

Watch-outs for implementation:

1. Make the divergent-options test sequential and explicit.
   - In one test, call `response.Shape` at least twice with the same input and different `Options{DeleteMode: ...}` values, e.g. `safe` then `full` or `full` then `safe`.
   - Assert both returned `_meta.delete_mode` values and, ideally, also assert `_meta.toolset` remains the expected default/core value so the test catches accidental metadata drift.
   - Do not use any package-level setup or environment mutation; the test should demonstrate that the only input is `response.Options`.

2. Consider extending the same regression to `Toolset`.
   - The task removed global state for both delete-mode and toolset. Existing toolset tests already cover option consumption, but a paired `Options{Toolset: core}` / `Options{Toolset: full}` assertion would make the no-leakage guarantee symmetrical.
   - This is non-blocking relative to the prompt, but it is a low-cost regression guard.

3. Run the affected package tests, not only `internal/response`.
   - Step 2 touched response shaping, tool registry wiring, resources, app startup wiring, athlete-profile shaping, and `icuvisor_list_advanced_capabilities`' captured toolset path.
   - Before marking Step 3 complete, run at least:
     - `go test ./internal/response ./internal/tools ./internal/resources ./internal/app ./internal/athleteprofile`
   - Step 4 can still own the full `make test`, race, lint, and build verification.

4. Include a grep check for removed test helpers.
   - Confirm no tests still reference `SetDeleteMode`, `SetToolset`, `response.DeleteMode`, or `response.Toolset`.
   - This directly supports the acceptance criterion that tests no longer depend on the removed globals.

No plan changes are required before implementation, provided "existing tests pass without `Set*`" is interpreted to include the affected packages above.
