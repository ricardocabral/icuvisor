# R002 Plan Review — Step 2: Split

Verdict: APPROVE, with one implementation note to fold into Step 2.

The Step 2 plan matches the task constraints: it keeps the split mechanical, preserves registry order/call sites, moves only declarations shared by 2+ tools into `fitness_shared.go`, and adds a pre-refactor catalog/schema snapshot before moving code. That snapshot step is especially important for the byte-identical acceptance criteria and should happen before any file moves.

Implementation note:

- The prompt mentions `internal/tools/get_fitness_test.go`, but the actual repository has `internal/tools/get_fitness_metrics_test.go`. That file combines the four fitness-related tool shape subtests with `TestExtendedMetricsDropsUnavailableFieldsAndConvertsUnits` and fixture helpers used by extended metrics. When mirroring the test split, avoid naively moving/renaming the entire file in a way that strands or duplicates extended-metrics helpers. Prefer either:
  - split the four `TestFitnessMetricsToolShapes` cases into per-tool `_test.go` files while keeping the extended-metrics test/helpers in place; or
  - extract shared fake/fixture helpers into a small package-local `*_test.go` helper file used by the new per-tool tests and the extended-metrics test.

Non-blocking reminders for the split:

- Keep constructors' `coreTool` vs `fullTool` wrappers exactly as-is (`get_power_curves` remains `fullTool`; the other three remain `coreTool`).
- Do not reorder declarations in `registry.go`; the registration order is observable through the catalog.
- After splitting imports across files, let `gofmt/goimports` clean them, but do not reformat schema map literals beyond what is required to compile unless the schema snapshot confirms no JSON-visible change.
