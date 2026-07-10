# Plan Review — TP-228 Step 2

## Verdict: REVISE

No Step 2 implementation plan was submitted: the unchecked task checklist repeats the required outcome but does not identify the API/signature changes or how the existing transport helpers will be used safely. This matters because the current `doJSONBody` always serializes a JSON payload, so reusing it for apply would send `{}` or `null`, contrary to the documented bodyless endpoint.

Revise the plan to specify:

1. Add the resolved `RecalcHRZones bool` (or equivalently named) field to `intervals.WriteSportSettingsParams`, and encode it on the update request as the required camel-case query key `recalcHrZones` using `strconv.FormatBool`. The sparse settings body must remain unchanged. Step 3, not a Go zero-value heuristic in the client, owns omitted-input/default-true versus explicit-false resolution.
2. Change `ApplySportSettings` to accept only `(ctx, sportSettingID)` so the old date cannot be passed accidentally. Make it issue one `PUT` to `.../{id}/apply` with an empty URL query and an empty request body (not JSON `{}` or `null`), and retain the same response/error, cancellation, and PUT retry handling.
3. Either add a body-plus-query helper for update and a bodyless request helper for apply, or define a shared helper that supports both without changing existing callers. It must recreate any request body for each retry, close all responses, and preserve the existing 422 and retry behavior. Do not use the GET-only `doJSONQuery`/`do` path for apply.
4. Remove the implicit apply branch from `UpdateSportSettings`. `EffectiveDate` may remain temporarily inert in the internal params type until Step 3 removes the MCP producer, but it must not influence any HTTP request; the final Step 3 plan must remove it entirely.
5. Replace the existing client tests that expect `/apply` and `oldest`. Add exact local-server wire assertions for update true and false query values (including no `recalcHrZones` in the JSON body), a single-update/no-apply regression, and explicit apply assertions for PUT, exact path, empty `RawQuery`, zero-byte body, and no JSON content type/payload. Cover invalid IDs and at least one retry/cancellation/error path if the new helper does not directly reuse already-covered behavior.

The plan should also name `internal/intervals/sport_settings_openapi_contract_test.go` as the location for the exact-wire cases and end with the required `go test ./internal/intervals -run 'SportSettings'` verification.
