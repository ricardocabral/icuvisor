# Plan Review — Step 2: Add eval/adversarial coverage

**Verdict: APPROVE.**

The revised Step 2 plan now pins the calendar-event edit path, separates edit-in-place adversarial expectations from the safe-mode surrender corpus, and names the advanced-capabilities guidance surface instead of trying to test safe-mode unavailability through delete handlers.

## Implementation notes

- In the new cookbook scenario, make the prompt and scenario fields clear enough for the judge to inspect arguments: `add_or_update_event` should use the existing `event_id`; omitting `event_id` should be listed as an anti-pattern or covered in `must_address`/prompt wording.
- Put the short/actionable/server-config-only guidance assertion in the existing `internal/tools/list_advanced_capabilities_test.go` surface, while using `internal/safety/adversarial_test.go` only for registration-time absence.
- The planned verification commands are sufficient for this step: `make eval-validate`, `go test ./internal/tools`, and `go test ./internal/safety`.
