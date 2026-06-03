# Plan Review: Step 1 — Audit repeat serialization and validation

Result: Approved with one required audit emphasis.

The Step 1 plan is appropriately scoped for an audit-only step: it reads the relevant WorkoutDoc serializer/parser/validator tests, checks write-tool coverage, records gaps in `STATUS.md`, and runs the targeted packages. I also ran `go test ./internal/workoutdoc ./internal/tools`; both packages pass currently.

Required emphasis before marking Step 1 complete:
- Audit the implementation paths as well as tests, especially `repeatLineRE` in `parse.go` and `repeatHeaderRE` / `IsStructuredStepLine` / `ValidateDescription` in `validate.go`. Current behavior worth recording is that canonical repeat headers are recognized, `- 3x` is treated as a malformed structured step, but `-3 x` may be treated as prose unless classifier logic changes. This discovery directly affects Step 2's "rejecting or warning when appropriate" requirement.

Suggested discovery items to capture if confirmed:
- Existing golden fixture `02-repeat-recovery-*` already covers described repeat header serialization as `Main Set 3x` through WorkoutDoc and create/event write paths.
- Missing explicit bare repeat-header assertion for `3x`.
- Missing explicit negative coverage for malformed repeat-like lines (`-3 x`, `- 3x`).

No broader scope changes are needed for Step 1.
