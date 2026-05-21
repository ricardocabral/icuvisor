# R021 code review — Step 4: Tests and verification

Verdict: REVISE

I ran:

```sh
git diff b8ba34ab28eddda8252d568f2727d0cbc7d2c524..HEAD --name-only
git diff b8ba34ab28eddda8252d568f2727d0cbc7d2c524..HEAD
go test -count=1 ./internal/tools
go test -count=1 ./internal/analysis ./internal/tools ./internal/toolcatalog ./internal/toolchecks ./internal/safety ./cmd/gendocs
golangci-lint run ./internal/tools
```

The targeted tests pass. `golangci-lint run ./internal/tools` still fails on implementation files (`unparam` in compute helpers and unused helpers), but those failures are outside this Step 4 test diff and match the prior review notes.

## Blocking findings

### 1. The new sport-filter negative fixture still does not prove sport filtering

- Location: `internal/tools/compute_tools_test.go:375-402`
- Severity: Medium

R020 asked for a non-`Run` completed activity that would be selected if auto-pairing ignored `sport`. The added fixture does include `act-ride-closer`, but it is not actually a more attractive match than the Run candidate the test expects:

```go
activityWithTime("act-targetless", "Run",  "2026-05-01T07:00:00", 3600, ...)
activityWithTime("act-ride-closer", "Ride", "2026-05-01T07:05:00", 3590, nil)
```

For `evt-auto` the target is `3600`, so `act-targetless` is an exact metric match and also sorts before/is no later than the Ride candidate. An implementation that ignored request sport during auto-pairing could still pick `act-targetless`, so the assertion at line 401 would pass while the sport-filter regression remains unpinned.

Please make the nonmatching activity strictly preferable when sport is ignored (for example, exact target/time on the Ride and a worse-but-still-compliant Run candidate), or add a focused separate test where ignoring `sport` deterministically selects the wrong activity.

### 2. Load-balance no-stream behavior is not asserted

- Location: `internal/tools/compute_tools_test.go:167-184`
- Severity: Medium

The Step 4 checklist explicitly calls for zone/load source-priority and no-stream coverage. The zone tests assert `client.intervalCalls == 0`, but the load-balance source-priority test only checks `training_load_total`. If `compute_load_balance` started calling `GetActivityIntervals` despite having precomputed activity zone/load fields, this test would still pass because the fake interval method returns an empty DTO and the test never checks the call count.

Add the same no-stream guard used by the zone tests (or make the fake fail on unexpected interval calls) to the load-balance precomputed-source path.
