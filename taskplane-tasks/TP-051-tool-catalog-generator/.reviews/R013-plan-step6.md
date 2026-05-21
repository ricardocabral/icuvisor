# R013 Plan Review — Step 6: Reconcile documentation conflicts

Decision: REQUEST CHANGES

I read `PROMPT.md`, `STATUS.md`, and spot-checked the current PRD/ROADMAP/README/registry/update-wellness state. The Step 6 checklist has the right high-level intent, but it still mirrors stale prompt assumptions in places and should be tightened before implementation so this step surfaces current truth rather than adding another stale note.

## Blocking plan adjustments

1. **Analyzer-family finding must reflect the current ROADMAP, not the original prompt.**  
   The TP-051 prompt says `ROADMAP.md` does not mention an analyzer phase, but the current tree now has `## v0.6 — Analyzers` at `ROADMAP.md:83` with `analyze_*`, `compute_*`, and `get_fitness_projection` bullets. The Step 6 plan should explicitly say to record the current verified state: PRD still targets analyzers and `~39 tools`, registry/catalog still have zero analyzer tools, README currently has no analyzer bullets, and ROADMAP now does have an analyzer phase. Do not write a STATUS/CHANGELOG note claiming the roadmap is missing an analyzer phase.

2. **Do not create a duplicate follow-up if TP-055 is already the follow-up.**  
   `taskplane-tasks/TP-055-reconcile-doc-conflicts/` already exists and covers the analyzer family, `get_planning_parameters`, and `update_wellness` documentation conflicts. The plan currently says to “note the follow-up TP title”; make that concrete: record TP-055 as the existing follow-up task, or explain why a new follow-up is still needed. Creating another “Reconcile analyzer family — defer in PRD or add roadmap phase” note would fragment the work.

3. **Keep the scope strictly “surface only.”**  
   Step 6 for TP-051 should update only TP-051 `STATUS.md` and `CHANGELOG.md` (plus regenerate `web/data/tools.json` only if needed by a prior descriptor change). It must not resolve TP-055-owned conflicts by editing `docs/prd/PRD-icuvisor.md`, `ROADMAP.md`, or the README tool list in this step. The plan should state that explicitly because TP-055’s prompt does include those resolution edits, but this TP-051 step does not.

4. **Clarify the `update_wellness` outcome with current evidence.**  
   The current MCP description and generated catalog summary already mention the required one-liner behavior: `internal/tools/update_wellness.go:18` says device-owned `sleepScore` and `_native` fields are not writable, and `web/data/tools.json` contains the same summary. The plan should say Step 6 will verify and record that no code/descriptor change is needed unless the generated catalog is stale. The fuller `field_not_writable: sleepScore (device-managed)` error contract remains a TP-052/TP-055 documentation follow-up, not a TP-051 code change.

## Suggested revised Step 6 plan

- Run greps/reads for the three conflicts and capture current file/line evidence in `STATUS.md`.
- For analyzer family, record: PRD includes analyzer target; registry/catalog include zero analyzer names; README hand-list has no analyzers; ROADMAP now has `v0.6 — Analyzers`; TP-055 is the existing follow-up for final reconciliation.
- For `get_planning_parameters`, record the two ROADMAP checked lines and registry absence/presence; do not edit ROADMAP here.
- For `update_wellness`, record the PRD error contract, README/website summary gap, exact code literal, and that generated summary already includes the device-owned-field rejection clause.
- Add one `[Unreleased]` / `Changed` changelog bullet for the generated catalog/docs change and one concise note that the known doc divergences were surfaced and delegated to TP-055.

Once those points are reflected in the plan/status, implementation should be straightforward.
