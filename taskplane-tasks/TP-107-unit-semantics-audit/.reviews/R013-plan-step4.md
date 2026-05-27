# Plan Review — Step 4

Verdict: **REVISE**

The Step 4 checklist is directionally correct, but `STATUS.md` does not yet contain an implementation plan for this step. Given Step 3 added user-visible response metadata for wellness hydration semantics, the changelog update is no longer optional: Step 4 should explicitly plan the entry and the verification commands before proceeding.

Required plan additions:

1. Add a concrete `CHANGELOG.md` plan. Put an `[Unreleased]` entry under the appropriate section noting that wellness rows now include `_meta.field_semantics` for `hydration` and `hydrationVolume` without renaming or inventing units. Keep it concise and user-visible.
2. Specify the final discovery updates to `STATUS.md`: no serializer behavior change beyond tests, energy/joule surfaces covered/audit-only findings, and hydration metadata added while null hydration fields avoid stale semantics.
3. Name the targeted tests to run for all affected packages, for example `go test ./internal/workoutdoc ./internal/units ./internal/response ./internal/tools` or narrower exact `-run` patterns covering the added Step 1-3 tests. Record the command and result in the execution log.
4. Clarify the boundary with Step 5: Step 4 should do affected-package verification; `make test`, `make build`, and `make lint` remain Step 5 unless the plan intentionally moves them earlier.
5. If no tool schema/catalog text changed, note that generated tool docs do not need regeneration; if any schema/description text is touched while writing the changelog, plan to regenerate/check the catalog.

Once these details are added, the step should be ready to execute.
