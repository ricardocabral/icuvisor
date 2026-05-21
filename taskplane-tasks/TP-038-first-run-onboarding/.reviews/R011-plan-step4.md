# Review R011 — Plan Review for Step 4: Write

Verdict: **REVISE**

I could not find a concrete Step 4 implementation plan to review. `STATUS.md` marks Step 4 as in progress, but it only contains the three checklist items from the prompt. Before coding, please add the Step 4 write plan to `STATUS.md` and resubmit.

## Required plan details

1. **Write sequencing and failure behavior.**
   - Specify the exact order after successful verification/offline collection: config existence/force handling, config write, credential `Set`, credential `Get` round-trip, and success/final-test output.
   - State what happens if one write succeeds and a later step fails. Avoid declaring success until both the config write and keychain round-trip have completed.
   - Preserve the existing guarantee that 401/403 and online network failures without `--offline` perform no writes.

2. **Credential-store contract.**
   - Use `credstore.IntervalsAPIKeyAccount` rather than duplicating the literal account string.
   - Call `Store.Set(ctx, credstore.IntervalsAPIKeyAccount, secret)` and immediately `Store.Get(ctx, credstore.IntervalsAPIKeyAccount)`.
   - Define the round-trip check precisely: `Get` must return the same secret; a mismatch is an error, and neither the mismatch nor any wrapped error should include the API key.
   - Confirm errors are wrapped with useful non-secret context only.

3. **Config writer shape and safety.**
   - Add the planned `internal/config` write helper and name its API. Because it performs I/O, prefer a context-aware helper consistent with repo conventions.
   - Do not marshal `config.Config` directly while it still has `json:"api_key"`; use a dedicated file/write struct that cannot emit `api_key`.
   - Write only non-secret setup fields: `athlete_id`, `timezone`, and `api_base_url` only when non-default/non-empty.
   - Validate/normalize `athlete_id` and `timezone` before writing, and ensure the file round-trips through `config.Load` when a key is supplied by the credential store.

4. **Clobber and filesystem behavior.**
   - Enforce the “refuse to clobber without `--force`” rule at the actual write point, not only with the earlier `os.Stat` prompt, to avoid a race where a file appears after the prompt.
   - Plan parent directory creation for the default `~/.config/icuvisor/config.json` path.
   - Specify file permissions and atomicity. A reasonable target is creating parent dirs with private permissions, writing a temp file without secrets, then renaming, while using exclusive create/no-clobber semantics unless `--force` is set.

5. **User-facing success/final-test output.**
   - Replace the current placeholder “writing continues” message with the final saved message from the UX script: key stored in OS keychain, non-secret config path, and next-step documentation pointer.
   - Include the final “Test connection OK: <name>, FTP <n> W” behavior for online setup, using the typed intervals client path rather than the MCP tool. If offline mode cannot verify, say that clearly instead of claiming a test passed.

6. **Tests to add or update with Step 4.**
   - Add table-driven tests for: happy path writes key + non-secret config; `api_key` is absent from the JSON; keychain `Set` failure; keychain `Get` failure; keychain round-trip mismatch; config already exists with and without `--force`; config write failure; offline write path; and existing bad-key/network-failure cases still produce no writes.
   - Include a config round-trip test that loads the written file with a fake credential store and verifies canonical `athlete_id`, timezone, and default/non-default `api_base_url` handling.

Once these points are captured in `STATUS.md`, the Step 4 plan should be reviewable.
