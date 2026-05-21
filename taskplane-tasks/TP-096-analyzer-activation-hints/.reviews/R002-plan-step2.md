# Plan Review — TP-096 Step 2

**Verdict:** Approved with required clarifications

Step 2 is the right scope for this task: update only the registered tool descriptions and add catalog/check coverage so the activation-hint wording does not drift. The Step 1 audit gives enough target information to proceed, but the implementation should observe the guardrails below before marking the step complete.

## Required clarifications for Step 2

1. **Use the full analyzer-family list, not only catalog group `analyzers`.** The checks/descriptions must cover all 11 tools from the audit:
   - `analyze_trend`
   - `analyze_distribution`
   - `analyze_correlation`
   - `analyze_efforts_delta`
   - `compute_zone_time`
   - `compute_load_balance`
   - `compute_baseline`
   - `compute_compliance_rate`
   - `compute_activity_segment_stats`
   - `get_activity_histogram`
   - `get_fitness_projection`

   `get_activity_histogram` and `get_fitness_projection` are analyzer-family tools for this roadmap item even though their current generated-doc groups are `activities`/`fitness`. Do not accidentally omit them by filtering on `ToolDescriptor.Group == "analyzers"`.

2. **Update existing catalog tests, not just add a new one.** `internal/tools/catalog_test.go` already has `TestCatalogIncludesFullAnalyzers`, but it is incomplete for TP-096 and appears to assert older summary wording for several analyzer tools. Step 2 should revise/replace that coverage so it asserts the new activation-hint contract rather than stale phrases.

3. **Be careful with toolcheck catalog generation.** `internal/toolchecks.GenerateToolCatalog` currently goes through `schemaCatalogToolNames`, which is a schema-stability allowlist and does not include the analyzer-family tools. A TP-096 check built on that filtered catalog can pass while checking none of the relevant tools. Either:
   - add a description-specific registered-catalog path that includes these analyzer-family tools, or
   - keep the live-catalog assertions in `internal/tools` and use `internal/toolchecks` only for pure helper logic.

   Avoid expanding the schema snapshot allowlist just to get descriptions unless Step 2 intentionally updates the corresponding schema snapshots; that would broaden the task.

4. **Make the compliance rule semantic but explicit.** The first sentence should lead with a concrete prompt shape (for example, `Use when the prompt/user asks ...`) instead of an implementation verb like `Project`, `Summarize`, or `Compute`. Separately, each relevant description should include explicit guidance not to pull/fetch `get_*` rows/streams and reduce/bin/compute them manually in chat.

5. **Handle `compute_activity_segment_stats` as the raw-stream exception.** Its description still needs the activation hint and should tell the model to use this server-side bounded-segment computation instead of calling `get_activity_streams` and calculating the segment manually. The check should not read as a blanket prohibition on the tool's internal stream use.

6. **Leave generated docs/changelog to Step 3 unless intentionally advancing.** Step 2 should focus on source descriptions and tests/checks. Regenerating `web/data/tools.json`, `web/content/reference/tools.md`, and updating `CHANGELOG.md` is already planned for Step 3.

## Verification note

I ran:

```sh
go test ./internal/toolchecks -count=1
```

It currently fails at package build time due to the known pre-existing `internal/tools/compute_baseline.go` duplicate helper/signature errors recorded in `STATUS.md`. Step 2 can still add the intended tests, but the worker should keep this blocker visible and not mark targeted toolcheck/catalog tests as passing until the build issue is resolved or explicitly documented as pre-existing for this lane.
