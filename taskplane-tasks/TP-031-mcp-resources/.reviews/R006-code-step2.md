# Code Review R006 — Step 2: `icuvisor://workout-syntax`

**Verdict: REVISE**

## Findings

### P2 — Coverage/parity guard is self-referential and can still drift from the serializer

`internal/workoutdoc/syntax.go:31` adds a hand-maintained syntax descriptor, and `internal/resources/workout_syntax_test.go:45-83` only verifies that this same descriptor's examples serialize and appear in the Markdown. This does not satisfy the Step 2 requirement that the resource be derived from the `internal/workoutdoc` grammar and that tests assert coverage parity with the serializer.

A future change can add a supported unit/family/step form in `serialize.go` and all of these tests will still pass unless the developer remembers to update the separate descriptor. For example, adding another accepted unit branch in `formatTarget` would not affect `requiredFeatures`, the golden file, or any descriptor example. The current descriptor also only samples representative cases, not the actual supported unit/alias matrix enforced by `formatTarget`/`formatDistance`.

Please make the parity source non-self-referential. A good fix is to move the supported serializer forms/units/aliases into `workoutdoc` data that is used to drive the serializer and the docs/tests, or otherwise add an explicit serializer-supported matrix in `workoutdoc` and assert every entry has a rendered resource example/section. The resource can still render Markdown from examples via `workoutdoc.Serialize`, but the test needs to fail when serializer coverage changes without a docs update.

## Verification

- `go test ./...` passes.
- `golangci-lint run ./...` currently fails on `internal/mcp/server.go:303` (`fmt.Errorf(genericResourceErrorMessage)`); that line was not changed in this Step 2 diff, but the branch is not lint-clean as-is.
