# Code Review — TP-039 Step 2: Config + feature flag

## Verdict

Revise.

The R006 blockers were mostly addressed: coach-only configs now load in `on`/effective-`auto`, and the registry/catalog drift test is present. However, coach mode still lets the legacy top-level `athlete_id` override `coach.default_athlete_id`, which defeats the Step 2 default-selection contract and can route the current single-athlete client to an athlete outside the coach roster until Step 3 routing lands.

## Findings

### 1. Top-level `athlete_id` overrides `coach.default_athlete_id` in enabled coach mode

- **Severity:** High
- **File:** `internal/config/config.go:553-559`

`validate` initializes `athleteID` from the normalized coach default, but then replaces it whenever the legacy top-level `athlete_id` is present:

```go
athleteID := coachConfig.DefaultAthleteID
if strings.TrimSpace(raw.athleteID) != "" || coach.EffectiveMode(coachMode, coachConfig) != coach.ModeOn {
    athleteID, err = NormalizeAthleteID(raw.athleteID)
    ...
}
```

So this valid coach-mode config loads with `cfg.Coach.DefaultAthleteID == "i333"` but `cfg.AthleteID == "i111"`:

```json
{
  "api_key": "json-key",
  "athlete_id": "111",
  "coach": {
    "athletes": [
      {"id": "222", "allowed_tools": ["get_*"]},
      {"id": "333", "allowed_tools": ["*"]}
    ],
    "default_athlete_id": "333"
  }
}
```

That contradicts the Step 2/R006 requirement to resolve the coach-mode runtime default from `coach.default_athlete_id`. It is also risky during the transition before Step 3 request routing: `intervals.NewClient` still consumes `Config.AthleteID`, so the server can start targeting the legacy single-athlete ID even when coach mode is enabled. If the legacy ID is not in `coach.athletes`, this also undermines the “operate only against configured roster” invariant until the later roster checks are added.

**Recommendation:** When `coach.EffectiveMode(coachMode, coachConfig) == coach.ModeOn`, set `Config.AthleteID` to `coachConfig.DefaultAthleteID` regardless of whether top-level `athlete_id` is present. Optionally fail loudly if a present top-level `athlete_id` differs from the coach default, but do not silently prefer it. Keep the existing top-level `athlete_id` requirement only when effective coach mode is off. Add a regression assertion to `TestLoadCoachConfigSchemaAndValidation` (it already has `athlete_id: "111"` and `default_athlete_id: "333"`) that `cfg.AthleteID == "i333"`, plus a case where top-level `athlete_id` is outside the roster.

## Verification

- `go test ./...` passes.
- `make lint` fails on the existing `internal/app/setup.go:254` staticcheck `ST1005` finding; that file is outside this diff.
