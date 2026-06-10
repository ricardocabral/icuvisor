# R002 Code Review — Step 1

**Verdict:** Request changes

## Findings

1. **False warning scope when `type` and `types` differ** (`internal/athleteprofile/profile.go:240-249`)
   `readinessSportTypes` always prepends `setting.Type` before `setting.Types`. The Step 1 design called for `Types` as authoritative with `Type` only as a fallback when `Types` is empty. If upstream returns both and they diverge (or a fixture/legacy payload has `Type: "Ride", Types: ["Run"]`), the code emits ride power warnings for a run setting and reports a `sport_types` list that does not match the shaped `sport_settings[].types` output. Please prefer `setting.Types` when non-empty and only fall back to `setting.Type` otherwise, with a test covering the fallback behavior.

2. **Heart-rate applicability is too broad for a readiness matrix** (`internal/athleteprofile/profile.go:267-276`)
   `usesHeartRateReadiness` treats every sport except `unknown`, `note`, and `other` as HR-planning applicable. That will create missing HR threshold/zones warnings for unsupported or non-endurance sport settings such as strength/yoga/workout-style types if they appear, which is exactly the kind of false warning the sport/metric matrix was meant to avoid. Please switch this to an explicit allowlist (at least the planned ride/run/swim families, plus any deliberately supported endurance types) or otherwise make unknown types produce no warning until mapped. Add a negative test for a non-applicable sport type.

## Verification

- Ran: `go test ./internal/athleteprofile ./internal/tools -run 'Test.*AthleteProfile|Test.*Sport'`
