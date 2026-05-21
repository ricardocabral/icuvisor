# TP-078: Installer/onboarding integration for keychain-backed credentials — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 13
**Iteration:** 2
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Audit current onboarding credential paths
**Status:** ✅ Complete

- [x] Define the audit matrix format, search terms, scope, and no-code-change boundary in STATUS.md.
- [x] Add concrete grep/read commands and expanded docs/assets scope for R002 plan feedback.
- [x] Inspect code surfaces that can accept, load, store, redact, or diagnose API keys.
- [x] Inspect onboarding, installer, client-config, tutorial, and troubleshooting docs/assets for API-key instructions.
- [x] Record the audit matrix and desired credential source of truth in STATUS.md before implementation.

**Audit plan:**
- Matrix columns: entrypoint/path; user-facing flow; accepts API key; how accepted; current storage/load behavior; desired source of truth; follow-up step/action.
- Code scope: `internal/app/setup*.go`, `internal/app/help.go`, `cmd/icuvisor/*`, `internal/config/load.go`, `internal/config/write.go`, `internal/config/dotenv.go`, `internal/config/validate.go`, `internal/config/redaction.go`, `internal/credstore/*`, `internal/diagnostics/*`, and `internal/app/diagnostics.go`.
- Docs/assets scope: `web/content/install/*`, `web/content/guides/api-key.md`, `web/content/connect/*`, `web/content/reference/config-file.md`, `web/content/reference/cli.md`, tutorials, troubleshooting, README, `docs/clients/*`, `.goreleaser.yaml`, and any other packaging/release installer text found by `find`/`grep`.
- Search terms: `api_key`, `INTERVALS_ICU_API_KEY`, `API key`, `keychain`, `credential`, `ReadSecret`, `credstore`, `config.Write`, `.mcp`, `client JSON`, `claude_desktop_config`, `setup`.
- Concrete audit commands: `grep -RInE "api_key|INTERVALS_ICU_API_KEY|API key|keychain|credential|ReadSecret|credstore|config\.Write|\.mcp|client JSON|claude_desktop_config|setup" internal cmd docs web README.md CHANGELOG.md .goreleaser.yaml`; `find . -maxdepth 4 \( -iname "*install*" -o -iname "*setup*" -o -iname "*.mcp*" -o -iname "*.plist" -o -iname "*.json" -o -iname "*.yaml" -o -iname "*.yml" \) -not -path "./.git/*" -not -path "./bin/*"`; then targeted reads of matched files in the code and docs/assets scopes above.
- Scope classification: process env plus legacy JSON/.env support are runtime/power-user fallback paths, not installer/onboarding write paths; document them as fallback unless Step 2 intentionally changes compatibility.
- Boundary: Step 1 only inspects and updates STATUS.md; code/docs changes belong to later steps.

**Code audit notes:**
- `icuvisor setup` (`internal/app/setup_cmd.go`, `setup_flow.go`, `setup.go`) accepts an API key only via masked `ReadSecret`; `--api-key` is rejected. It uses `credstore.OSKeychain()` by default, writes config through `config.Write`, stores/verifies the secret with `credstore.Store.Set/Get` using service `icuvisor` and account `intervals-icu-api-key`, and supports `--offline` without changing the storage path.
- Generated setup config (`internal/config/write.go`) writes only `athlete_id`, `timezone`, and optional non-default `api_base_url`; it has no `api_key` field in the write payload.
- Runtime loading (`internal/config/load.go`, `validate.go`, `dotenv.go`) still supports process env, keychain, and legacy JSON/.env keys; env wins, keychain beats plaintext files when env is absent, plaintext config logs a redacted migration warning, and missing-key guidance names env/keychain/legacy fallback. This is a runtime/power-user fallback, not an onboarding write path.
- Redaction/diagnostics surfaces (`internal/config/redaction.go`, `internal/app/diagnostics.go`, `internal/diagnostics/*`) expose key source labels and redacted metadata only; diagnostics suppress config-load logs and print `config_source`, never the secret. `internal/app/wire.go` logs `api_key_source` but currently logs the concrete athlete ID on server start (pre-existing, out of direct API-key scope).

