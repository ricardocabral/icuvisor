# Plan Review — TP-002 Step 2

## Verdict

APPROVE

## Findings

No blocking findings.

## Notes

- The Step 2 scope is appropriately limited to the HTTP client core: constructor/config plumbing, Basic Auth, `User-Agent`, context-aware requests, response-body closure, and typed JSON decoding.
- When implementing, preserve the Step 1 clean-room/auth decision exactly: intervals.icu Basic Auth uses username `API_KEY` and the configured API key as the password. Do not place the key in URLs, error strings, fixtures, or logs.
- Prefer a small `internal/intervals` API such as `New(config.Config, version string, httpClient *http.Client)` plus an internal request/decode helper. Keep retries and structured error taxonomy in Step 3 unless the helper needs minimal temporary status handling.
- Ensure the constructor does not create a new client per request: use the injected shared `*http.Client`, defaulting to one with `config.HTTPTimeout` only when none is supplied.
- Build URLs from the configured base URL (`https://intervals.icu/api/v1` by default, already slash-trimmed by config) with path-safe joining so later `/athlete/{id}` calls do not accidentally drop `/api/v1`.
- Use `http.NewRequestWithContext`, set `User-Agent: icuvisor/<version>` on every request, and close bodies on every path, including JSON decode errors and non-2xx responses.
- Typed structs should match the Step 1 planned `/athlete/{id}` / `WithSportSettings` shape; avoid adding raw `map[string]any` payload capture unless a later reviewed step justifies it.
