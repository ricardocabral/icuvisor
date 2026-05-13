# TP-027 — `input_examples` on complex write tools

## Mission

Add `input_examples` to the JSON Schema of every complex v0.3 write tool so the LLM can pattern-match instead of inferring. Targets: `add_or_update_event`, `create_workout`, `update_workout`, `create_custom_item`, `update_custom_item`, `apply_training_plan`, `update_wellness`, `update_sport_settings`.

Roadmap items (ROADMAP.md v0.3):

- `input_examples` on complex write tools.

PRD anchors: §7.2.D LLM-readable schemas.

Complexity: Blast radius 1, Pattern novelty 1, Security 1, Reversibility 1 = 4 → Review Level 1. Size: S.

## Dependencies

- **TP-020** — event write cluster
- **TP-021** — wellness write
- **TP-022** — sport-settings write
- **TP-023** — workout-library CRUD
- **TP-024** — custom-items write
- **TP-026** — `apply_training_plan`

## Context to Read First

- `CLAUDE.md` (schema-description requirements)
- `docs/prd/PRD-icuvisor.md` §7.2.D
- `ROADMAP.md` v0.3
- All v0.3 write tools landed above

## File Scope

Expected files:

- Each tool's `.go` file gains `input_examples` on its schema
- `internal/tools/*_test.go` — at least one test that asserts every targeted tool exposes a non-empty `input_examples` array
- `CHANGELOG.md`
- `taskplane-tasks/TP-027-input-examples-on-write-tools/STATUS.md`

## Steps

### Step 1: Curate examples

- [ ] Per tool, write 2–3 examples covering the common shapes (minimal, with `workout_doc` where relevant, with optional fields populated)
- [ ] Examples must validate against the tool's own schema — add a unit test that enforces this
- [ ] Examples must not contain real athlete data; use plausible fixtures only

### Step 2: Wire into schema

- [ ] Where the MCP SDK supports `examples` (JSON Schema 2020-12 `examples` keyword), use it directly
- [ ] If the SDK surfaces a separate `input_examples` field, mirror the examples there
- [ ] Document the chosen convention in `STATUS.md`

### Step 3: Catalog-wide invariant test

- [ ] One test that iterates all v0.3 write tools and asserts non-empty examples
- [ ] Test fails when a new write tool lands without examples — this guards the convention going forward

### Step 4: Verify

- [ ] `make test`, `make build`, `make lint`

## Acceptance Criteria

- Every targeted tool has at least 2 examples.
- Examples validate against their tool's schema (enforced by test).
- Catalog-wide invariant test prevents regressions.

## Do NOT

- Do not include real athlete IDs, API keys, or wellness measurements in examples.
- Do not let examples drift from current schemas — the validation test is the contract.

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-027`, for example: `TP-027 add input_examples to add_or_update_event`.

---

## Amendments

_Add amendments below this line only._