**Docs/assets audit notes:**
- Web install/API-key/connect docs mostly already instruct users to run `icuvisor setup`, keep API keys out of MCP client JSON, and use Keychain/Credential Manager/libsecret service `icuvisor` account `intervals-icu-api-key` (`web/content/install/*`, `web/content/guides/api-key.md`, `web/content/connect/*`, `web/content/reference/cli.md`, `web/content/reference/config-file.md`, `web/content/guides/troubleshooting.md`, `web/content/tutorials/getting-started-chatgpt.md`).
- Manual keychain snippets in `web/content/guides/api-key.md` and `web/content/install/macos.md` are explicit power-user/headless paths; they do not write JSON but they do show placeholder API-key values in shell commands and may need clearer fallback framing in Step 4.
- `docs/clients/codex-local.md` is non-web client validation documentation that still leads with exported `INTERVALS_ICU_API_KEY` and `.env` for real validation; classify this as maintainer/headless fallback and update if affected by Step 4.
- Packaging metadata `.goreleaser.yaml` has Homebrew caveats telling users to run setup/keychain, but it says to point clients at `icuvisor serve`; no `serve` command exists in current CLI help, so this stale packaging copy should be fixed with docs in Step 4 if in scope.
- No GUI/basic installer code path separate from CLI setup was found in the repo; installer pages and release caveats are docs-only onboarding surfaces.

**Audit matrix / source of truth:**

| Entrypoint/path | Flow | Accepts API key? | How accepted | Current storage/load behavior | Desired source of truth | Follow-up |
|---|---|---:|---|---|---|---|
| `icuvisor setup` (`internal/app/setup*.go`) | CLI first-run onboarding | Yes | Masked terminal prompt via `ReadSecret`; no `--api-key` flag | Writes non-secret config via `config.Write`; stores/verifies key with `credstore.Store` | Existing `internal/credstore` OS keychain service `icuvisor`, account `intervals-icu-api-key` | Preserve; add regression coverage for failure guidance/no plaintext |
| `icuvisor setup --offline` | Offline first-run onboarding | Yes | Same masked prompt plus manual athlete ID/timezone | Same keychain store path; skips network verification | Same `internal/credstore` service/account | Preserve headless/offline behavior; test no plaintext config |
| Installer pages / release caveats (`web/content/install/*`, `.goreleaser.yaml`) | Non-CLI installer onboarding copy | Yes, user is told to obtain/paste into setup | Docs direct users to setup/manual keychain; no separate GUI writer found | Docs-only; no code writes plaintext config. `.goreleaser.yaml` caveat has stale `icuvisor serve` wording | Docs should point to setup/keychain and non-secret client config only | Step 4 docs/packaging copy update |
| API-key guide/manual storage (`web/content/guides/api-key.md`, `web/content/install/macos.md`) | Advanced/manual keychain onboarding | Yes | OS keychain tools (`security`, Credential Manager/`cmdkey`, `secret-tool`) with placeholders | Direct keychain storage, not JSON; env documented as fallback | Same `internal/credstore` service/account; env only deliberate fallback | Clarify fallback framing if needed |
| Client config docs (`web/content/connect/*`, `web/content/reference/config-file.md`, `web/content/reference/cli.md`) | MCP client setup | No secret should be entered | Non-secret env/config values and command paths | Mostly says keep API key out of JSON; legacy `api_key` documented as compatibility fallback | Keychain/setup for normal users; env/legacy JSON only runtime fallback | Step 4 remove/mark stale JSON-key instructions if any remain |
| Maintainer Codex doc (`docs/clients/codex-local.md`) | Local validation/headless fallback | Yes | Process env export and `env_vars` pass-through; optional untracked `.env` | Runtime env/.env fallback, not onboarding writer | Prefer setup/keychain for normal onboarding; env only deliberate maintainer/headless fallback | Step 4 update if affected |
| Runtime config load (`internal/config/load.go`, `validate.go`, `dotenv.go`) | Server startup/power-user fallback | Yes | Process env, keychain, legacy JSON/.env | Env wins; keychain loaded when env absent; plaintext config/.env accepted with warning | Keep compatibility but do not use as generated/onboarding write target | Step 2 preserve compatibility/redaction |
| Diagnostics/logging (`internal/app/diagnostics.go`, `internal/config/redaction.go`) | Loggable support output | No | Source labels only | Prints `config_source`; `Config.String` redacts key/athlete; diagnostics suppresses load logs | No secrets in diagnostics/loggable output | Step 3 regression coverage |

**Desired source of truth:** all normal onboarding/install flows should store the intervals.icu API key only through `internal/credstore` using service `icuvisor` and account `intervals-icu-api-key`. Generated config and MCP client JSON should contain only non-secret metadata. Process env and legacy plaintext JSON/.env remain runtime/power-user compatibility fallbacks, not installer/onboarding write paths.

---

