# Plan Review — TP-078 Step 1

**Verdict:** Changes requested before executing Step 1.

## Summary

The current Step 1 plan in `STATUS.md` only repeats the task checklist. For a credential-handling audit, that is too underspecified: it does not define the surfaces to inspect, the search method, or the audit artifact that will become the source of truth for Steps 2–4.

I checked the task prompt/status and sampled the current credential paths. The existing code already has multiple relevant paths and distinctions that the audit must capture precisely:

- `icuvisor setup` reads the key via masked prompt and persists through `credstore.Store` (`internal/app/setup_flow.go`).
- `config.Write` writes only non-secret fields (`internal/config/write.go`).
- runtime loading still accepts process env, keychain, and legacy JSON/.env fallbacks (`internal/config/load.go`).
- docs and installer pages mention setup/keychain, manual keychain storage, and deliberate headless/env fallback across more than just `web/content/install/*`.

Without an explicit audit matrix, Step 2 could accidentally conflate legacy runtime fallback with onboarding write paths, or miss a docs-only installer/client setup path that still tells users to paste keys into JSON.

## Required plan additions

Before implementation, update `STATUS.md` with a concrete Step 1 audit plan and then record the results in the same section. Minimum required shape:

1. **Audit matrix format** with columns like:
   - entrypoint/path
   - user-facing flow (CLI setup, offline setup, installer doc, manual client config, runtime fallback, etc.)
   - whether it accepts an API key
   - how the key is accepted
   - where it is stored or loaded from today
   - desired source of truth
   - follow-up step/action

2. **Explicit search/file scope.** Include at least:
   - `internal/app/setup*.go`, `internal/app/help.go`, `cmd/icuvisor/*`
   - `internal/config/load.go`, `write.go`, `dotenv.go`, `validate.go`, `redaction.go`
   - `internal/credstore/*`
   - `internal/diagnostics/*` and `internal/app/diagnostics.go` for loggable/diagnostic output
   - all user onboarding/client docs, not only installer pages: `web/content/install/*`, `web/content/guides/api-key.md`, `web/content/connect/*`, `web/content/reference/config-file.md`, `web/content/reference/cli.md`, and tutorials/troubleshooting if they mention API keys
   - packaging/installer assets if present

3. **Concrete grep/read commands** to prevent missed surfaces, for example searching for:
   - `api_key`, `INTERVALS_ICU_API_KEY`, `API key`, `keychain`, `credential`, `ReadSecret`, `credstore`, `config.Write`, `.mcp`, `client JSON`

4. **Scope classification rule.** The plan should state that process env and legacy JSON/.env support are runtime/power-user fallback paths, not installer/onboarding write paths. They should be documented as fallback only unless Step 2 intentionally changes compatibility.

5. **No-code-change boundary.** Step 1 should only inspect and update `STATUS.md`; code/docs changes belong to later steps.

## Notes for the audit

Initial sampling suggests the CLI setup path may already be keychain-backed, but the task is specifically about every non-CLI onboarding/installer path. If no separate GUI/basic installer onboarding code exists, record that explicitly in the audit matrix instead of leaving it implicit.

The desired source of truth should remain the existing `internal/credstore` service/account (`icuvisor` / `intervals-icu-api-key`); do not introduce another credential naming scheme.
