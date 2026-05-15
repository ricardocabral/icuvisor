# Review R012 — Plan Review for Step 4: Write

Verdict: **REVISE**

The updated `STATUS.md` adds a real Step 4 plan and it covers many of the previous gaps: no-write guarantees before verification, use of `credstore.IntervalsAPIKeyAccount`, `Set` + `Get` round-trip matching without secret disclosure, a dedicated non-secret config writer, private filesystem permissions, atomic/no-clobber intent, and the main failure-path tests.

Two plan-level issues still need clarification before coding because they affect observable behavior and test expectations.

## Blocking issues

1. **Config overwrite semantics are contradictory.**
   - The plan says the existing key/config prompts still happen before reading the new key, but also says the config writer uses no-clobber semantics unless `--force` is true.
   - With the current Step 2 prompt (`A config file already exists ... Overwrite? [y/N]`), a user can answer `yes` without `--force`; as planned, the later write would still refuse to overwrite. That means setup would read the secret and verify the profile, then fail at the write point despite the user accepting the overwrite prompt.
   - Pick and document one behavior:
     - If `--force` is required to overwrite config files, change the pre-secret flow/message to cancel early and tell the user to rerun with `--force`, then keep the write-point no-clobber guard for races; or
     - If interactive confirmation is allowed to authorize overwrite, pass an explicit `AllowOverwrite`/`Force` value to the config writer after confirmation and define how the race guard behaves.
   - This must be reflected in tests for existing config with and without `--force`.

2. **The final “Test connection” is only described as output, not as a verification call.**
   - The mission requires setup to run a final “Test connection” using the same typed client path as `get_athlete_profile` and print athlete name + FTP.
   - The plan currently says to print `Test connection OK: ...` after config write and keychain `Set`/`Get`, but does not say whether it performs a second profile fetch after the writes or just reuses the earlier pre-write profile.
   - Make the final online behavior explicit: after the keychain round-trip succeeds, call the injected/default profile fetcher again using the stored/retrieved secret (or load the written config with the credential store and call the typed intervals client), then print the result. Offline mode should continue to say verification was skipped.

## Non-blocking clarification

- The `internal/config.Write(ctx, path, Config, options...) (or equivalent)` API is still a little loose. Prefer naming the option shape in the plan, e.g. `WriteOptions{Force bool}` or `WriteOptions{AllowOverwrite bool}`, so implementation and tests can align with the overwrite decision above. The important part remains: use a dedicated file/write struct that cannot marshal `api_key`.

Once those two behavioral points are resolved in `STATUS.md`, the Step 4 plan should be ready to implement.
