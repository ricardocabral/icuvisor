# R008 Code Review — Step 3

Verdict: REVISE

## Findings

1. **Output schemas for `get_events` and `get_wellness_data` do not document the new metadata.**  
   The implementation now conditionally returns `_meta.as_of`, `_meta.as_of_date`, `_meta.as_of_weekday`, and `_meta.timezone` for current-day ranges, but the output schema descriptions still describe only the old metadata shapes (`internal/tools/get_events.go:311-312`, `internal/tools/get_wellness_data.go:526-527`). MCP schemas are part of the LLM-facing contract in this repo; without these descriptions, clients and assistants will not know these anchors exist for the two tools. `get_activities` was updated, so please make the events and wellness schema descriptions match the new behavior.

## Verification

- Ran `git diff df49fe3..HEAD --name-only`
- Ran `git diff df49fe3..HEAD`
- Ran `go test ./internal/tools`
- Ran `go test ./...`
