# Code Review — Step 3: Mirror the shape across the four tools

Decision: **approve**.

## Findings

No blocking findings.

The Step 3 revision addresses the previous interval-shape blocker: unavailable `get_activity_intervals` payloads now use a dedicated response struct and no longer emit fabricated `analyzed:false`. The shared classification/fallback helper is small, preserves context-cancellation behavior, and the streams/splits constructors and registry wiring now carry an explicit `ActivityDetailsClient` for Strava-block detection.

## Verification

- `go test ./internal/tools` — passes.
- `go test ./...` — passes.
- `gofmt -l internal/tools/activity_unavailable.go internal/tools/get_activity_details.go internal/tools/get_activity_streams.go internal/tools/get_extended_metrics.go internal/tools/registry.go` — no output.
