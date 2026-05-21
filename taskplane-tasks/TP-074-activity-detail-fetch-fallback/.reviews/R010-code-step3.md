# Code Review — Step 3: Mirror the shape across the four tools

Decision: **revise**.

## Findings

1. **`get_activity_intervals` unavailable responses still fabricate `analyzed:false`.**  
   `unavailableActivityIntervalsResponse` still returns `getActivityIntervalsResponse` (`internal/tools/get_activity_details.go:205-210`), whose `Analyzed bool` field is always serialized as `json:"analyzed"` (`internal/tools/get_activity_details.go:46-49`). As a result, every terminal unavailable interval response (and the Strava-blocked success-payload unavailable path via `stravaUnavailableIntervalsResponse`) includes `"analyzed": false` even though the tool did not retrieve interval analysis state. The Step 3 goal was to mirror the unavailable shape across the four tools and avoid fabricated detail data; streams/splits use dedicated unavailable response structs, but intervals still leaks a success-path field with a misleading false value. Please use a dedicated interval unavailable response shape, or make `Analyzed` optional while preserving success responses by setting it explicitly on successful payloads.

## Verification

- `go test ./internal/tools` — passes.
- `go test ./...` — passes.
