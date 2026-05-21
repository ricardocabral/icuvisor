# Plan Review — Step 2: Mechanical replacement across tool files

## Verdict

CHANGES REQUESTED before coding Step 2.

The revised status incorporates most of R004: it calls out the package-local `decodeStrict`, raw-field prechecks, checked `json.Marshal` sites, and targeted greps/tests. One important behavior-preservation gap remains around empty arguments for the existing `decodeStrict` callers.

## Blocking finding

### Preserve existing empty-input behavior for old `decodeStrict` callers, or document the intentional change

`DecodeStrict[T]` returns the zero value for empty/whitespace input. The existing package-local `decodeStrict(raw, &args)` in `get_fitness.go` does not: it returns `arguments must be a JSON object` for empty/whitespace and non-object input.

The current Step 2 plan says both:

- replace package-local `decodeStrict(raw, &args)` callers with `DecodeStrict[T](raw)`, and
- preserve bespoke empty-input/raw-field validation ordering.

That still leaves the common old-helper callers ambiguous. A direct mechanical replacement will silently change user-facing validation errors for many tools. Examples:

- `decodeDeleteEventRequest` would move from `arguments must be a JSON object` on empty input to downstream `event_id is required`.
- `decodeDateRangeRequest` would move from `arguments must be a JSON object` to `start_date and end_date must be YYYY-MM-DD`.
- Similar shifts can occur across write/delete/read decoders currently relying on the package-local helper.

Because the prompt explicitly says not to change user-visible error-message wording, the plan should add a concrete empty-input migration rule before implementation. For example:

1. audit all current `decodeStrict(raw, &args)` callers;
2. for callers whose current empty-input behavior must stay `arguments must be a JSON object`, add a minimal precheck before `DecodeStrict[T]`; and
3. only preserve existing explicit empty-allowed wrappers (`get_custom_items`, `get_training_plan`, `get_workout_library`, etc.) as they are today.

If the task owner intentionally accepts normalizing empty input to zero-value validation for all these tools, record that exception in `STATUS.md` first, because it conflicts with the task's "Do NOT change user-visible error-message wording" guardrail.

## Notes

- The `TextResult` replacement rule is now close enough: keep it to ignored-marshal-error sites or shaped values that are JSON-marshalable by construction. `encodeShaped` is a reasonable candidate only because `response.Shape` already JSON round-trips the payload before returning `shaped`.
- Keep the R004 raw-field ordering guardrails: `rawObjectFields`, `rawObjectHasField`, and `rejectReadOnlyWellnessFields` should still run before strict decoding where they do today.
- Before moving to Step 3, the proposed `go test ./internal/tools` plus greps for `DisallowUnknownFields`, `decodeStrict(`, and `ContentTypeText` are the right targeted verification.
