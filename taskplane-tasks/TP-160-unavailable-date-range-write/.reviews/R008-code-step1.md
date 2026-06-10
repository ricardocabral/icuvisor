# Code Review: Step 1 — Design the range-write contract

Verdict: REVISE

I ran `git diff 9594870..HEAD --name-only`, `git diff 9594870..HEAD`, `git diff --check 9594870..HEAD`, and `go test ./internal/tools -run 'Unavailable|DateRange|Event'`. `diff --check` is clean. The targeted test command still fails to compile on the intentionally missing Step 2 symbols (`newAddUnavailableDateRangeTool`, `addUnavailableDateRangeExternalID`, etc.).

## Findings

1. **The schema contract is still only checked for example presence.**
   `TestAddUnavailableDateRangeRegistrationMetadata` verifies metadata plus `examples`/`input_examples`, but it does not pin the public argument schema for the new MCP tool (`internal/tools/add_unavailable_date_range_test.go:69-74`). The Step 1 contract requires `start_date`, `end_date`, and `category` as required fields, a closed unavailability category surface, `include_full` defaulting false, and useful descriptions. An implementation could expose an incomplete or overly broad schema while all current handler tests pass. Add assertions for the schema `required` list and key property metadata/enums so the LLM-facing contract is protected before Step 2.

2. **Invalid input tests allow preflight I/O before validation.**
   The invalid-input table only asserts no writes (`internal/tools/add_unavailable_date_range_test.go:390-403`). A buggy implementation could call `ListEvents` for malformed dates, reversed/excessive ranges, or unsupported categories before returning the validation error, which would still pass. Please also assert `len(client.listCalls) == 0` for these invalid cases so rejected requests are side-effect-free and do not hit upstream unnecessarily.
