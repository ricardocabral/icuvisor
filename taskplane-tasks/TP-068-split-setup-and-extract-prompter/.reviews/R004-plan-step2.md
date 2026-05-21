# Plan Review — TP-068 Step 2

## Verdict

Approved with guardrails. The planned Step 2 is appropriately narrow: slice `RunSetup` into focused helpers, keep behavior unchanged, and run the existing `internal/app` regression suite added in Step 1.

## Plan notes / guardrails

- Keep Step 2 mechanical. Do not extract `terminalPrompter` or move setup arg parsing yet; those are explicitly Step 3 and Step 4.
- Prefer a small private flow/dependencies struct if it keeps helper signatures readable. `RunSetup` currently normalizes many dependencies (`Stdout`, keychain store, prompter, config writer, profile fetcher, timezone detector); duplicating that parameter list across helpers would make the split noisier rather than clearer.
- Preserve the exact side-effect order:
  1. context check and dependency defaults,
  2. welcome text,
  3. existing-key overwrite prompt/cancel before config checks,
  4. config overwrite prompt/cancel before reading the secret,
  5. secret read/trim/empty validation,
  6. profile/timezone collection,
  7. config write,
  8. keychain set + round-trip verification,
  9. success/final online verification/next-step output.
- Keep stdout and prompt strings byte-identical. The happy-path test now checks exact stdout and prompt lists; do not reformat output while introducing helpers.
- Be careful with `--force`: it only skips the config overwrite prompt; it must not skip the existing API-key overwrite confirmation.
- Keep the “Saved...” line after successful keychain round-trip verification, not merely after config write or keychain set.
- After the slice, run at least `go test ./internal/app`; `go test ./...` is a good cheap extra because `RunSetup` is also exercised through CLI dispatch tests.

## Suggested helper boundaries

A reasonable 4–6 helper split would be along these existing seams:

- normalize/default `SetupOptions` into private flow state,
- print setup intro,
- confirm existing API-key overwrite,
- confirm config overwrite and return `AllowOverwrite`,
- read/validate API key,
- persist config/key and perform final verification/output.

Existing `setupProfile` and `setupTimezone` are already good step helpers; avoid churning them unless needed for call-site clarity.

## Blocking findings

None.
