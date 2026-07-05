# Strength training and gym support upstream gap

## Current best-effort support

icuvisor does not currently expose first-class strength-training tools or structured strength-set writes. The current safe representation is calendar-level planning:

- Use `add_or_update_event` with category `NOTE` to schedule a gym time block, mobility reminder, or free-text strength plan note.
- If the athlete account has a documented upstream workout/activity `type` for the intended session, such as `WeightTraining`, use a simple `WORKOUT` event only for date/time, optional duration/load targets, name, tags, and free-text description.
- Treat `WeightTraining` as an upstream activity/event type label only. Do not infer an exercise catalog, structured set list, or strength-session schema from the type name.
- Do not encode sets, reps, external load, bodyweight, rest periods, exercise libraries, or progression rules into `workout_doc`; that DSL is for intervals.icu structured endurance workouts and target steps.

This lets a user reserve gym time and keep coach notes visible on the calendar without implying that icuvisor can round-trip structured strength data.

## Upstream gap

The product scope already treats strength-training data as conditional on upstream API support. The PRD lists strength training only as included if the intervals.icu API exposes it, and the roadmap keeps strength training endpoints in the v1.x bucket behind the same assumption. The repository-held OpenAPI baseline currently shows `WeightTraining` as an activity/event sport type value, but not as a first-class strength-session schema with exercises, sets, reps, loads, rest, or dedicated strength endpoints.

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

## Implementation criteria

First-class strength/gym tools should not be added until the evidence above can answer each contract question below:

- **Endpoint contract:** documented or black-box verified read/write endpoints, authentication scope, pagination model, stable identifiers, and representative success/error payloads.
- **Schema contract:** typed fields for exercise, set, rep, load, bodyweight, rest, effort, side/limb, superset/circuit grouping, notes, planned-versus-completed state, and units/scale labels.
- **Response contract:** a terse default shape that can summarize one session and a list of sessions without flooding the model context, plus an explicit `include_full: true` expansion for full set-level payloads.
- **Write-safety contract:** idempotency or retry semantics for create/update, partial-update behavior, conflict handling, and destructive-operation behavior that can be gated by icuvisor's existing safety policy.
- **Round-trip contract:** proof that data written by icuvisor is returned by the upstream API without silently losing structured sets, loads, notes, or calendar linkage.

Until that evidence exists, docs and prompts should steer assistants to schedule gym blocks as notes or simple supported calendar events and explicitly avoid inventing structured strength-set support.
