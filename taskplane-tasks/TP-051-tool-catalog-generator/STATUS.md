# TP-051-tool-catalog-generator — Status

**Current Step:** Step 8: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-17
**Review Level:** 2
**Review Counter:** 20
**Iteration:** 1
**Size:** M

---

### Step 1: Design `ToolDescriptor`

**Status:** ✅ Complete

- [x] Decide the exact `ToolDescriptor` fields and JSON shape.
- [x] Define deterministic catalog ordering and pretty-printing requirements.
- [x] Define summary extraction from MCP descriptions as the first sentence.

### Step 2: Wire `Catalog()` into the registry

**Status:** ✅ Complete

- [x] Add `tools.Catalog()` that returns every tool with tier, safety, group, summary, and anchor without applying safety/toolset gates.
- [x] Add a table-driven `catalog_test.go` covering descriptor uniqueness, snake_case names, PRD registered-tool overlap, analyzer exclusions, and registry parity.
- [x] R004: Share the registry enumeration path between `Register` and `Catalog()` instead of duplicating constructor order.
- [x] R004: Make catalog group metadata exhaustive so unknown groups fail tests instead of rendering as `other`.
- [x] R004: Fix catalog lint preallocation and rerun targeted tests/lint.

### Step 3: Generator binary

**Status:** ✅ Complete

- [x] Add `cmd/gendocs` with `--out` support that marshals `tools.Catalog()` deterministically and writes atomically.
- [x] Add a generator golden test that writes to a tempdir and compares against `cmd/gendocs/testdata/tools.golden.json`.
- [x] Generate and commit `web/data/tools.json` from the binary.

### Step 4: Hugo rendering

**Status:** ✅ Complete

- [x] R008: Record that TP-050/Hextra is absent in this worktree and implement the minimal non-Hextra reference layout/navigation needed for `/reference/tools/` instead of assuming Hextra.
  - Note: R008 verified this worktree has the pre-TP-050 custom Hugo site (no Hextra theme, no `_default/single.html`, no existing `reference/` section). Step 4 therefore adds the minimal non-Hextra single-page layout/navigation required to render `/reference/tools/` from generated data.
- [x] Add `web/content/reference/tools.md` with frontmatter, intro, and `tool-catalog` shortcode call.
- [x] Add shared Hugo partials plus a shortcode wrapper to group `site.Data.tools`, render readable group labels, and render tier/safety badges.
- [x] Configure landing-page featured tools in `web/hugo.toml`, validate stale names with Hugo `errorf`, and replace hardcoded chips with the shared data-source partial.
- [x] Add the minimal toolset-tier page/link target and CSS for catalog tables and tier/safety badges.

### Step 5: Makefile + CI guard

**Status:** ✅ Complete

- [x] Add a `docs-tools` Makefile target that runs `go run ./cmd/gendocs --out web/data/tools.json` idempotently.
- [x] List `docs-tools` in `make help`.
- [x] Add a CI guard step that runs `make docs-tools` and `git diff --exit-code web/data/tools.json` with a clear failure message.

### Step 6: Reconcile documentation conflicts (surface only)

**Status:** ✅ Complete

- [x] Verify and record the current analyzer-family PRD/registry/README/ROADMAP divergence in `STATUS.md`: PRD targets analyzers/~39 tools, registry/catalog have zero analyzers, README has no analyzers, ROADMAP now has `v0.6 — Analyzers`, and TP-055 is the existing reconciliation follow-up.
- [x] Verify and record the `get_planning_parameters` ROADMAP contradiction in `STATUS.md`, including registry truth, without editing ROADMAP.
- [x] Verify and record the `update_wellness` error-contract gap and current evidence that the MCP/generated summary already mentions device-owned `sleepScore`/`_native` rejection; leave full docs resolution to TP-052/TP-055.
- [x] Update `CHANGELOG.md` `[Unreleased]` with the generator change and a concise note that known doc divergences were surfaced and delegated to TP-055.

Scope note: Step 6 is surface-only. Do not edit `docs/prd/PRD-icuvisor.md`, `ROADMAP.md`, or the README tool list here; README prose replacement is Step 7.

Step 6 findings:

