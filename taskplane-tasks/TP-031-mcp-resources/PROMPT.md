# TP-031 — MCP Resources for long-form schema content

## Mission

Move long-form, slow-changing content out of the per-session tool-description budget (KR5) by shipping it as MCP Resources. Four resources: the workout DSL syntax, the event-category enum, the per-`item_type` custom-item `content` schemas, and the auto-refreshing athlete profile.

Roadmap items (ROADMAP.md v0.4):

- MCP Resources: `icuvisor://workout-syntax`, `icuvisor://event-categories`, `icuvisor://custom-item-schemas`, `icuvisor://athlete-profile`. Long-form schema content moves out of inline tool descriptions.

PRD anchors: §7.2.G MCP Resources and Prompts, §7.2.C catalog (custom-item §7.2.C note already points `content` docs at `icuvisor://custom-item-schemas`), KR5 (§6), assumption §7.4 #13 (clients honoring Resources).

Complexity: Blast radius 2 (tool descriptions trimmed to point at resources), Pattern novelty 3 (first MCP Resources in the server), Security 1, Reversibility 2 = 8 → Review Level 2. Size: M.

## Dependencies

- **TP-019** — `workout_doc` serializer / `internal/workoutdoc`; the workout-syntax resource documents the same DSL grammar (single source of truth — derive, don't re-author).
- **TP-013** — workout-library + custom-items reads; the custom-item-schemas resource documents the `content` shapes the read side already surfaces.
- **TP-012** — events reads; the event-categories resource documents the same upstream category enum.
- **TP-004** — `get_athlete_profile`; the athlete-profile resource serves the same shaped profile.

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C, §7.2.G, §6 KR5, §7.4 #13
- `ROADMAP.md` v0.4
- `internal/mcp/` — server wiring; check the Go SDK's `resources/list` + `resources/read` support
- `internal/workoutdoc/` — DSL grammar source for `workout-syntax`
- `internal/tools/` — custom-item reads, event reads, athlete-profile tool
- Go SDK docs for MCP Resources (record the canonical link in `STATUS.md`)

## File Scope

Expected files:

- `internal/mcp/resources.go` (or `internal/resources/`) — resource registration + handlers for `resources/list` and `resources/read`
- `internal/resources/*_test.go`
- `internal/resources/testdata/` — golden content for the three static resources
- `internal/tools/` — trim inline tool descriptions that currently carry long-form schema prose; replace with a one-line pointer to the resource URI
- `README.md` — document the four resource URIs
- `CHANGELOG.md`
- `taskplane-tasks/TP-031-mcp-resources/STATUS.md`

## Steps

### Step 1: Resource registration plumbing

- [ ] Wire `resources/list` and `resources/read` into the MCP server via the Go SDK
- [ ] Define a small internal interface so each resource is one greppable registration, mirroring the tool registry pattern
- [ ] Decide static vs dynamic per resource; document in `STATUS.md`

### Step 2: `icuvisor://workout-syntax`

- [ ] Content derived from the `internal/workoutdoc` grammar — do not hand-author a second copy that can drift
- [ ] Covers every step/target type the serializer supports; a test asserts coverage parity with `workoutdoc`

### Step 3: `icuvisor://event-categories`

- [ ] Full event-category enum with one-line descriptions, sourced from the same enum the event tools use
- [ ] Static content; golden-file locked

### Step 4: `icuvisor://custom-item-schemas`

- [ ] Per-`item_type` schema for the `content` field (chart/field/stream/panel/zones)
- [ ] Reuses the schema samples the custom-item reads/writes already validate against — single source of truth
- [ ] Golden-file locked

### Step 5: `icuvisor://athlete-profile`

- [ ] Serves the shaped athlete profile; auto-refreshing (define and document the refresh/staleness policy in `STATUS.md` — reuse the client + cache conventions, no unbounded calls)
- [ ] Honors the same unit/timezone/`_meta` shaping as `get_athlete_profile`

### Step 6: Trim inline tool descriptions

- [ ] Remove long-form schema prose from tool descriptions now covered by a resource; replace with a one-line `see icuvisor://...` pointer
- [ ] Confirm the schema-stability CI guard (TP-015) still passes — description trims must not break confusability/first-sentence checks
- [ ] README: document the four resource URIs and what each carries

### Step 7: Verify

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual: `resources/list` shows four entries; `resources/read` returns each; confirm at least one MCP client renders them (note any client that ignores `resources/list` in `STATUS.md` per §7.4 #13)

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) ships Resources; may be consulted for URI ergonomics only. Do not copy.
- `mvilanova/intervals-mcp-server` is GPLv3 — do not read, copy, paraphrase, or transliterate.

## Acceptance Criteria

- Four resources registered and readable via `resources/list` / `resources/read`.
- `workout-syntax` and `custom-item-schemas` derive from existing single sources (no drift-prone duplicates); coverage parity asserted by tests.
- `athlete-profile` auto-refreshes with a documented policy and no unbounded upstream calls.
- Inline tool descriptions trimmed; schema-stability CI guard still green.
- README, CHANGELOG updated.

## Do NOT

- Do not hand-author a second copy of the workout DSL grammar or custom-item schemas — derive from the existing source.
- Do not let `athlete-profile` refresh hammer the upstream API; cache with a documented TTL.
- Do not break the TP-015 schema-stability / confusability CI guards when trimming descriptions.

## Documentation

Must update:

- `STATUS.md`
- `README.md`
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-031`, for example: `TP-031 wire MCP resource registration plumbing`.

---

## Amendments

_Add amendments below this line only._
