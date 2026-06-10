# R005 Code Review — Step 2

**Verdict:** Request changes

## Findings

1. **Warning actions still do not point to the actual write fields.** In `internal/athleteprofile/profile.go:219` the HR warning tells the model to “set LTHR/FTHR”, but `update_sport_settings` only accepts `threshold_hr`; zone actions likewise say “set power/heart-rate/pace zones” without the actual `zones` field/kind. Step 2 explicitly required the warnings/update-tool guidance to line up with `update_sport_settings` fields (`ftp`, `threshold_hr`, `threshold_pace`, `zones`) so an LLM can send valid arguments. Please update the action strings and add assertions that the serialized warnings mention the accepted fields and do not imply raw upstream field names as the only instruction.

2. **Missing warning coverage still bypasses the tool serialization boundary.** The missing Ride/Run/Swim and complete-settings tests in `internal/tools/get_athlete_profile_test.go:195` and `:284` call `newGetAthleteProfileResponse` directly, not `tool.Handler` with `{}` and decoded `TextResult`. That does not prove `_meta.warnings` survives terse response shaping/serialization, which was the main Step 2 contract. Add handler-level tests for missing warnings and for complete settings; include alias completeness (`FTHR`, `PaceThreshold`) so alias regressions are caught.

## Verification

- `go test ./internal/tools ./internal/resources`
