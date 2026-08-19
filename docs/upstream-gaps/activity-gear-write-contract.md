# Activity gear assignment write contract gap

Status: unsupported (2026-08-19)

The permitted repository evidence verifies activity gear on reads only:

- Activity list/detail fixtures expose a `gear_id` field, which is normalized to a
  string and resolved through `GET /api/v1/athlete/{id}/gear`.
- The documented activity write route is `PUT /api/v1/activity/{id}`.
- That route accepts the full `Activity` schema and returns an `Activity`, but the
  published schema does not identify a field-specific gear-assignment request key.
  The read-side `gear_id` field must not be inferred to be writable.
- The available contract does not define a gear-assignment response round trip,
  validation of an existing gear ID, or the semantics for an empty/null value.
  Therefore clear-versus-omit behavior is also unverified.

`update_activity` intentionally does not advertise `gear_id`, `gear`, or any
assignment/clear argument. Strict request decoding rejects those attempts before
an updater or HTTP call, including when combined with a supported sparse metadata
field. Existing read tools continue to resolve retained activity IDs through the
per-athlete gear cache; this read path is independent of writes.

Do not expose assignment until permitted evidence records the exact writable key,
request and response behavior, clear/omit semantics, and a sanitized successful
round trip using an existing gear item. Gear creation, editing, retirement, and
activity assignment remain separate follow-up work.
