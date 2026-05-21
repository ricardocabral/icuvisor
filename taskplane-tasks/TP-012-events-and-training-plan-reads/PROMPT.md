# TP-012 — Events reads with `get_event_by_id` upstream-inconsistency handling, plus `get_training_plan`

## Mission

Ship the calendar / training-plan reads, with deliberate handling for the upstream 404 inconsistency between `get_events` (list) and `get_event_by_id` (detail) so the LLM does not get stuck in a retry loop on IDs the list endpoint just returned.

Roadmap items (ROADMAP.md v0.2):

- `get_events`, `get_event_by_id`
- `get_event_by_id` upstream-inconsistency handling: structured `unavailable: { reason: "upstream_inconsistency" }` when the detail endpoint 404s on an ID `get_events` just listed.
- `get_training_plan`

PRD anchors: §7.2.C Events & workouts (read side only), §7.4 #16.

Complexity: Blast radius 2, Pattern novelty 2, Security 1, Reversibility 1 = 6 → Review Level 2. Size: M.

## Dependencies

- **TP-007** — response shaping primitives

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C Events & workouts (read side), §7.4 #16
- `ROADMAP.md` v0.2
- Upstream issues upstream reports #79 and #106 (read the PRD summary; do not read the GPL source)
- Public intervals.icu API docs for events and training-plan endpoints

## File Scope

Expected files:

- `internal/intervals/events.go` — list + detail client methods
- `internal/intervals/training_plan.go`
- `internal/tools/get_events.go`
- `internal/tools/get_event_by_id.go`
- `internal/tools/get_training_plan.go`
- `_test.go` for each
- `internal/intervals/testdata/events/inconsistent/` — fixtures reproducing the list→detail 404 mismatch (synthetic, plus any real reproducers captured during development)
- `CHANGELOG.md`
- `taskplane-tasks/TP-012-events-and-training-plan-reads/STATUS.md`

## Steps

### Step 1: Implement `get_events` and `get_training_plan`

- [ ] `get_events`: date-range list; terse-by-default; TZ rendering; `include_full` for raw payload
- [ ] `get_training_plan`: fetch the active plan as exposed upstream; if not exposed, document the gap in `STATUS.md` (see §7.4 #3)
- [ ] Surface event categories using the upstream enum value; long-form category docs will move to an MCP Resource later (TP / v0.4)

### Step 2: Implement `get_event_by_id` with fallback

- [ ] Try the upstream detail endpoint first
- [ ] On 404, fall back to scanning `get_events` over the event's known date range and match by ID
- [ ] If still not found, return a structured `unavailable: { reason: "upstream_inconsistency", retried: ["detail", "list_scan"] }` rather than propagating a raw 404
- [ ] Cap the date-range scan to a sane window (e.g. ±30 days from any caller-provided hint, otherwise a documented default); record the window in `_meta.scanned_range`
- [ ] Capture every real reproducer encountered under `testdata/events/inconsistent/` with the originating list response so the root-cause investigation has a corpus

### Step 3: Tests

- [ ] Table-driven tests using `httptest.Server` + fixtures
- [ ] Cover: list / detail happy path; detail 404 → list-scan recovery; detail 404 + list-scan miss → structured `unavailable`; TZ rendering on event dates
- [ ] `make test`, `make build`, `make lint` pass

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted. Do not depend on it.
- GPL/copyleft implementation code is off limits.

## Acceptance Criteria

- Three read tools registered: `get_events`, `get_event_by_id`, `get_training_plan`.
- `get_event_by_id` returns the structured `unavailable` shape on persistent upstream inconsistency rather than a raw 404.
- Tests cover the fallback and the miss case.

## Do NOT

- Do not implement event writes (`add_or_update_event`, `delete_event`, `delete_events_by_date_range`) — those are v0.3 / TP-write tasks.
- Do not retry the upstream detail endpoint indefinitely; one attempt then list-scan.
- Do not invent an event when both detail and list-scan miss.

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md`
- README catalog (add three tools)
- `testdata/events/inconsistent/` README explaining the reproducer format if any were captured

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-012`, for example: `TP-012 add get_event_by_id fallback`.

---

## Amendments

_Add amendments below this line only._
