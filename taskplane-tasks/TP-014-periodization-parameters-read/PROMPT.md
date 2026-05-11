# TP-014 — Periodization parameters read (`get_planning_parameters`) or documented upstream gap

## Mission

Probe whether intervals.icu's public API exposes athlete-level planning parameters (ramp-rate %, recovery-week cadence, taper % drop, intensity-distribution preference). If exposed, ship `get_planning_parameters` as a terse read. If not, document the gap and file an intervals.icu feature request rather than computing them client-side.

Roadmap item (ROADMAP.md v0.2):

- Periodization parameters via `get_planning_parameters` (and write counterpart) if exposed upstream (PRD §7.4 #18) — ramp-rate %, recovery-week cadence, taper % drop, intensity distribution. Gap documented and an intervals.icu API feature request filed if absent.

PRD anchors: §7.4 #18, §7.4 #4 (per-field availability rule).

Complexity: Blast radius 1, Pattern novelty 2 (discovery may yield nothing), Security 1, Reversibility 1 = 5 → Review Level 1. Size: S.

## Dependencies

- **TP-007** — response shaping primitives

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.4 #18, §7.4 #4
- `ROADMAP.md` v0.2
- Forum thread 123739 posts #28, #30 references (via PRD)
- Public intervals.icu API docs — search athlete profile and training-plan endpoints for periodization fields

## File Scope

Expected files (one of two outcomes):

**If exposed upstream:**

- `internal/intervals/planning_parameters.go`
- `internal/tools/get_planning_parameters.go`
- `internal/tools/get_planning_parameters_test.go`
- `internal/intervals/testdata/planning/` — fixtures
- `CHANGELOG.md`
- `taskplane-tasks/TP-014-periodization-parameters-read/STATUS.md`

**If not exposed upstream:**

- `docs/upstream-gaps/periodization-parameters.md` — describes the requested fields and the API feature request filed with intervals.icu
- Reference to the filed request (Andrew Coggan's forum / support channel) in `STATUS.md`
- `taskplane-tasks/TP-014-periodization-parameters-read/STATUS.md`

The write counterpart (`update_planning_parameters`) is v0.3 / TP-write, not part of this task.

## Steps

### Step 1: Probe upstream availability

- [ ] Inspect every athlete-profile and training-plan endpoint response for fields matching ramp-rate %, recovery-week cadence, taper % drop, intensity-distribution preference
- [ ] Use `.env` credentials only if present; record only availability, never values
- [ ] Document each field's availability (yes / no / partial) in `STATUS.md`

### Step 2A: If exposed — implement the tool

- [ ] Typed read against the discovered endpoint
- [ ] Expose only the fields confirmed available; drop the rest per §7.4 #4
- [ ] Pass through TP-007 shaping (terse, `_meta.server_version`, `_meta.missing_fields` when stripping)
- [ ] Tests using `httptest.Server` + fixtures

### Step 2B: If not exposed — document the gap and file the request

- [ ] Write `docs/upstream-gaps/periodization-parameters.md` describing each requested field and its athlete-coach use case (cite forum #28, #30)
- [ ] File an intervals.icu API feature request through the public channel (Coggan's forum) and link it in the doc; do not implement client-side computation
- [ ] Note the deferral in `STATUS.md` and `ROADMAP.md` (or leave the roadmap item ticked with a footnote — coordinate with the maintainer)

### Step 3: Tests and verify

- [ ] If 2A: `make test`, `make build`, `make lint` pass with the new tool
- [ ] If 2B: no code changes; ensure docs render cleanly and `STATUS.md` is unambiguous about the outcome

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted to confirm whether the upstream exposes these fields. Do not depend on it.
- `mvilanova/intervals-mcp-server` is GPLv3 — off limits.

## Acceptance Criteria

- A clear "yes / partial / no" availability verdict per requested field is recorded in `STATUS.md`.
- If any field is exposed, `get_planning_parameters` ships with only the available fields and tests pass.
- If none are exposed, `docs/upstream-gaps/periodization-parameters.md` exists with the rationale and a link to the filed feature request.

## Do NOT

- Do not compute periodization parameters client-side from activity history (out of scope per §7.4 #18).
- Do not zero-fill missing fields.
- Do not implement the write counterpart here.

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md` if a tool ships
- README catalog if a tool ships
- `docs/upstream-gaps/periodization-parameters.md` if any field is missing

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-014`, for example: `TP-014 add get_planning_parameters` or `TP-014 document periodization upstream gap`.

---

## Amendments

_Add amendments below this line only._
