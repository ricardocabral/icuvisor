# TP-020 — Event write cluster (`add_or_update_event`, `link_activity_to_event`, `add_activity_message`)

## Mission

Land the non-destructive event-side writes: planned workouts and races (`add_or_update_event`), manual pairing of activities to planned events when auto-pair misses (`link_activity_to_event`, upstream public discussion), and free-text activity annotations (`add_activity_message`).

Roadmap items (ROADMAP.md v0.3):

- `add_or_update_event` (free-text `description` preserved verbatim, `workout_doc` for structured steps, `tags` supported).
- `link_activity_to_event` — manual pairing for compliance scoring.
- `add_activity_message`.

PRD anchors: §7.2.C event/activity catalog, §7.4 #14 compliance scoring, §7.2.D safety model (writes ungated; only deletes gated).

Complexity: Blast radius 2, Pattern novelty 2, Security 2 (mutates the athlete's calendar), Reversibility 2 = 8 → Review Level 2. Size: M.

## Dependencies

- **TP-018** — safety gate (`CanWrite`)
- **TP-019** — `workout_doc` serializer (events may carry structured steps)
- **TP-012** — events reads (round-trip parity)

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C, §7.4
- `ROADMAP.md` v0.3
- `internal/tools/get_events*.go`, `internal/tools/get_event_by_id*.go`
- `internal/workoutdoc/`

## File Scope

Expected files:

- `internal/tools/add_or_update_event.go` + `_test.go`
- `internal/tools/link_activity_to_event.go` + `_test.go`
- `internal/tools/add_activity_message.go` + `_test.go`
- `internal/intervals/` — typed write methods if not yet present
- `CHANGELOG.md`
- `README.md` catalog
- `taskplane-tasks/TP-020-event-write-cluster/STATUS.md`

## Steps

### Step 1: `add_or_update_event`

- [ ] Inputs: `date` (athlete-TZ), optional `event_id` (update vs create), `category`, `name`, `description` (free-text, preserved verbatim), `workout_doc` (structured; serialized via TP-019), `tags[]`, `target_load` / planned metrics where supported upstream
- [ ] Server emits the DSL string for `workout_doc` on upload; never sends the structured form to intervals.icu
- [ ] Response shape matches the read shape for the same event ID (round-trip parity)
- [ ] Tests: create, update, free-text round-trip, workout_doc round-trip via the TP-019 golden fixtures, tag preservation

### Step 2: `link_activity_to_event`

- [ ] Inputs: `activity_id`, `event_id`; both normalized via existing helpers
- [ ] Documents that this is the manual escape hatch when auto-pair misses (upstream public discussion); include this in the tool description
- [ ] Tests: success path, mismatched-date warning surfaced in response `_meta.warnings`, idempotent re-link

### Step 3: `add_activity_message`

- [ ] Inputs: `activity_id`, `message` (free text), optional `private` flag if upstream supports it
- [ ] Append-only; never overwrites prior messages
- [ ] Tests: success, empty-message rejection, athlete-ID normalization

### Step 4: Schema descriptions

- [ ] Every arg has an LLM-readable description with units / scale where relevant
- [ ] Tool descriptions explicitly state non-destructive intent so the LLM does not seek a `confirm` arg (there is none — TP-018)

### Step 5: Verify

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual smoke against the maintainer's test athlete account: create, update, link, message

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for endpoint shapes. Do not depend on it.
- GPL/copyleft implementation code is off limits — do not read, copy, paraphrase, or transliterate.

## Acceptance Criteria

- The three tools are registered in `safe` and `full` modes, absent in `none`.
- No tool accepts a `confirm` argument.
- `add_or_update_event` preserves free-text `description` byte-for-byte and round-trips `workout_doc` via the TP-019 serializer.
- `link_activity_to_event` succeeds against the test athlete and produces a corrected compliance reading on the next `get_events` call.
- Tests cover the cases in Steps 1–3.

## Do NOT

- Do not delete or replace existing events when "updating" without an explicit `event_id`.
- Do not send the structured `workout_doc` to intervals.icu; only the DSL string.
- Do not dogfood against production athletes; v0.3 dogfood is TP-029 and uses a dedicated test account.
- Do not log raw `description` / `message` bodies — they may contain personal context.

## Documentation

Must update:

- `STATUS.md`
- `README.md` catalog
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-020`, for example: `TP-020 add add_or_update_event tool`.

---

## Amendments

_Add amendments below this line only._
