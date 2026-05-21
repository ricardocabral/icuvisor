# Code Review: TP-020 Step 1 (`add_or_update_event`)

**Verdict: Approve**

I reviewed the full diff from `a4ff0a370fa81970ec14bcb8af141bb860aac96c^..HEAD` and read the changed implementation/tests for the intervals client, tool registration, schema catalog, README/CHANGELOG, and task status.

## Verification run

- `go test ./internal/intervals ./internal/tools ./internal/toolchecks ./internal/mcp` — pass
- `go test ./...` — pass
- `go test -race ./...` — pass
- `make lint` — pass
- `make build` — pass
- `go run ./scripts/snapshot_tool_schemas.go` followed by `git diff --exit-code -- internal/tools/schema_snapshot` — no schema snapshot drift

## Findings

No blocking findings.

The prior R002 issues appear addressed:

- Planned write fields now use upstream planned target keys (`load_target`, `distance_target`, `time_target`, `elapsed_time_target`) and read shaping exposes them separately from completed metrics.
- `add_or_update_event` is represented in schema snapshot generation via `schemaCatalogClient`, has a committed snapshot, and has regression coverage for catalog inclusion.
- README catalog and `[Unreleased]` changelog entries were added.

## Notes

- The new write helper deliberately avoids retrying `POST` creates while allowing retries for same-resource `PUT` updates, which matches the duplicate-create concern from R001.
- Free-text `description` is passed through without trimming, while structured `workout_doc` is serialized into the upstream `description` field and not sent as a structured upstream key.
