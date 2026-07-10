# Code Review — TP-228 Step 3

## Verdict: REVISE

### P1 — Reject `recalc_hr_zones: null` instead of treating it as omitted

`internal/tools/update_sport_settings.go:189` resolves a nil `*bool` to `true`. Go's JSON decoding makes both an omitted `recalc_hr_zones` and an explicitly supplied JSON `null` nil. The registered MCP handler uses the non-generic go-sdk `Server.AddTool`, which explicitly leaves input-schema validation to the caller, so this is reachable rather than being rejected by the schema. Consequently, `{"sport":"Ride","ftp":290,"recalc_hr_zones":null}` violates the declared boolean schema yet reaches the write client and requests HR-zone recalculation.

Distinguish omission from an explicit `null` in the raw arguments and return the existing terse invalid-arguments error for null. Add a no-client-call regression case alongside the legacy/unknown argument tests.

## Verification

- `go test ./internal/tools ./internal/toolchecks`
- `go test -race ./internal/tools ./internal/toolchecks`
- `go vet ./internal/tools ./internal/toolchecks`
- Schema freshness and stability against the supplied baseline: PASS
- `make docs-tools` followed by generated catalog diff check: clean
