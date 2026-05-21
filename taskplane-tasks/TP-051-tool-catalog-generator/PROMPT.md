# TP-051 — Tool catalog generator (code is source of truth)

## Mission

Make the in-code MCP tool registry the single source of truth for the public tool catalog, and render the website's tool-reference page (and the landing-page tool list) from generator output. Today the catalog is described in three places that drift independently:

1. `README.md` §"MCP tool catalog" — hand-maintained list of ~36 tools.
2. `web/layouts/index.html` — hand-coded `<span class="tool">…</span>` chips (~18 tools, already stale and indicative-only).
3. `internal/tools/registry.go` (and the per-tool files) — the actual registration code.

`docs/prd/PRD-icuvisor.md` §7.2.C is the product spec for the catalog and currently disagrees with what is implemented: it lists an "analyzer family" (`analyze_*`, `compute_*`, `get_fitness_projection`) and a "~39 tools at v1.0" target, while `internal/tools/registry.go` registers zero analyzers and the README lists 36 tools. This task **does not change the PRD or implement analyzers** — but it must surface the divergence in `STATUS.md` and in the `[Unreleased]` `CHANGELOG.md` entry, and it must register a follow-up taskplane note so the next planning pass either defers the analyzers in the PRD or files a roadmap phase for them. Generator output should reflect _what is registered_, not what the PRD aspires to.

PRD anchors: §7.2.C (tool catalog), §6 KR5 (token efficiency — duplication of tool descriptions across README/site/code is exactly the kind of drift KR5 punishes), §7.4 #11 (schema/catalog stability).

ROADMAP positioning: docs and DX polish. No new user-facing capability.

Complexity: Blast radius 2 (no protocol changes; generator output consumed by website only), Pattern novelty 2 (first code-generated artifact in this repo for docs), Security 1, Reversibility 1 = 6 → Review Level 1–2. Size: M.

## Dependencies

- **TP-050** must land first (Hextra site has a `reference/` section to render into).
- All read/write tool TPs (TP-004 through TP-031, TP-039) are upstream sources of truth in the registry — no changes to those tasks; this task only **reads** the registry.

## Context to Read First

- `CLAUDE.md` — "default to `internal/`", "no `panic` outside `main`", Conventional Commits.
- `internal/tools/registry.go` — registration entrypoint; understand how every tool is wired in.
- `internal/tools/*.go` — sample a handful (e.g., `get_activities.go`, `update_wellness.go`, `apply_training_plan.go`) to see what metadata is attached to each tool (name, description, JSON schema, safety/toolset gating).
- `internal/safety/` — the delete-mode gate that determines which tools are registered.
- `internal/toolset/` (or equivalent) — the `core` vs `full` tier gate.
- `README.md` §"MCP tool catalog" — current hand-written prose (this is the content to replace; do **not** copy into the generator output verbatim — the generator should produce equivalent prose from code).
- `docs/prd/PRD-icuvisor.md` §7.2.C — the spec; note where it diverges from the registry (analyzer family).
- `web/layouts/index.html` — the existing hand-coded tool chips (will be replaced by data-driven rendering).
- Hugo data files: <https://gohugo.io/templates/data-templates/> — the rendering side reads `web/data/tools.json` via `site.Data.tools`.

## File Scope

Expected files (new unless noted):

- `cmd/gendocs/main.go` — new binary, Go-only, walks the tool registry and emits `web/data/tools.json`. Use the existing `internal/tools` registration entry point — do **not** duplicate metadata.
- `internal/tools/catalog.go` (new, or extend `registry.go`) — exposes a typed, pure-Go `Catalog()` function that returns `[]ToolDescriptor` with: `Name`, `Description`, `Tier` (`core`/`full`), `Safety` (`read`/`write`/`delete`), `Group` (e.g., `activities`, `wellness`, `events`, `workout-library`, `custom-items`, `meta`), one-line summary, link anchor. This function must be callable with no live MCP server and no API client — it is metadata only.
- `internal/tools/catalog_test.go` — table-driven test asserting every registered tool appears in `Catalog()` exactly once and vice versa; this is the drift guard.
- `web/data/tools.json` — committed generator output, pretty-printed, deterministic (stable ordering).
- `web/content/reference/tools.md` — Hextra page that renders `site.Data.tools` as a grouped table (one section per `Group`, columns for name / summary / tier / safety). Frontmatter only; the table is built in the layout.
- `web/layouts/shortcodes/tool-catalog.html` (new) — shortcode that renders the catalog or a filtered subset (used by `reference/tools.md` and by the landing page).
- `web/layouts/index.html` — replace the hardcoded `<span class="tool">` chips with a call to the same data source (e.g., a `tool-chips` shortcode that picks a short featured subset by group).
- `Makefile` — add `docs-tools` target: `go run ./cmd/gendocs > web/data/tools.json` (or equivalent; prefer a target that writes the file rather than relying on shell redirection).
- `.github/workflows/ci.yml` — add a CI guard: run `make docs-tools` then `git diff --exit-code web/data/tools.json` so a PR cannot land if the generator output is stale.
- `taskplane-tasks/TP-051-tool-catalog-generator/STATUS.md`.
- `CHANGELOG.md` under `[Unreleased]`.

