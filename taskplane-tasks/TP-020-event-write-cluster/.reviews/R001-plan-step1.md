# Plan Review: TP-020 Step 1 (`add_or_update_event`)

**Verdict: Changes requested**

I read `PROMPT.md`, `STATUS.md`, the PRD event-write requirements, roadmap v0.3, and the existing event/workoutdoc/client patterns. I do not see an implementation plan beyond the generated checklist in `STATUS.md`, so this cannot be approved as-is. Before coding, please record a concrete plan that resolves the points below.

## Blocking gaps to address in the plan

1. **Input contract drift: `workout_doc` vs PRD `steps[]`**
   - `PROMPT.md` says optional `workout_doc`, but PRD §7.2.C specifies the tool input accepts either structured `steps[]` or raw `description`, not both.
   - The plan must choose and document the public schema intentionally. If using `workout_doc`, ensure it still maps to the TP-019 `workoutdoc.WorkoutDoc`/`steps` model and is not confused with the upstream `workout_doc` field that must never be sent on write.
   - Add validation that structured workout input and free-text `description` are mutually exclusive.

2. **Write payload shape and endpoint semantics are unspecified**
   - The plan needs to identify the exact `internal/intervals` methods and HTTP methods/paths for create vs update, including how `event_id` selects update without creating/replacing an unrelated event.
   - It should define a typed request payload that preserves `description` byte-for-byte and includes supported upstream fields (`date`, `category`, `name`, `tags`, planned load/metrics) without using `map[string]any` as the primary contract.
   - Do not log raw descriptions or serialized workout DSL.

3. **`workout_doc` upload asymmetry must be explicit**
   - The implementation must serialize structured workout steps with `internal/workoutdoc.Serialize` and submit only the resulting DSL string in the upstream `description` field.
   - The plan should include tests proving no structured `workout_doc` key is sent to the fake upstream client.
   - Unsupported/lossy serialization errors should become short actionable user errors; if lossy fields are possible, surface `_meta.lossy_fields`/warnings rather than silently dropping them.

4. **Round-trip parity with read shape is not planned concretely**
   - The response should reuse the existing event shaping (`eventRow` / `get_event_by_id`-compatible shape) instead of introducing a divergent write confirmation shape.
   - The plan should say whether the write endpoint returns a full event or whether the tool will re-fetch by ID after create/update to guarantee parity.
   - Include default terse behavior and optional raw payload behavior consistently with existing read tools.

5. **Safety gate registration is missing**
   - `add_or_update_event` must be `RequirementWrite`, registered in `safe` and `full`, and absent in `none` via the existing registration-time capability filter. It must not have any `confirm` argument.
   - Add registry tests for safe/full/none behavior, or extend existing protocol capability tests to cover the real tool.

6. **Schema/examples/snapshots are missing from the plan**
   - Every argument needs an LLM-readable description with date/timezone and unit semantics.
   - PRD §7.2.E requires `input_examples` for `add_or_update_event`; the plan should cover how this project represents examples in schemas/tool metadata.
   - If schema snapshots are expected for newly added tools, include updating/adding the snapshot and running the schema stability check.

7. **HTTP write helper behavior must be deliberate**
   - Existing `internal/intervals` helpers are GET-oriented and retry 429/5xx. The plan should specify how POST/PUT bodies are encoded, content type is set, response bodies are closed, and retries are limited to safe/idempotent operations where appropriate.
   - Creating events is not safely retryable unless upstream provides idempotency; avoid hidden duplicate event creation on retry.

8. **Test coverage needs more specificity**
   - Add table-driven tests for: create, update, unknown fields rejected, missing required fields, `description` preserved byte-for-byte, description+structured-workout mutual exclusion, structured serialization via TP-019 golden fixtures, tags preserved, planned metrics serialized, response parity, and no raw/structured `workout_doc` upload.
   - Include athlete timezone handling for the `date` input and normalized metadata in the response.

## Suggested plan shape

A revised plan should list the files to edit and the key decisions:

- `internal/intervals/events.go`: add typed create/update event payload and client methods; add write-specific HTTP JSON helper.
- `internal/tools/add_or_update_event.go`: typed request/response, strict validation, workoutdoc serialization, sanitized user errors, read-shape response reuse.
- `internal/tools/registry.go`: register only when the client implements the write interface; mark `RequirementWrite`.
- Tests in `internal/intervals` and `internal/tools`, plus README/CHANGELOG/schema catalog updates.

Once those decisions are documented, the plan should be reviewable/approvable.
