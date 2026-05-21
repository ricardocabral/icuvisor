# Review R014 — Code Review for Step 4: Write

Verdict: **REVISE**

## Findings

1. **`RunSetup` prints the saved/success message before the final online verification can fail.**  
   Location: `internal/app/setup.go:211-218`

   Step 4's approved plan explicitly says the final typed-client test should happen after the keychain round-trip and that setup should only then print the saved/final-test output; if the final post-write profile fetch fails, it should return a non-secret error and avoid the success message. The current code prints:

   ```text
   Saved. Your key is in the OS keychain; athlete id ...
   ```

   before calling `profileFetcher(ctx, storedSecret)`. If that final fetch returns a network/auth error, the command returns `final test connection failed`, but stdout has already claimed a successful save. This is confusing for the first-run path and violates the Step 4 behavior the plan committed to.

   Suggested fix: move the `Saved...` line until after the final online verification succeeds. For offline mode, it is fine to print the saved line immediately after the keychain round-trip because no final test is attempted. Add a regression test where the first profile fetch succeeds, config/keychain writes succeed, and the second/final fetch fails; assert stdout does not contain `Saved.` or `Test connection OK`.

## What I checked

- Reviewed the full diff from `8aeec43a99576c5f25c4abdadae3605cd35d450c..HEAD`.
- Read the changed setup/config implementation and tests.
- Ran `go test ./...` — passing.
