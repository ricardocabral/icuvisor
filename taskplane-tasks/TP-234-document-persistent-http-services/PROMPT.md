# Task: TP-234 - Add persistent loopback HTTP service recipes

**Created:** 2026-07-10
**Size:** M

## Review Level: 1 (Plan Only)

**Assessment:** This adds cross-platform operational documentation without changing runtime code. The main risk is publishing insecure service definitions that expose credentials or bind beyond loopback.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 1, Security: 1, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-234-document-persistent-http-services/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Document durable, OS-native ways to keep icuvisor's local Streamable HTTP endpoint running for clients that connect to an existing URL rather than launching stdio. Add tested recipes for a macOS LaunchAgent, a Linux systemd user service, and Windows Task Scheduler or an equally standard built-in mechanism. Every recipe must bind `127.0.0.1`, rely on the existing keychain/credential chain rather than plaintext secrets, include start/status/stop/remove instructions, and explain when hosted HTTPS is the correct alternative. Do not introduce PM2, Docker, or a new runtime dependency.

## Dependencies

- **Task:** TP-232 (hosted HTTP troubleshooting guidance must be corrected first)

## Context to Read First

**Tier 2 (area context):**

- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**

- `web/content/guides/http-transport.md` — current foreground HTTP setup and loopback warning
- `web/content/explain/privacy.md` — transport trust boundary
- `web/content/reference/cli.md` — exact flags and environment variables
- `web/content/reference/config-file.md` — persistent non-secret configuration

## Environment

- **Workspace:** `web/` documentation site
- **Services required:** Hugo for website build; no background service should actually be installed

## File Scope

- `web/content/guides/persistent-http-service.md`
- `web/content/guides/http-transport.md`
- `web/content/guides/_index.md`
- `web/content/guides/troubleshooting.md`
- `scripts/tests/test_http_service_docs.py`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] TP-232 is complete
- [ ] Confirm current macOS, Windows, and Linux binary/config paths from existing install docs

### Step 1: Design secure service recipes

**Plan-review checkpoint**

- [ ] Choose one built-in lifecycle mechanism per OS: launchd user agent, systemd user service, and Windows Task Scheduler or documented built-in equivalent
- [ ] Keep credentials out of service definitions and use the existing config/keychain lookup
- [ ] Pin every command and sample config to `127.0.0.1:8765` and `/mcp`
- [ ] Include install/start/status/log/stop/remove commands and failure recovery
- [ ] Distinguish same-machine HTTP from provider-hosted remote connector requirements

**Artifacts:**

- `web/content/guides/persistent-http-service.md` (new)

### Step 2: Write and integrate the guide

- [ ] Add complete copy-pasteable recipes with placeholders that cannot be mistaken for secrets
- [ ] Link the new guide from HTTP transport, guide index, and relevant troubleshooting content
- [ ] Warn against punctuation-heavy connector keys where clients reject them; use `icuvisor` consistently
- [ ] Preserve the unauthenticated-LAN warning and link remote-only clients to hosted mode
- [ ] Run targeted checks: `make web-build`

**Artifacts:**

- `web/content/guides/persistent-http-service.md` (new)
- `web/content/guides/http-transport.md` (modified)
- `web/content/guides/_index.md` (modified)
- `web/content/guides/troubleshooting.md` (modified if needed)

### Step 3: Add documentation contract coverage

- [ ] Add a static documentation test requiring all three OS sections, loopback binding, lifecycle instructions, credential-store guidance, and hosted-mode fallback
- [ ] Assert the guide never includes an API-key assignment or LAN wildcard bind in executable examples
- [ ] Run targeted tests: `python3 scripts/tests/test_http_service_docs.py`

**Artifacts:**

- `scripts/tests/test_http_service_docs.py` (new)

### Step 4: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run documentation contract test: `python3 scripts/tests/test_http_service_docs.py`
- [ ] Build website: `make web-build`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`
- [ ] Verify clean Markdown/diff: `git diff --check`

### Step 5: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**

- `web/content/guides/persistent-http-service.md` — new cross-platform persistent-service guide
- `web/content/guides/http-transport.md` — link to persistence instructions
- `web/content/guides/_index.md` — expose the guide in navigation
- `CHANGELOG.md` — record new operational documentation under Unreleased

**Check If Affected:**

- `web/content/guides/troubleshooting.md` — link for process-not-running failures
- `web/content/explain/privacy.md` — ensure loopback and hosted trust boundaries remain consistent
- Platform install docs — verify binary paths used in examples

## Completion Criteria

- [ ] macOS, Linux, and Windows persistent user-service recipes exist
- [ ] Every executable sample binds loopback only
- [ ] No plaintext API key appears in service definitions
- [ ] Lifecycle and log troubleshooting commands are provided
- [ ] Hosted mode is recommended for provider-hosted remote connector UIs
- [ ] New documentation test exists and passes
- [ ] Full tests, web build, lint, and binary build pass

## Git Commit Convention

Commits happen at step boundaries. All commits MUST include TP-234:

- **Step completion:** `docs(TP-234): complete Step N — description`
- **Bug fixes:** `docs(TP-234): description`
- **Tests:** `test(TP-234): description`
- **Hydration:** `hydrate: TP-234 expand Step N checkboxes`

## Do NOT

- Install or start a background service during task execution
- Put `INTERVALS_ICU_API_KEY` values in service files or examples
- Bind `0.0.0.0`, `::`, or a LAN IP in executable examples
- Recommend PM2, Docker, ngrok, cloudflared, or another new dependency/tunnel
- Change HTTP runtime code
- Duplicate TP-232's correction instead of building on it
- Commit without TP-234 in the message

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if prerequisites or instructions are contradictory. -->
