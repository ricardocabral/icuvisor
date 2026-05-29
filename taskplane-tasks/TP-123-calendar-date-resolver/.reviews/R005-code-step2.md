# Code Review — Step 2

Verdict: REVISE

## Findings

1. **Blocking: full test suite is failing due stale catalog guard surfaces.** Registering `resolve_calendar_dates` in `registryBaseTools` (`internal/tools/catalog.go:75`) changes the public tool catalog, but `cmd/gendocs/testdata/tools.golden.json` and the static safety catalog in `internal/safety/adversarial_test.go:23` were not updated. `go test ./...` fails in `cmd/gendocs` (`TestRunWritesToolsCatalogGolden`) and `internal/safety` (`TestAdversarialStaticCatalogMatrix`, counts are +1 for safe/full/none). Update those golden/static catalog surfaces so CI can pass.

2. **Minor: invalid timezone is reported as invalid arguments.** `shapeResolveCalendarDates` returns timezone load failures from `time.LoadLocation` (`internal/tools/resolve_calendar_dates.go:121`), but the handler wraps every shape error with `invalidResolveCalendarDatesMessage` (`internal/tools/resolve_calendar_dates.go:78-80`). For a bad athlete/config timezone, the user gets told to fix `base_date`/`offsets` instead of “check athlete timezone”, even though `fetchResolveCalendarDatesMessage` exists for that case.

## Verification

- `go test ./internal/tools ./internal/toolcatalog` — pass
- `go test ./internal/toolchecks` — pass
- `go test ./...` — fail as described above
