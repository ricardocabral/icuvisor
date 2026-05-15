# Code Review — TP-039 Step 2: Config + feature flag

## Verdict

Revise.

The implementation covers most of the Step 2 surface (`ICUVISOR_COACH_MODE`, `.env` parsing, coach stanza decoding/normalization, ACL typo checks, and redacted `Config.String()`). However, coach-mode configs still cannot be loaded unless the legacy single-athlete `athlete_id` is also configured, which conflicts with the new coach-mode config model and will block the advertised `auto`/`on` path for coach-only configs.

## Findings

### 1. Coach mode still requires top-level single-athlete `athlete_id`

- **Severity:** High
- **File:** `internal/config/config.go:540-604`

`validate` normalizes `raw.athleteID` before parsing/validating the coach stanza or resolving effective coach mode:

```go
athleteID, err := NormalizeAthleteID(raw.athleteID)
if err != nil {
    return Config{}, err
}
```

That means a config like this cannot load, even though it has a valid coach roster and default athlete:

```json
{
  "api_key": "...",
  "coach": {
    "athletes": [
      {"id": "i222", "allowed_tools": ["*"]}
    ],
    "default_athlete_id": "i222"
  }
}
```

with `ICUVISOR_COACH_MODE=on` or `auto`; it fails with the existing "missing athlete ID" error before coach mode is considered.

This conflicts with the TP-039 model: in coach mode the target athlete comes from `coach.default_athlete_id` / `athlete_id` request selection, while the configured single athlete is only relevant to non-coach mode. The Step 2 schema and acceptance criteria describe `coach.default_athlete_id` as the coach-mode default, and mention the configured single athlete only for non-coach mode rejection.

**Recommendation:** Parse/validate coach mode and coach config before requiring the top-level athlete ID. When effective coach mode is on, either:

- allow top-level `athlete_id` to be omitted and set `Config.AthleteID` to `coach.default_athlete_id` as a temporary compatibility value until Step 3 request routing lands, or
- store it empty but make sure all code that constructs an intervals client/server can tolerate that in coach mode.

When effective coach mode is off, keep the existing `NormalizeAthleteID(raw.athleteID)` requirement unchanged. Add tests for `on` and `auto` with a populated coach roster and no top-level `athlete_id`.

### 2. Shared catalog constants are not yet the source used by tool definitions

- **Severity:** Medium
- **Files:** `internal/toolcatalog/catalog.go`, `internal/tools/registry.go:96-101`

The approved Step 2 plan called for `internal/toolcatalog` to be the shared catalog contract via exported canonical tool-name constants consumed by registry/config consumers. Config validation now uses `toolcatalog`, but the actual tool constructors still use their existing private string constants (for example, `getAthleteProfileName`, `deleteEventName`, etc.). The registry only checks at runtime that each private name appears in `toolcatalog`.

That catches drift late, but it does not make `toolcatalog` the source of truth: a future tool rename/addition can compile and then fail during registration unless tests catch it. It also means ACL validation and registered tool names are still maintained in two places.

**Recommendation:** Either update tool definitions to use the exported `toolcatalog` constants directly, or add an explicit drift-prevention test that compares the registered catalog against `toolcatalog.AthleteScopedToolNames()` and fails if either side has an extra/missing athlete-scoped tool. The current full-catalog test is close, but the production registry runtime check is still compensating for duplicated constants rather than removing the duplication.

## Verification

- `go test ./...` passes.
- `make lint` currently fails on a pre-existing `internal/app/setup.go:254` staticcheck `ST1005` finding outside this diff.
