# Plan Review — Step 1: Snapshot pre-refactor output

**Decision: Approved.**

The updated Step 1 plan addresses the blocking issues from R001: it names the five fixture files, requires deterministic metadata/options setup, adds an automated golden comparison/regeneration mechanism, avoids network calls, and keeps the fixture commit before any `shaper.go` refactor. That is sufficient safety net planning for this step.

## Things to preserve during implementation

- Keep the golden test in `internal/response` so it can use `resetRuntimeCatalogMetadataForTest`, and do not run the cases in parallel unless catalog runtime state is fully isolated per case.
- Make the normal test path compare generated canonical, indented JSON bytes against checked-in fixtures; fixture updates should be explicit, e.g. via an update flag or generator command, not automatic on every test run.
- Use synthetic typed DTO structs with JSON tags for at least some cases, not only `map[string]any`, so the snapshots cover the current marshal/tag/omitempty behavior that the refactor is trying to preserve.
- Pin stable `Options` values in the case table (`ServerVersion`, `DeleteMode`, `Toolset`, `UnitSystem`, `FetchedAt` when debug metadata is enabled) and reset catalog metadata before each case.
- For `get_activities_full.golden.json`, include explicit nulls in the input so `include_full: true` proves null preservation.
- For `wellness_provenance.golden.json`, include `_meta.provenance.<field>.fetched_at` plus ordinary debug-looking fields elsewhere, so the special provenance exemption is covered.

No additional plan changes are required before Step 1 implementation proceeds.