Out of scope:

- Implementing the analyzer family — only surface the PRD-vs-registry divergence in `STATUS.md`.
- Generating MCP resource / prompt catalogs (separate follow-up if needed).
- Modifying any tool's actual behaviour or schema.
- Migrating the rest of the end-user docs — TP-052.

## Steps

### Step 1: Design `ToolDescriptor`

- [ ] Decide the exact fields and JSON shape. Recommended:
  ```json
  {
    "name": "get_activities",
    "group": "activities",
    "tier": "core",
    "safety": "read",
    "summary": "Lists activities for a date range with pagination and terse rows by default.",
    "anchor": "get_activities"
  }
  ```
- [ ] Stable ordering: sort by `group` (alphabetical) then `name` (alphabetical), then write with two-space indent. Determinism is required for the CI guard.
- [ ] Source of the `summary` field: pull from the registered tool's MCP description, **trimming** to the first sentence so the website renders consistently. Long-form description stays in the MCP description (which the LLM reads); the website shows the one-liner.

### Step 2: Wire `Catalog()` into the registry

- [ ] Add `internal/tools/catalog.go` exporting `func Catalog() []ToolDescriptor`. Implementation must enumerate every registration site without requiring `ICUVISOR_DELETE_MODE=full` or `ICUVISOR_TOOLSET=full` — i.e., bypass the safety/toolset gate and instead annotate each descriptor with its `tier` and `safety` so the website can show every tool, with badges indicating gating.
- [ ] Table-driven test (`catalog_test.go`):
  - Every tool name returned by `Catalog()` is non-empty, unique, snake_case.
  - Every tool name documented in PRD §7.2.C that is also _registered_ appears in `Catalog()`. (Tools listed in the PRD but not yet registered — e.g., analyzers — must **not** appear; the test asserts that and the divergence is logged.)
  - Every tool registered in `internal/tools/registry.go` (any tier, any safety) has a `Catalog()` entry.

### Step 3: Generator binary

- [ ] `cmd/gendocs/main.go`: parse flags `--out web/data/tools.json` (default), call `tools.Catalog()`, marshal with stable encoding, write atomically.
- [ ] Unit test that runs `gendocs` into a tempdir and compares against a golden file in `cmd/gendocs/testdata/`.

### Step 4: Hugo rendering

- [ ] `web/content/reference/tools.md` — Hextra page, frontmatter `title: "Tool reference"`, body: short intro paragraph + `{{< tool-catalog >}}` shortcode call.
- [ ] `web/layouts/shortcodes/tool-catalog.html` — iterate `site.Data.tools`, group by `.group`, render a `<table>` per group. Tier badge: "core" plain, "full" with a small "advanced" badge linking to the explanation page about toolset tiers. Safety badge: "read" plain, "write" amber, "delete" red.
- [ ] Landing-page chips: pick a short featured list (5–8 names) and render them through the same data source so they cannot drift. Either by a curated `featured: true` flag on selected descriptors, or by listing names in `hugo.toml` params.

### Step 5: Makefile + CI guard

- [ ] `make docs-tools` — runs `go run ./cmd/gendocs --out web/data/tools.json`. Idempotent.
- [ ] `make help` — list the new target.
- [ ] CI: a job step that runs `make docs-tools` then `git diff --exit-code web/data/tools.json`. If a contributor adds or removes a tool without regenerating, CI fails with a clear "run `make docs-tools` and commit the result" message.

### Step 6: Reconcile documentation conflicts (surface, do not silently fix)

