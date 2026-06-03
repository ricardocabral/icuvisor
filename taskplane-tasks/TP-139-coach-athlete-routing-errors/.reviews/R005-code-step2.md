# Code Review — Step 2

Result: **revise**

## Findings

1. **`select_athlete` still accepts credential-like parameters instead of rejecting them.**
   - Location: `internal/tools/select_athlete.go:46-48`
   - The handler uses plain `json.Unmarshal` into `selectAthleteRequest`, so calls such as `{"athlete_id":"i123","api_key":"secret"}` are accepted and silently ignored. Step 2 explicitly requires that tool/chat parameters do not accept API keys, and the project hard rule says API keys must never be accepted as tool parameters. The new schema-only regression in `internal/tools/registry_test.go` verifies these fields are not advertised, but it does not catch runtime acceptance by handlers.
   - Suggested fix: decode `select_athlete` arguments with `DecodeStrict[selectAthleteRequest]` (or equivalent) and add a regression that a credential-like extra field returns a public invalid-argument/user error without changing the selected athlete.

## Tests run

- `go test ./internal/coach ./internal/config ./internal/tools ./internal/mcp`