- Analyzer-family divergence verified: `docs/prd/PRD-icuvisor.md` §7.2.C lists the analyzer family (`analyze_trend`, `analyze_distribution`, `analyze_correlation`, `analyze_efforts_delta`, `compute_zone_time`, `compute_load_balance`, `compute_baseline`, `compute_activity_segment_stats`, `compute_compliance_rate`, `get_fitness_projection`) and `~39 tools` at v1.0; `internal/tools/registry.go` plus `tools.Catalog()` register zero analyzer names; the current README tool list has no analyzer bullets; `ROADMAP.md` now has `## v0.6 — Analyzers` with those tools planned. Existing follow-up: `TP-055-reconcile-doc-conflicts`; do not create a duplicate follow-up.
- `get_planning_parameters` contradiction verified: `ROADMAP.md` line 22 says the tool is deferred until intervals.icu exposes athlete-level periodization parameters, while line 29 lists periodization parameters via `get_planning_parameters` as checked off. Registry truth: `internal/tools/registry.go`, `internal/toolcatalog/catalog.go`, `tools.Catalog()`, and `web/data/tools.json` contain no `get_planning_parameters` entry. ROADMAP resolution remains out of scope for TP-051 and is covered by TP-055.
- `update_wellness` error-contract gap verified: `docs/prd/PRD-icuvisor.md` §7.2.C requires rejecting device-owned `sleepScore` with `field_not_writable: sleepScore (device-managed)`. Current code already enforces this (`internal/tools/update_wellness.go` read-only fields include `sleepScore`/`_native`, and validation returns `field_not_writable: sleepScore (device-managed)` or `_native (bridge-managed)`). The MCP description and generated `web/data/tools.json` summary already include the one-line device-owned `sleepScore`/`_native` rejection clause, and the current README also mentions that rejection. Full website error-contract prose remains TP-052/TP-055 scope.

### Step 7: Replace duplicate prose

**Status:** ✅ Complete

- [x] Replace the hand-written README MCP tool catalog list with a short paragraph linking to `https://icuvisor.app/reference/tools/`.
- [x] Remove any remaining hardcoded landing-page tool-chip/count prose in `web/layouts/index.html` (including the stale `~25 MCP tools` feature-list claim) so chips/counts render from generated catalog data or avoid numeric claims.
- [x] Update `web/README.md` to note that `web/data/tools.json` is generated and must not be hand-edited.
- [x] R018: Remove the remaining stale README feature bullet that claims `~25 MCP tools`.

### Step 8: Verify

**Status:** ✅ Complete

- [x] `make docs-tools` produces a clean diff.
- [x] `make build`, `make test`, `make test-race`, and `make lint` all pass.
- [x] CI generated-catalog guard fails for a temporary stale catalog and the throwaway change is reverted.
- [x] `cd web && hugo --minify --gc` produces `reference/tools/` from `web/data/tools.json`.
- [x] Manual count: website tools match `len(tools.Catalog())` and `web/data/tools.json` entry count.
  - Verified counts: `len(tools.Catalog())` = 40, `web/data/tools.json` entries = 40, rendered `reference/tools/` table rows = 40.

| 2026-05-17 11:22 | Task started | Runtime V2 lane-runner execution |
| 2026-05-17 11:22 | Step 1 started | Design `ToolDescriptor` |
| 2026-05-17 11:26 | Review R001 | plan Step 1: APPROVE |
| 2026-05-17 11:30 | Review R002 | code Step 1: APPROVE |
| 2026-05-17 11:33 | Review R003 | plan Step 2: APPROVE |
| 2026-05-17 11:40 | Review R004 | code Step 2: UNKNOWN |
| 2026-05-17 11:46 | Review R005 | code Step 2: APPROVE |
| 2026-05-17 11:49 | Review R006 | plan Step 3: APPROVE |
| 2026-05-17 11:52 | Review R007 | code Step 3: APPROVE |
| 2026-05-17 11:55 | Review R008 | plan Step 4: UNKNOWN |
| 2026-05-17 11:57 | Review R009 | plan Step 4: APPROVE |
| 2026-05-17 12:04 | Review R010 | code Step 4: APPROVE |
| 2026-05-17 12:10 | Review R011 | plan Step 5: APPROVE |
| 2026-05-17 12:12 | Review R012 | code Step 5: APPROVE |
| 2026-05-17 12:15 | Review R013 | plan Step 6: REVISE |
| 2026-05-17 12:17 | Review R014 | plan Step 6: APPROVE |
| 2026-05-17 12:21 | Review R015 | code Step 6: APPROVE |
| 2026-05-17 12:24 | Review R016 | plan Step 7: REVISE |
| 2026-05-17 12:25 | Review R017 | plan Step 7: APPROVE |
| 2026-05-17 12:29 | Review R018 | code Step 7: UNKNOWN |
| 2026-05-17 12:49 | Review R020 | code Step 7: APPROVE |

| 2026-05-17 12:54 | Worker iter 1 | done in 5540s, tools: 257 |
| 2026-05-17 12:54 | Task complete | .DONE created |