- [ ] In `STATUS.md`, log the three known doc conflicts that overlap this refactor (the executing agent must verify them against the code at the time of execution; do not trust this prompt blindly):
  - **Analyzer-family ghost:** `docs/prd/PRD-icuvisor.md` §7.2.C describes `analyze_*` / `compute_*` / `get_fitness_projection` and a "~39 tools at v1.0" target; `internal/tools/registry.go` registers zero analyzers; `README.md` lists ~36 tools; `ROADMAP.md` does not mention an analyzer phase. **Outcome of this task:** generator emits only registered tools; `STATUS.md` records the divergence; file a follow-up TP titled "Reconcile analyzer family — defer in PRD or add roadmap phase".
  - **`get_planning_parameters` contradiction:** `ROADMAP.md` lists this tool as both checked-off (line ~22) and deferred (line ~29). **Outcome of this task:** generator behaviour reflects registry truth (registered or not); the ROADMAP edit itself is TP-054 / out of scope here — but `STATUS.md` must call it out.
  - **`update_wellness` error contract:** `docs/prd/PRD-icuvisor.md` line ~247 specifies `field_not_writable: sleepScore (device-managed)` as part of the error contract; `README.md` does not surface this. **Outcome of this task:** the `update_wellness` descriptor's `summary` must include the device-managed-field rejection in one clause (e.g., "rejects device-owned `sleepScore`/`_native` fields"), so the website's tool reference shows it. The full error contract documentation lands on the website in TP-052; this task only ensures the one-liner is faithful.

### Step 7: Replace duplicate prose

- [ ] Remove the hand-written tool list from `README.md` §"MCP tool catalog" — replace with one paragraph + a link to `https://icuvisor.app/reference/tools/`. (TP-054 owns the broader README slim-down; this task owns the deletion of the tool list specifically, since the generator now produces the SoT.)
- [ ] Remove the hand-coded `<span class="tool">` chips in `web/layouts/index.html` — render from data instead.

### Step 8: Verify

- [ ] `make docs-tools` produces a clean diff on a fresh checkout.
- [ ] `make build`, `make test`, `make test-race`, `make lint` all green.
- [ ] CI guard fails when a tool is renamed or added without regeneration (test by adding a throwaway tool locally, running CI guard, then reverting).
- [ ] `cd web && hugo --minify --gc` produces a `reference/tools/` page that lists every tool currently registered and matches `web/data/tools.json`.
- [ ] Manual: count of tools shown on the website matches `len(tools.Catalog())` matches `wc -l < web/data/tools.json`'s entry count.

## Acceptance Criteria

- `tools.Catalog()` returns every registered tool with stable metadata (name, group, tier, safety, summary, anchor).
- `make docs-tools` writes `web/data/tools.json` deterministically; CI fails when the committed file is stale.
- The website's `reference/tools/` page is generated from `web/data/tools.json` and visually groups tools by domain with tier/safety badges.
- The landing page's tool chips are rendered from the same data source — no possibility of drift between landing and reference.
- `README.md` no longer hand-maintains the tool list; it points to `https://icuvisor.app/reference/tools/`.
- `STATUS.md` records the three doc-conflict findings (analyzer family, `get_planning_parameters`, `update_wellness` error contract) with concrete pointers for follow-up.
- All tests pass, including `go test -race`.

## Do NOT

- Do not implement the analyzer family in this task. Only surface the divergence.
- Do not change any registered tool's MCP description, schema, or behaviour to make generation easier — adapt the generator, not the tool.
- Do not write the catalog by hand into JSON. The whole point is that code is the source.
- Do not skip the CI drift guard. The motivation for this task is to prevent the drift that currently exists.
- Do not strip the `_meta`, `next_page_token`, or any per-tool detail from MCP descriptions to fit the website. The website shows summaries; full MCP descriptions stay in code.
- Do not couple the generator to a running MCP server, a network call, or a live intervals.icu account.

## Documentation

Must update:

- `STATUS.md` — including the three doc-conflict findings.
- `README.md` — replace tool list with link to website.
- `CHANGELOG.md` `[Unreleased]` — under "Changed".
- `web/README.md` — note that `web/data/tools.json` is generated; do not hand-edit.

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-051`, e.g., `TP-051 add tools.Catalog() and generator binary`.

---

## Amendments

_Add amendments below this line only._
