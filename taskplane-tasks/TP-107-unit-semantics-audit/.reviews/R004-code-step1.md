# Code Review — Step 1 (R004)

Verdict: **REVISE**

## Findings

- **Coverage gap in the supported unit matrix** (`internal/workoutdoc/workoutdoc_test.go:69-85`): the new table marks Step 1 complete, but it does not cover all currently supported target unit aliases from `workoutTargetUnits`. In particular, pace blank units are supported and default to `% Pace` (`internal/workoutdoc/syntax.go:83`), and HR percent aliases `%HR` and `HR` are supported (`internal/workoutdoc/syntax.go:80`) but have no regression cases. Add cases for those aliases or document why they are intentionally out of scope before checking this step off.

## Verification

- Ran `go test ./internal/workoutdoc` — passed.
