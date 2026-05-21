# Plan Review — TP-078 Step 4

**Verdict:** APPROVE

## Findings

No blocking findings.

## Notes

- The Step 4 plan covers the required user-facing documentation work: keychain-backed setup, generated non-secret `credential_ref` metadata, stale JSON/env guidance, packaging copy, and `CHANGELOG.md`.
- When executing, make sure the docs update includes the exact generated config shape (`credential_ref.type/service/account`) in `web/content/reference/config-file.md`, while continuing to frame `api_key`, `.env`, and `INTERVALS_ICU_API_KEY` as compatibility/headless fallbacks rather than normal onboarding.
- The packaging copy fix should remove the stale `icuvisor serve` command from `.goreleaser.yaml` and direct users to the actual default server invocation/client config path.
- Please include a final grep/check for stale normal-user instructions that put API keys in MCP client JSON or config JSON, plus the stale `icuvisor serve` wording, before marking the step complete.
