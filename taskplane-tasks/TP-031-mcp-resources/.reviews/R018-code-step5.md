# Code Review R018 — Step 5: `icuvisor://athlete-profile`

**Verdict: REVISE**

I reviewed the diff from `cff9600ad5ada5f525b5015fb04c34f5e35eacd4..HEAD`, read the changed files, and ran:

- `go test ./...`
- `make lint`

Both commands pass locally.

## Findings

### 1. Concurrent waiters retry sequentially when the shared refresh fails

**File:** `internal/resources/athlete_profile.go:102-128`

The in-flight refresh coordination now makes cancellation-aware waiters, but waiters only learn that `waitCh` closed. If the shared refresh returns an error, no cache is populated; every waiter re-enters the loop, observes `hasCached == false` and `refreshing == false`, and starts a new upstream `GetAthleteProfile` call. Under an upstream outage, N concurrent `resources/read` calls can therefore produce N serialized upstream calls even though they all arrived behind the same cold/expired refresh.

That conflicts with the Step 5 policy recorded in `STATUS.md`: concurrent reads should share the same cache/refresh, and failed refreshes should return the short safe error without retry loops or request amplification. It also weakens the “do not let athlete-profile refresh hammer the upstream API” acceptance constraint.

Please store the in-flight refresh result/error in the shared refresh state (or equivalent) so all waiters for the same refresh return the same result: on success they can use the newly cached payload, and on failure they should return the same failure without initiating another refresh. A later independent read may try again, but callers already waiting on a refresh should not cascade retries. Add a focused test with several concurrent reads, a blocked client that then returns an error, and an assertion that `GetAthleteProfile` was called once and all callers return the safe failure path.

## Notes

- The R017 cancellation issue is fixed: waiters select on `ctx.Done()` while another refresh is in flight.
- The shared `internal/athleteprofile` shaper keeps the tool and resource output paths aligned, and the resource is wired through production app startup with the configured intervals client.
