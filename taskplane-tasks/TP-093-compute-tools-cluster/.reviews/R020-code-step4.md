# R020 code review — Step 4: Tests and verification

Verdict: REVISE

I ran the requested/recorded targeted checks:

```sh
go test -count=1 ./internal/analysis ./internal/tools ./internal/toolcatalog ./internal/toolchecks ./internal/safety ./cmd/gendocs
make docs-tools
```

The targeted tests pass and `make docs-tools` leaves no generated-doc diff. The unused `ptrFloat64` helper from R019 has been removed.

## Blocking findings

### 1. Compliance sport/event-type filtering is still not actually covered

- Location: `internal/tools/compute_tools_test.go:384`
- Severity: Medium

`TestComputeCompliancePairingDeltasAndBreakdowns` is the only new compliance golden that supplies both `sport` and `event_type`, but every scheduled event in the fixture has `Type: "Workout"` and every activity has sport `"Run"`. As a result, the test would still pass if `compute_compliance_rate` ignored the `event_type` filter, or if auto-pairing ignored the request `sport`, because there are no non-matching candidates to exclude.

This was one of the explicit R017 requirements: when both filters are supplied, scheduled filtering must use `event_type` while auto-pairing compares completed activity sport to request `sport`. Please add a negative fixture to this test (or a separate focused test) with at least:

- a non-`Workout` scheduled event on the same date with a valid target that must not affect `scheduled_count`/series when `event_type:"Workout"` is set; and
- a same-date non-`Run` activity with a closer target match that must not be selected when `sport:"Run"` is set.

Without those negative candidates, the Step 4 checklist item for compliance `sport/event_type` filtering is marked complete but the regression it is meant to prevent is not pinned.

## Notes

`golangci-lint run ./internal/tools` still reports pre-existing implementation lint issues outside this Step 4 test diff (`unparam` and unused helpers in the compute implementations). I did not treat those as Step 4 blockers, but Step 5 will need to resolve or explicitly document them.
