# TP-024 — Custom items create/update (`create_custom_item`, `update_custom_item`)

## Mission

Ship the custom-item write path. Creates and updates are ungated; delete is handled in the destructive-cluster task (TP-025).

Roadmap items (ROADMAP.md v0.3):

- Custom-item create/update.

PRD anchors: §7.2.C custom-items catalog.

Complexity: Blast radius 1, Pattern novelty 1, Security 2, Reversibility 2 = 6 → Review Level 1. Size: S.

## Dependencies

- **TP-018** — safety gate
- **TP-013** — custom-items reads (schema parity)

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C custom-items
- `ROADMAP.md` v0.3
- `internal/tools/get_custom_items*.go`, `internal/tools/get_custom_item_by_id*.go`

## File Scope

Expected files:

- `internal/tools/create_custom_item.go` + `_test.go`
- `internal/tools/update_custom_item.go` + `_test.go`
- `internal/intervals/` — typed write methods if not present
- `CHANGELOG.md`
- `README.md` catalog
- `taskplane-tasks/TP-024-custom-items-write/STATUS.md`

## Steps

### Step 1: `create_custom_item`

- [ ] Inputs match the per-schema shape (custom-items are schema-driven); inputs validated against the schema before upload
- [ ] Response is the read shape for the new item
- [ ] Tests: create per schema, schema-violation rejection at the validation layer

### Step 2: `update_custom_item`

- [ ] Inputs: `item_id` + sparse fields against the same schema
- [ ] Partial-update semantics
- [ ] Tests: update single field, schema-violation rejection

### Step 3: Verify

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual smoke against the test athlete with at least one custom-item schema

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for endpoint shapes. Do not depend on it.
- GPL/copyleft implementation code is off limits — do not read, copy, paraphrase, or transliterate.

## Acceptance Criteria

- Both tools registered in `safe` and `full`; absent in `none`.
- Schema validation happens before upload.
- No `confirm` argument.

## Do NOT

- Do not implement delete here — that is TP-025.
- Do not hardcode schemas; pull them from the read side.

## Documentation

Must update:

- `STATUS.md`
- `README.md` catalog
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-024`, for example: `TP-024 add create_custom_item tool`.

---

## Amendments

_Add amendments below this line only._