### Step 2: Route onboarding credential writes through the keychain
**Status:** ✅ Complete

- [x] Store and verify setup API keys through `internal/credstore` before writing generated config, so keychain failures do not leave a fresh onboarding config behind.
- [x] Add non-secret generated-config credential metadata that references the existing keychain service/account and is accepted by config loading.
- [x] Preserve process-env, keychain, legacy JSON/.env fallback order and redacted logging/diagnostics behavior.

---

### Step 3: Add regression tests for secret handling
**Status:** ✅ Complete

- [x] Add tests proving setup writes credential metadata but no plaintext API key to generated config and diagnostics/loggable output.
- [x] Add/update tests for keychain failure messages, recovery guidance, and no config written when keychain storage fails.
- [x] Add config regression tests for supported/unsupported `credential_ref` metadata without weakening env/keychain/legacy fallback behavior.
- [x] Run targeted tests for `internal/app`, `internal/config`, and `internal/credstore`.

---

### Step 4: Update install and API-key documentation
**Status:** ✅ Complete

- [x] Update end-user docs/reference to describe generated keychain metadata and the keychain-backed setup path.
- [x] Remove or clearly mark stale instructions that imply API keys belong in JSON/env for normal onboarding, including maintainer/client docs where affected.
- [x] Fix packaging/install copy found in the audit that points users at stale command wording.
- [x] Update CHANGELOG.md for user-visible setup/config behavior changes.
- [x] Revise `docs/clients/codex-local.md` examples so keychain/setup is the primary path and API-key env passthrough is clearly fallback-only.

---


### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] Targeted tests passing
- [x] FULL test suite passing: `make test`
- [x] Build passes: `make build`
- [x] Lint passes: `make lint`
- [x] All failures fixed or documented as pre-existing unrelated failures

---

### Step 6: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged
- [x] Final commit includes task ID

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | .reviews/R001-plan-step1.md |
| R002 | Plan | 1 | REVISE | .reviews/R002-plan-step1.md |
| R003 | Plan | 1 | APPROVE | inline reviewer |
| R004 | Code | 1 | APPROVE | inline reviewer |
| R005 | Plan | 2 | APPROVE | .reviews/R005-plan-step2.md |
| R006 | Code | 2 | APPROVE | .reviews/R006-code-step2.md |
| R007 | Plan | 3 | APPROVE | .reviews/R007-plan-step3.md |
| R008 | Code | 3 | APPROVE | .reviews/R008-code-step3.md |
| R009 | Plan | 4 | APPROVE | .reviews/R009-plan-step4.md |
| R010 | Code | 4 | REVISE | .reviews/R010-code-step4.md |
| R011 | Code | 4 | APPROVE | inline reviewer |
| R012 | Plan | 5 | APPROVE | .reviews/R012-plan-step5.md |
| R013 | Code | 5 | APPROVE | inline reviewer |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| No new out-of-scope discoveries during final delivery review. | No follow-up needed. | Step 6 review |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 10:18 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 10:18 | Step 0 started | Preflight |
| 2026-05-20 11:06 | Worker iter 1 | done in 2888s, tools: 169 |
| 2026-05-20 11:22 | Worker iter 2 | done in 949s, tools: 63 |
| 2026-05-20 11:22 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
| 2026-05-20 10:22 | Review R001 | plan Step 1: REVISE |
| 2026-05-20 10:24 | Review R002 | plan Step 1: REVISE |
| 2026-05-20 10:26 | Review R003 | plan Step 1: APPROVE |
| 2026-05-20 10:31 | Review R004 | code Step 1: APPROVE |
| 2026-05-20 10:34 | Review R005 | plan Step 2: APPROVE |
| 2026-05-20 10:39 | Review R006 | code Step 2: APPROVE |
| 2026-05-20 10:41 | Review R007 | plan Step 3: APPROVE |
| 2026-05-20 10:49 | Review R008 | code Step 3: APPROVE |
| 2026-05-20 10:51 | Review R009 | plan Step 4: APPROVE |
| 2026-05-20 10:56 | Review R010 | code Step 4: REVISE |
| 2026-05-20 11:16 | Review R011 | code Step 4: APPROVE |
| 2026-05-20 11:17 | Review R012 | plan Step 5: APPROVE |
| 2026-05-20 11:20 | Review R013 | code Step 5: APPROVE |
| 2026-05-20 | Step 6 doc review | README points users to website and only lists project layout; tools reference/catalog not affected; PRD already requires OS keychain and headless env for power users. |
