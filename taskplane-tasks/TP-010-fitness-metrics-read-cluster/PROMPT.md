# TP-010 — Fitness, best efforts, power curves, training summary, extended metrics

## Mission

Land the second-order analytics reads that turn raw activity data into coachable signals. Treat `get_extended_metrics`'s field set as load-bearing and gated by upstream availability per PRD §7.4 #4 — drop fields the API does not expose, do not zero-fill.

Roadmap items (ROADMAP.md v0.2):

- `get_fitness` (CTL/ATL/TSB)
- `get_best_efforts`
- `get_power_curves`
- `get_training_summary`
- `get_extended_metrics` per PRD §7.2.C field set, gated by §7.4 #4

PRD anchors: §7.2.C Athlete & fitness + Activities (extended metrics), §7.4 #4.

Complexity: Blast radius 2, Pattern novelty 2, Security 1, Reversibility 1 = 6 → Review Level 2. Size: M.

## Dependencies

- **TP-007** — response shaping primitives
- **TP-008** — unit enum

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C, §7.4 #4
- `ROADMAP.md` v0.2
- `internal/response/`, `internal/units/`
- Public intervals.icu API docs for fitness, best-efforts, power-curve, training-summary, and extended-metric endpoints

## File Scope

Expected files:

- `internal/intervals/fitness.go` etc. — typed client methods
- `internal/tools/get_fitness.go`
- `internal/tools/get_best_efforts.go`
- `internal/tools/get_power_curves.go`
- `internal/tools/get_training_summary.go`
- `internal/tools/get_extended_metrics.go`
- `_test.go` for each, with `testdata/` fixtures
- `testdata/extended-metrics/availability.md` — per-field upstream availability table
- `CHANGELOG.md`
- `taskplane-tasks/TP-010-fitness-metrics-read-cluster/STATUS.md`

## Steps

### Step 1: Black-box probe extended-metric availability (§7.4 #4)

- [ ] Map each candidate field in PRD §7.2.C (`get_extended_metrics`) to whether intervals.icu exposes it via public API
- [ ] Record findings in `testdata/extended-metrics/availability.md`; cite the endpoint and a fixture for each "yes"
- [ ] Fields with no upstream exposure are **dropped from the tool**, not zero-filled

### Step 2: Implement the four straightforward reads

- [ ] `get_fitness`: date-range CTL / ATL / TSB; honour TZ; terse-by-default
- [ ] `get_best_efforts`: PRs across sports; structure by sport + duration buckets
- [ ] `get_power_curves`: mean-maximal curve; date-range arg; `include_full` for raw sample arrays
- [ ] `get_training_summary`: aggregated volume / TSS / zones over a date range

### Step 3: Implement `get_extended_metrics`

- [ ] Expose only the fields confirmed in Step 1
- [ ] Use the canonical unit enum (TP-008) and the shaping pipeline (TP-007)
- [ ] Heavy payloads (raw stream-derived series) gated behind `include_full`

### Step 4: Tests

- [ ] Table-driven tests using `httptest.Server` + fixtures; never hit the network
- [ ] Cover: TZ-correct date math on fitness rows; sport-buckets on best-efforts; curve-shape correctness on power curves; field-drop behavior on extended metrics for fixtures that omit a tracked field
- [ ] `make test`, `make build`, `make lint` pass

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted. Do not depend on it.
- GPL/copyleft implementation code is off limits.

## Acceptance Criteria

- All five tools registered with the MCP server.
- `get_extended_metrics` exposes only upstream-available fields, with the availability table committed.
- `_meta.server_version` and `_meta.units` (where applicable) are present on every response.
- Tests cover Step 4 cases.

## Do NOT

- Do not zero-fill fields the API does not expose.
- Do not compute fields client-side that competitors expose server-side; the goal is fidelity, not a coaching layer.
- Do not emit raw stream samples by default on power curves.

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md`
- README catalog (add the five tools)
- `testdata/extended-metrics/availability.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-010`, for example: `TP-010 add get_fitness tool`.

---

## Amendments

_Add amendments below this line only._
