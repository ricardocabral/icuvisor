# Code Review — TP-039 Step 2: Config + feature flag

## Verdict

Approve.

The Step 2 implementation now satisfies the reviewed config/feature-flag scope: `ICUVISOR_COACH_MODE` supports `off`/`on`/`auto`, `.env` parsing includes the flag, the `coach` stanza is decoded and normalized, ACL patterns are validated against the shared athlete-scoped tool catalog, and enabled coach mode resolves `Config.AthleteID` from `coach.default_athlete_id` instead of the legacy top-level `athlete_id`.

## Findings

No blocking findings.

## Notes

- `internal/toolcatalog` plus the new registry drift tests provide the typo-defense boundary needed for Step 2 without introducing an import cycle from config into tools.
- The config state-machine coverage includes the important regression cases from R006/R007: coach-only configs load in `on`/effective-`auto`, and a legacy top-level `athlete_id` no longer overrides the coach default when coach mode is enabled.
- `Config.String()` remains redacted for credentials, top-level athlete ID, coach roster IDs, and labels.

## Verification

- `go test ./...` passes.
- `make lint` still fails on the pre-existing `internal/app/setup.go:254` staticcheck `ST1005` issue outside this diff.
