# Plan Review — Step 2: Mechanical replacement across tool files

## Verdict

CHANGES REQUESTED before coding Step 2.

The high-level checklist is directionally right, but the Step 2 plan needs a few explicit guardrails from the Step 1 review. This step is "mechanical" only if the implementer distinguishes the existing decode/result variants; a blind replace will either miss most call sites or change user-facing error behavior unnecessarily.

## Blocking plan gaps

1. **Account for the existing package-local `decodeStrict` helper.**
   Most tool decoders already call the unexported `decodeStrict(raw, &args)` currently defined in `get_fitness.go`, rather than spelling out `DisallowUnknownFields` inline. The Step 2 plan should explicitly update those callers to `DecodeStrict[T](raw)` and remove the old `decodeStrict` helper once it has no callers. Keeping it as a wrapper would preserve duplicate boilerplate/abstraction that this task is meant to delete.

2. **Preserve bespoke empty-input/pre-decode behavior where it exists.**
   `DecodeStrict` intentionally returns the zero request for empty/whitespace input, but not every current decoder does that. The plan should call out the exceptions before doing broad edits, especially:
   - `decodeGetActivitiesRequest` currently returns `oldest is required unless next_page_token is supplied` for empty input and also needs the trimmed raw JSON for the `supplied` struct used in token validation.
   - `decodeActivityReadRequest` currently rejects empty input as `arguments must be a JSON object` before field validation.
   - Decoders with raw-field prechecks (`rawObjectFields`, `rawObjectHasField`, `rejectReadOnlyWellnessFields`) should keep their existing ordering unless a behavior change is intentionally documented.

   Without this, Step 2 can silently change LLM-facing validation messages, which the task explicitly says not to do.

3. **Define the `TextResult` replacement rule for checked `json.Marshal` sites.**
   Some exact `Result{Content: ..., StructuredContent: ...}` sites currently ignore marshal errors and are safe direct replacements. Others currently return checked errors such as `encoding get_activities response: %w`, `encoding update_wellness response: %w`, or the shared `encodeShaped` error. The plan should state which checked sites will be converted because their shaped payload is JSON-marshalable by construction, and which (if any) will be left alone. This carries forward the R003 note: do not use `TextResult` where it changes observable behavior unless the plan explicitly justifies that the error path is impossible/accepted.

## Suggested concrete Step 2 shape

- Replace package-local calls like:
  ```go
  var args someRequest
  if err := decodeStrict(raw, &args); err != nil { ... }
  ```
  with:
  ```go
  args, err := DecodeStrict[someRequest](raw)
  if err != nil { ... }
  ```
  while preserving the existing local variable names and all subsequent normalization/validation.
- Remove the old `decodeStrict` function from `get_fitness.go`; run gofmt/goimports so obsolete `bytes`, `io`, `encoding/json`, or `fmt` imports disappear where appropriate.
- For inline decoders (`get_activities.go`, `get_activity_details.go`, `get_athlete_profile.go`), replace only the strict decode portion and retain any existing special-case validation that precedes it.
- Replace exact text-result construction with `TextResult(...)`, including the shared `encodeShaped` helper if the plan accepts the no-error helper semantics there.
- Before moving to Step 3, run at least `go test ./internal/tools` plus the acceptance greps for `DisallowUnknownFields`, `decodeStrict(`, and `ContentTypeText` to catch missed mechanical sites early.

Once these guardrails are added to the Step 2 plan, the implementation should be safe to proceed.
