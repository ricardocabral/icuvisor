# Code Review R019 — Step 5: `icuvisor://athlete-profile`

**Verdict: REVISE**

I reviewed the diff from `cff9600ad5ada5f525b5015fb04c34f5e35eacd4..HEAD`, read the changed files, and ran:

- `go test ./...` — passes
- `make lint` — fails

## Findings

### 1. Lint fails on the athlete-profile resource reader

**File:** `internal/resources/athlete_profile.go:134`

`make lint` currently fails with staticcheck:

```text
internal/resources/athlete_profile.go:134:3: SA4004: the surrounding loop is unconditionally terminated (staticcheck)
		return result, err
		^
```

`athleteProfileReader.Read` is written as a `for` loop, but every control path either returns immediately or waits for the in-flight refresh result and then returns. Because the loop never actually iterates, CI lint will fail.

Please remove the unnecessary loop or restructure the function so the retry/wait behavior is explicit and accepted by staticcheck. The existing concurrency behavior from R017/R018 should be preserved: waiters should select on `ctx.Done()` while a refresh is in flight, and waiters behind a failed refresh should return that same failure instead of starting serialized retries.

## Notes

- The R017 and R018 behavioral fixes appear to be covered by focused tests.
- `go test ./...` passes locally.
- The shared `internal/athleteprofile` shaper keeps the tool and resource output paths aligned, and production app startup now wires the configured intervals client into the dynamic resource registry.
