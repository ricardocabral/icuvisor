# Plan Review — TP-237 Step 2

## Verdict: APPROVE

The Step 2 plan is implementation-ready when read with the approved Step 1 contract and amendment. Its ownership is narrow and coherent: add `CoachingHandoffPrompt()` to the static registry, update the registry count/name ordering and shared prompt expectations, add the prompt to the golden table, and commit the deterministic `testdata/coaching_handoff.md` rendering. The existing focused tests already cover bounded/default arguments, source separation, athlete-local anchoring, privacy exclusions, the read-only tool list, and the advanced-capability fallback, so those do not need to be recreated in this step.

While updating `catalog_test.go`, include the new prompt in every applicable hard-coded cross-prompt list, including client-pack linkage and terse resource-citation coverage, so registration and the Step 1 portable pack cannot drift independently. Then run `go test ./internal/prompts -run 'Prompt|CoachingHandoff'` as planned.
