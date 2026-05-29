# Strength training and gym support upstream gap

## Current best-effort support

icuvisor does not currently expose first-class strength-training tools or structured strength-set writes. The current safe representation is calendar-level planning:

- Use `add_or_update_event` with category `NOTE` to schedule a gym time block, mobility reminder, or free-text strength plan note.
- If the athlete account has a documented upstream workout/activity `type` for the intended session, use a simple `WORKOUT` event only for duration, name, tags, and free-text description.
- Do not encode sets, reps, load, rest periods, exercise libraries, or progression rules into `workout_doc`; that DSL is for intervals.icu structured endurance workouts and target steps.

This lets a user reserve gym time and keep coach notes visible on the calendar without implying that icuvisor can round-trip structured strength data.

## Upstream gap

The product scope already treats strength-training data as conditional on upstream API support. The PRD lists strength training only as included if the intervals.icu API exposes it, and the roadmap keeps strength training endpoints in the v1.x bucket behind the same assumption.

Open questions before implementation:

- Which upstream endpoint reads strength sessions, templates, or exercises?
- Which endpoint writes them, and is the write idempotent or safe to retry?
- What schema represents exercises, sets, reps, external load, bodyweight, rest, RPE/RIR, sides, supersets/circuits, and notes?
- Does the calendar expose strength as a `WORKOUT` type, a separate event category, custom items, or another resource?
- What response shape identifies completed strength work versus planned gym notes?
- Which fields are device-owned or computed upstream and must be read-only?

## Evidence required for first-class tools

Before adding strength-training MCP tools, collect black-box or public API evidence for:

1. Read endpoint paths, required query parameters, pagination, and example responses.
2. Write endpoint paths, required fields, partial-update behavior, idempotency semantics, and error shapes.
3. Calendar integration: how planned strength appears in `get_events` and how completed sessions appear in `get_activities` or any dedicated strength endpoint.
4. Supported schema fields and units, including whether exercise names are free text or selected from an upstream catalog.
5. Safe-delete/update behavior so destructive operations can be registered behind icuvisor's existing capability gates.
6. Terse response shape that summarizes the session without dumping large exercise/set payloads unless `include_full: true` is requested.

Until that evidence exists, docs and prompts should steer assistants to schedule gym blocks as notes or simple supported calendar events and explicitly avoid inventing structured strength-set support.
