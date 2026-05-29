# Code Review: Step 1 — Audit wellness readiness semantics

Verdict: Request changes

## Findings

1. **Incomplete fallback-field discovery.** `STATUS.md:84` says it records the available fallback evidence, but it omits several wellness fields that the audited code exposes and/or annotates: `motivation`, `spO2`, `respiration`, `steps`, `vo2max`, and `baevskySI` (`internal/tools/get_wellness_data.go:174-197`, `291-292`). Since Step 2 will use this discovery to update prompt/docs guidance, please either add these as “where present / supporting only” fallback signals or explicitly state why they are out of scope.

## Verification

- Ran `go test ./internal/tools ./internal/prompts` — pass.
