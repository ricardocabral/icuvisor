# TP-015 — Tool-name disambiguation pass and CI schema-stability guards

## Mission

Audit the v0.2 read-tool catalog for confusable clusters and lock the additive-only schema rule in CI before v0.5 dogfooding exposes real users to schema churn. The goal is that an LLM, reading only the first sentence of each tool's description, can pick the right tool inside a confusable cluster (`get_activity_details` / `_intervals` / `_streams`), and that no future PR can silently remove or rename a stable tool argument.

Roadmap items (ROADMAP.md v0.2):

- Tool-name disambiguation pass on read clusters (`get_activity_details` / `_intervals` / `_streams`); CI guard for new confusable clusters.
- Tool-schema stability rules enforced in CI: additive-only on stable tools; renames/removals require a new tool name.

PRD anchors: §7.2.E (toolset tiers / disambiguating first sentences), §7.4 #7 (MCP schema caching → additive-only argument changes).

Complexity: Blast radius 2 (touches catalog + CI), Pattern novelty 2, Security 1, Reversibility 2 = 7 → Review Level 2. Size: M.

## Dependencies

- **TP-009**, **TP-010**, **TP-011**, **TP-012**, **TP-013** — catalog must be populated before audit
- **TP-014** — may or may not have added a tool; either is fine

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.E, §7.4 #7
- `ROADMAP.md` v0.2
- `internal/tools/` — the populated catalog
- Existing CI config in `.github/workflows/`

## File Scope

Expected files:

- `internal/tools/*.go` — first-sentence rewrites where ambiguity is detected
- `internal/tools/registry.go` (or equivalent) — single `Register()` call with a tool-catalog snapshot
- `internal/tools/schema_snapshot/` — checked-in JSON snapshots of each tool's argument schema, one file per tool
- `scripts/check_schema_stability.go` (or shell + jq) — CI helper that compares current schemas against snapshots
- `scripts/check_confusable_names.go` — CI helper that flags new tool clusters whose first-sentence descriptions are >X% similar
- `.github/workflows/ci.yml` — wire both checks into CI
- `CONTRIBUTING.md` — document the additive-only rule and the snapshot-update workflow
- `CHANGELOG.md`
- `taskplane-tasks/TP-015-catalog-disambiguation-and-schema-stability/STATUS.md`

## Steps

### Step 1: Audit confusable clusters

- [ ] Enumerate read-tool clusters that share a prefix or domain (`get_activity_*`, `get_wellness_*`, `get_workout_*`, `get_custom_item*`, `get_event*`)
- [ ] For each cluster, read every member's first description sentence; rewrite where the access pattern is not obvious from name alone (the §7.2.E rule)
- [ ] Record before/after first sentences in `STATUS.md`

### Step 2: Snapshot every tool's argument schema

- [ ] Generate a JSON-Schema snapshot per tool from the live registry; commit under `internal/tools/schema_snapshot/<tool_name>.json`
- [ ] Define the canonical serialization (key ordering, indentation) so diffs are reviewable
- [ ] Document the snapshot-update workflow in `CONTRIBUTING.md`: snapshots may grow (new optional arguments) but cannot shrink or rename existing arguments on a stable tool; the only way to "rename" is to ship a new tool name

### Step 3: Implement the CI schema-stability check

- [ ] On every PR, regenerate snapshots and diff against the checked-in versions
- [ ] **Additive-only** rule: new properties are allowed; removed or renamed properties fail the check
- [ ] An override mechanism exists for genuine new-tool introductions: a new tool name produces a new snapshot file with no prior version, which the check accepts
- [ ] Output the failing-tool list as an annotated CI summary, not just a non-zero exit

### Step 4: Implement the confusable-names check

- [ ] Compute a similarity metric (token Jaccard or Levenshtein on the first description sentence) within each prefix cluster
- [ ] Threshold tuned so existing v0.2 clusters pass after Step 1's rewrites; new PRs that push two tools above threshold fail with a suggested-rewrite hint
- [ ] Document the threshold in `CONTRIBUTING.md`

### Step 5: Tests and verify

- [ ] Unit tests for both CI helpers covering: clean diff (pass); added argument (pass); removed argument (fail); renamed argument (fail); new tool (pass); confusable pair (fail); rewritten pair (pass)
- [ ] Wire both checks into `.github/workflows/ci.yml`
- [ ] `make test`, `make build`, `make lint` pass
- [ ] Run the checks locally against the v0.2 catalog; commit the resulting snapshots

## Reference Implementation Policy

- Neither reference implementation has equivalent CI; nothing to consult.

## Acceptance Criteria

- Every confusable cluster in the v0.2 read catalog has a distinguishing first description sentence.
- Argument-schema snapshots are committed for every read tool.
- CI fails on a PR that removes or renames a stable tool argument.
- CI fails on a PR that introduces a confusable pair, with an actionable hint.
- `CONTRIBUTING.md` documents both rules.

## Do NOT

- Do not rewrite tool names — that is itself a breaking change. First sentences only.
- Do not exempt write tools from the additive-only rule once v0.3 lands; the snapshot directory should already cover them when that task arrives.
- Do not commit snapshots in a non-canonical serialization that produces churn diffs.

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md`
- `CONTRIBUTING.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-015`, for example: `TP-015 add schema snapshot guard`.

---

## Amendments

_Add amendments below this line only._
