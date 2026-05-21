# Review: Step 1 plan

Verdict: **request changes before executing Step 1**.

The intent of Step 1 is right — re-verify against the current tree — but the current checklist still assumes the prompt's stale conflict shape. The tree I inspected already differs materially:

- `docs/prd/PRD-icuvisor.md:286-330` describes analyzers as **planned v0.6 roadmap scope** and says the generated catalog is the current registered-tool source of truth; I did not find the old `~39 tools at v1.0` target in the PRD section.
- `ROADMAP.md:82-98` already contains `## v0.6 — Analyzers`.
- `ROADMAP.md:22` has a single deferred `get_planning_parameters` line; I did not find a second checked-off ROADMAP line in the current file.
- `README.md` is now slimmed down and no longer contains an `MCP tool catalog` section.
- `internal/tools/update_wellness.go:18,178-207` already exposes/returns `field_not_writable: sleepScore (device-managed)` and `_native` equivalent; `web/data/tools.json:259-264` includes the same summary.

Required plan adjustments:

1. **Change the wording from “confirm these still exist” to “verify whether they still exist.”** For example, the `get_planning_parameters` bullet should not require proving both ROADMAP lines still contradict when the current tree appears to have only the deferred line. Step 1 should record “resolved/no longer present” where applicable, with grep evidence.

2. **Do not rely only on `internal/tools/registry.go` for registered tools.** The actual base catalog is built in `internal/tools/catalog.go` via `registryBaseTools`; `registry.go` delegates to that helper. Analyzer and `get_planning_parameters` verification should include `internal/tools/catalog.go`, `internal/toolcatalog/catalog.go` if relevant, and/or `tools.Catalog()` / `web/data/tools.json` grep evidence.

3. **Account for TP-051/TP-052/TP-054 landing.** Since README is no longer the main tool reference, Step 1 should explicitly check and record the current replacement surfaces for Conflict C: `web/data/tools.json` and `web/content/reference/tools.md`/generated tool reference behavior, in addition to noting that README has no tool-catalog bullet to edit.

4. **Record current-state evidence, not expected remediation.** `STATUS.md` should capture file paths, line numbers, and grep outputs showing which conflicts remain, which are already resolved, and whether later steps become no-ops or only need status/changelog reconciliation. Avoid reintroducing analyzer deferral text or ROADMAP edits that are already present.

Once those adjustments are made, the Step 1 plan is appropriate for a re-verification pass.
