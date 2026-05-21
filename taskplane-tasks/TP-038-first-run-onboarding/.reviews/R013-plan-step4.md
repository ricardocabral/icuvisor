# Review R013 — Plan Review for Step 4: Write

Verdict: **APPROVE**

The updated Step 4 plan in `STATUS.md` resolves the blocking issues from R012 and is concrete enough to implement.

## What is now covered

- Existing key/config prompts still happen before secret entry, and the first online profile verification remains the no-write gate for 401/403 and normal network failures.
- Credential writes use `credstore.IntervalsAPIKeyAccount`, perform `Set` followed by `Get`, require an exact secret match, and avoid including secret values in errors.
- Config overwrite semantics are now consistent: interactive confirmation and `--force` both authorize overwrite via `WriteOptions{AllowOverwrite bool}`, while a race-created file remains protected when overwrite was not authorized.
- The config writer shape is sufficiently specified: context-aware `internal/config.Write`, dedicated non-secret write struct, no `api_key` marshalling path, athlete/timezone validation, private permissions, parent directory creation, and atomic/no-clobber behavior.
- The final online test is explicit: after the keychain round-trip, call the injected/default profile fetcher again using the stored/retrieved secret, then print the `Test connection OK` output. Offline mode correctly avoids claiming verification success.
- The planned tests cover the important write/failure cases, including no-secret JSON, keychain failures, round-trip mismatch, overwrite/no-clobber paths, offline writes, and preserving no-write behavior for pre-write auth/network failures.

## Implementation notes

- Be careful that the config writer's no-clobber path protects the final target, not just the temporary file. Using an exclusive create/link-style flow is preferable to a temp-file `rename` that can overwrite an existing target unexpectedly.
- If the final post-write profile fetch fails after config/keychain writes have succeeded, return a non-secret error and avoid the success message as planned. No rollback is required by the prompt, but tests should make the partial-write behavior explicit if you add that case.

Proceed with Step 4 implementation.
