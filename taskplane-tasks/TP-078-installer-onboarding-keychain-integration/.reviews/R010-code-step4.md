# Code Review — TP-078 Step 4

**Verdict:** REVISE

## Findings

1. **Codex validation examples still model API-key env passthrough as the default path** (`docs/clients/codex-local.md:61-79`, `docs/clients/codex-local.md:93-104`, `docs/clients/codex-local.md:115-130`).
   The new prose says `INTERVALS_ICU_API_KEY` should only be passed for the deliberate headless fallback, but every subsequent command example still includes `INTERVALS_ICU_API_KEY` in `mcp_servers.icuvisor.env_vars`, and the profile section still says “With real credentials in the process environment”. That contradicts Step 4’s requirement to remove or clearly mark stale env/JSON API-key instructions for normal onboarding. Please either make the primary examples keychain/setup-based (omit `INTERVALS_ICU_API_KEY`, or pass only non-secret vars / a `--config` path) and move API-key env passthrough to a clearly labeled fallback snippet, or annotate each example so maintainers do not copy it as the normal path.

## Notes

- I ran `git diff 3c469d1..HEAD --name-only`, reviewed the full diff, read the changed docs/config release files, and ran `git diff --check 3c469d1..HEAD` (clean).
- The `credential_ref` shape documented in `web/content/reference/config-file.md` matches the current config/tests (`type: keychain`, service `icuvisor`, account `intervals-icu-api-key`).
