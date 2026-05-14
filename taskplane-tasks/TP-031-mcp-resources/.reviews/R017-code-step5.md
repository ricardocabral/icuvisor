# Code Review R017 — Step 5: `icuvisor://athlete-profile`

**Verdict: REVISE**

I reviewed the diff from `cff9600ad5ada5f525b5015fb04c34f5e35eacd4..HEAD`, read the changed files, and ran:

- `go test ./...`
- `make lint`

Both commands pass locally.

## Findings

### 1. Canceled reads can block indefinitely behind an in-flight refresh

**File:** `internal/resources/athlete_profile.go:88-104`

`athleteProfileReader.Read` checks `ctx.Err()` before calling `r.mu.Lock()`, but `sync.Mutex.Lock` itself is not context-aware. If one `resources/read` is doing a slow or stuck `GetAthleteProfile` refresh while holding the mutex, a second read whose context is canceled while waiting for the mutex will not return until the first refresh finishes. That violates the Step 5 policy recorded in `STATUS.md`: “Context cancellation is honored before acquiring/refreshing the cache.”

This is also a realistic MCP behavior problem: a client-side timeout/cancel for `icuvisor://athlete-profile` can still tie up the request goroutine behind another refresh instead of returning promptly.

Please change the cache/single-refresh coordination so waiting readers select on `ctx.Done()` while waiting for the refresh/cache state, then add a focused test with one blocked refresh and a second already-canceled or soon-canceled read proving the second returns `context.Canceled` without waiting for the first refresh to unblock. A channel-based in-flight refresh state (or equivalent context-aware coordination) would satisfy this without allowing duplicate upstream calls.

## Notes

- The shared `internal/athleteprofile` shaper and shape-parity test direction look good.
- The resource is wired through production app startup with the real intervals client, and protocol coverage now asserts all four resources when a profile client is configured.
