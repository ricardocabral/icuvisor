# Code Review — TP-228 Step 2

## Verdict: APPROVE

The client now sends the required `recalcHrZones` query on sparse sport-settings updates, preserves explicit true/false values, and never includes it in the JSON body. `ApplySportSettings` has a date-free signature and issues a bodyless, queryless PUT. `UpdateSportSettings` no longer invokes apply.

The shared helpers create a fresh request/body for every retry and retain the previous response/error handling. Exact local wire tests cover both recalculation values, a zero-byte apply request, and invalid IDs.

Verified:

- `go test ./internal/intervals -run 'SportSettings' -count=1`
- `go test ./... -count=1`
- `go vet ./internal/intervals`
- `git diff --check 735cca45ccd77f1d2e4157500779cdb4a6ca53df..HEAD`
