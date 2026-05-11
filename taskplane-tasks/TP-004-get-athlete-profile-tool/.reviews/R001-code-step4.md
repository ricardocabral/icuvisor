# Code Review — Step 4: Add tests

## Verdict

**Request changes.** The new tests are a good start and `go test ./...` passes, but two Step 4 checklist items are only partially covered.

## Findings

1. **Full-mode secret/debug omission is not actually tested.**  
   `TestGetAthleteProfileOmitsForbiddenDebugAndSecretFields` only calls the handler with default arguments (`{}`) at `internal/tools/get_athlete_profile_test.go:333-334`. Step 4 marks “default/full response omits debug, raw, URL, header, credential, and timestamp fields” as complete, and the contract says `include_full: true` must still avoid raw upstream payloads, credentials, URLs, headers, debug data, and timestamps. Please run the same forbidden-output assertion against an `{"include_full":true}` response as well, ideally with full-mode fields present so the high-risk path is covered.

2. **The context propagation assertion is too weak to catch regressions.**  
   In `TestGetAthleteProfileHandlerSuccess`, the test checks only that `client.ctx` is non-nil (`internal/tools/get_athlete_profile_test.go:129-138`). That would still pass if the handler called the fake client with `context.Background()` or a fresh context instead of the MCP request context. Since the Step 4 plan explicitly called for context capture and the tool contract requires using the request context, please pass a sentinel context (for example with `context.WithValue`) and assert the fake receives that same context/value.

## Validation run

- `go test ./internal/tools` — passed
- `go test ./...` — passed
