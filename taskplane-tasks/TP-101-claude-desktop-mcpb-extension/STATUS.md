# TP-101: ICUvisor Desktop Extension (.mcpb) for Claude Desktop — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 3
**Review Counter:** 16
**Iteration:** 3
**Size:** L

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Research and decide bundle shape
**Status:** ✅ Complete

- [x] Read the MCPB manifest spec and examples; record required fields and binary-server support in STATUS.md.
- [x] Decide artifact naming (`.mcpb`) and whether to also mention legacy `.dxt` compatibility.
- [x] Decide how user_config maps to icuvisor config/env/keychain without writing plaintext secrets.
- [x] R001 plan: record bundle archive layout, excluded files, supported platform slice, release-signing assumption, and artifact naming.
- [x] R001 plan: record required MCPB manifest fields, binary server config sketch, compatibility/privacy policy fields, and catalog summary strategy.
- [x] R001 plan: record a user_config-to-env table with sensitivity, required/default behavior, and no-plaintext-config rationale.
- [x] R001 plan: record validation tooling/commands and `.mcpb` versus legacy `.dxt` docs wording.

**Step 1 research notes / decisions**

**MCPB manifest spec research (2026-05-20):** Use MCPB manifest spec `manifest_version: "0.3"`. Required fields are `manifest_version`, `name`, semver `version`, `description`, `author.name`, and `server`. For icuvisor the server must be `server.type: "binary"`; binary extensions are self-contained, need no Node/Python runtime constraints, and support platform-specific executables. `server.entry_point` points at the bundled executable. `server.mcp_config` replaces manual Claude Desktop JSON and should launch the binary with `command: "${__dirname}/server/icuvisor"`, no args, and environment variables for configuration. The spec notes clients automatically append `.exe` on Windows, but Step 3 will still prefer per-platform bundles so each archive contains exactly one signed executable and the manifest `compatibility.platforms` matches that artifact. `${__dirname}` is supported for package-relative paths. `user_config` supports `sensitive: true`; sensitive values are masked/stored securely by the host and are best passed as environment variables. `privacy_policies` is required when external services process user data; include the project privacy/security docs and intervals.icu privacy policy because icuvisor connects to intervals.icu with user data.

**Artifact naming and extension wording:** Release artifacts use the new MCP Bundle extension only: `icuvisor_{{ .Version }}_{{ .Os }}_{{ .Arch }}.mcpb`, with `icuvisor_{{ .Version }}_darwin_universal.mcpb` for the first macOS universal artifact. Docs may say `.dxt` was the former Desktop Extension name and existing `.dxt` bundles still work in compatible clients, but new docs/downloads/releases must call the artifact `.mcpb` and must not publish `.dxt`-only artifacts.

**Credential and configuration mapping:** The bundle will not create, ship, or point `ICUVISOR_CONFIG` at any generated JSON file. It will rely on Claude Desktop/MCPB `user_config.sensitive` storage for the Desktop Extension install path; Claude stores the API key in the OS secret vault and substitutes it at launch. Passing that Desktop-managed secret to icuvisor through `INTERVALS_ICU_API_KEY` is acceptable for this path because no plaintext file is written and it matches icuvisor's existing env-first runtime input. The existing `icuvisor setup` keychain flow remains the manual fallback for non-MCPB installs.

| user_config key | Type | Sensitive | Required/default | Runtime mapping |
|---|---|---:|---|---|
| `api_key` | string | yes | required | `INTERVALS_ICU_API_KEY=${user_config.api_key}` |
| `athlete_id` | string | no | required; accepts `12345` or `i12345` | `INTERVALS_ICU_ATHLETE_ID=${user_config.athlete_id}` |
| `timezone` | string | no | optional default `UTC`; IANA name preferred | `ICUVISOR_TIMEZONE=${user_config.timezone}` |
| `toolset` | string | no | optional default `core`; user may enter `full` | `ICUVISOR_TOOLSET=${user_config.toolset}` |

Non-user env set by manifest: `ICUVISOR_TRANSPORT=stdio`. Do not expose `ICUVISOR_DELETE_MODE` in the first Desktop Extension UI; destructive deletes stay on the runtime default (`safe`, delete tools hidden) unless a power user uses manual configuration outside the bundle.

**Bundle layout, exclusions, platform slice, and signing:** Package layout is `manifest.json`, `server/icuvisor` (or `server/icuvisor.exe` in Windows bundles), `README.md`, `LICENSE`, and `CHANGELOG.md`; include icon assets only if a project-owned permissive asset already exists or is added in scope. Explicitly exclude `.env`, `icuvisor.json`/local config, keychain exports, generated secrets, `.git`, taskplane state, tests, source worktrees, unsigned development binaries, and release credentials. Step 3 will start with the first supported release artifact slice `darwin_universal.mcpb` using the signed/notarized macOS release binary from the existing release pipeline; per-platform Linux/Windows `.mcpb` artifacts may be added when the packaging script can select their signed/release binaries. MCPB packaging must never rebuild an unsigned binary for release.

**Manifest sketch and catalog strategy:** `manifest.json` will include `$schema` if the packaged validator can resolve it, `manifest_version: "0.3"`, `name: "icuvisor"`, `display_name: "icuvisor"`, version placeholder populated by release packaging, short/long descriptions, MIT license, homepage/documentation/support/repository URLs, `author.name`, keywords, `privacy_policies`, `compatibility.platforms` matching the artifact, and `server`. Server sketch: `type: "binary"`, `entry_point: "server/icuvisor"`, `mcp_config.command: "${__dirname}/server/icuvisor"`, `args: []`, `env` mapping listed above, and no runtime constraints. For Windows, rely on per-platform bundles and the MCPB `.exe` behavior rather than shipping multiple binaries in one archive. Because icuvisor's registered tools vary by `ICUVISOR_TOOLSET`, `ICUVISOR_DELETE_MODE`, and coach mode, manifest metadata will use a concise curated `tools` summary plus `tools_generated: true`; prompts will use a concise curated summary plus `prompts_generated: true`; resources are documented in the long description/README because current manifest schema has no first-class resources array.

**Validation tooling:** Later steps will use `npx @anthropic-ai/mcpb@latest validate packaging/mcpb/manifest.json` when supported by the CLI, `npx @anthropic-ai/mcpb@latest pack <staging-dir> <artifact>.mcpb` for local packing, and a local JSON/schema smoke test checked into the repo so CI does not depend solely on a globally installed CLI. The older `@anthropic-ai/dxt` CLI/package exists for legacy `.dxt` flows, but TP-101 will document `mcpb` first and mention `.dxt` only as the former extension name.

---

### Step 2: Create MCPB packaging assets
**Status:** ✅ Complete

- [x] Add `packaging/mcpb/manifest.json` for a binary server using the bundled icuvisor executable.
- [x] Declare user-visible metadata, tools/prompts/resources summary where useful, platform compatibility, icon assets if available, and sensitive config fields.
- [x] R004 plan: record direct-valid manifest strategy, schema-aware field inventory, exact user_config/env mapping, packaging script behavior, README acceptance criteria, and icon decision.
- [x] Add local packaging README/script that validates and packs the bundle with `mcpb pack`.
- [x] R004 plan: ensure the packaging script stages only approved files, fails closed for invalid binary input, and avoids dev secrets/local config.
- [x] R005: add platform/binary-format validation so `ICUVISOR_MCPB_PLATFORM` cannot package the wrong executable type.
- [x] R005: fix the manifest tool summary to use the registered `icuvisor_list_advanced_capabilities` tool name.

**Step 2 implementation plan / R004 response**

- Files: `packaging/mcpb/manifest.json` is a directly valid development manifest, `packaging/mcpb/README.md` documents local packaging, `packaging/mcpb/assets/icon.png` uses the existing project-owned generated/app icon asset, and `scripts/package_mcpb.sh` stages/validates/packs the archive. No `manifest.json.tmpl` is needed.
- Manifest strategy: the checked-in manifest uses schema-valid dev defaults (`version: "0.0.0"`, `compatibility.platforms: ["darwin"]`, `server.entry_point: "server/icuvisor"`). The packaging script copies it into a temporary staging directory, then substitutes the semver release/dev version and platform-specific binary path before validation.
- Field inventory to keep schema-aware: `$schema`, `manifest_version`, `name`, `display_name`, `version`, `description`, `long_description`, `author`, `repository`, `homepage`, `documentation`, `support`, `icon`, `license`, `keywords`, `privacy_policies`, `server`, `tools`, `prompts`, `compatibility`, and `user_config`. Resource detail stays in README/long description unless the MCPB schema accepts a first-class resources array.
- `user_config` and env mapping: `api_key` is the only sensitive required string and maps to `INTERVALS_ICU_API_KEY`; required non-sensitive `athlete_id` maps to `INTERVALS_ICU_ATHLETE_ID`; optional `timezone` defaults to `UTC` and maps to `ICUVISOR_TIMEZONE`; optional `toolset` defaults to `core` and maps to `ICUVISOR_TOOLSET`; the manifest sets `ICUVISOR_TRANSPORT=stdio`. `ICUVISOR_DELETE_MODE` is not exposed in the Desktop Extension UI.
- Packaging script behavior: `scripts/package_mcpb.sh` accepts environment overrides for binary path, version, platform, output path, and MCPB CLI package; creates a temp staging tree with only `manifest.json`, `server/icuvisor` or `server/icuvisor.exe`, `README.md`, `LICENSE`, `CHANGELOG.md`, and the owned icon; sets executable permissions; validates with `npx --yes @anthropic-ai/mcpb@latest validate`; and packs with `mcpb pack`. It must fail closed for missing/non-executable development inputs and must never rebuild the binary or copy `.env`, `icuvisor.json`, taskplane state, `.git`, local config, or generated secrets.
- README acceptance: document local pack commands, signed/release binary expectations, no-secret archive policy, Claude Desktop sensitive `user_config` handling, and minimal install/smoke guidance. Broader public install docs remain Step 4.

---

### Step 3: Integrate with releases
**Status:** ✅ Complete

- [x] Step 3 plan: document release integration approach for the first supported macOS universal MCPB slice.
- [x] Update GoReleaser/workflows/scripts to produce per-platform `.mcpb` artifacts or a documented first supported platform slice.
- [x] Ensure the bundle includes the correct signed binary and no development secrets.
- [x] Add smoke/validation step for manifest schema.

**Step 3 implementation plan**

- Scope the first release-integrated MCPB artifact to macOS universal: `dist/icuvisor_${version}_darwin_universal.mcpb`. Linux/Windows MCPB artifacts remain supported by `scripts/package_mcpb.sh` locally, but are not published until their signed/release-binary selection is wired explicitly.
- In `.github/workflows/release.yml`, add Node setup where the pinned `ICUVISOR_MCPB_CLI_PACKAGE` is used, validate `packaging/mcpb/manifest.json` during release preflight, then package the macOS universal MCPB in the macOS release job after GoReleaser has produced `dist/*_darwin_all/icuvisor`.
- Sign and verify the standalone macOS universal binary before calling `scripts/package_mcpb.sh`; the packaging script stages only approved bundle files and excludes local config/secrets by construction.
- Upload the `.mcpb` to the draft GitHub release and include it in the final release download/checksum regeneration step.

---

### Step 4: Test install and document
**Status:** ✅ Complete

- [x] Step 4 plan: document local install/smoke strategy and docs update scope.
- [x] R009 plan: clarify that non-GUI MCP smoke is supplementary and cannot complete the Claude Desktop install checkbox if GUI confirmation is unavailable.
- [x] Test local installation in Claude Desktop by dragging/opening the `.mcpb` and confirming stdio tool call works.
- [x] R011: record actual Claude Desktop-mediated installed-extension `tools/call` evidence, or mark the install checkbox incomplete with a manual-validation blocker.
- [x] R012: strip trailing whitespace from the committed R011 review artifact so `git diff --check` is clean.
- [x] R013: strip trailing whitespace from the committed R012 review artifact so `git diff --check` is clean.
- [x] Run supplementary packaged-binary MCP stdio smoke with `tools/list` and `icuvisor_list_advanced_capabilities` call using dummy non-secret env.
- [x] Update Claude Desktop install docs with extension-first path plus manual fallback.
- [x] Update CHANGELOG.md.

**Step 4 implementation plan**

- Build a local MCPB with `scripts/package_mcpb.sh`, inspect the archive contents, and attempt the Claude Desktop install path on this macOS worker (`/Applications/Claude.app` is present). Record exact evidence in STATUS: bundle path, Claude version if available, whether the extension config/install UI appeared, and whether a simple tool call could be confirmed. If GUI confirmation cannot be driven from the worker, leave the Claude Desktop install checkbox incomplete and record a blocker/manual-validation requirement rather than counting fallback smoke as completion.
- Independently verify the packaged stdio server path by extracting the bundle and running MCP `initialize`, `tools/list`, and `tools/call` for the no-network `icuvisor_list_advanced_capabilities` tool with dummy/non-secret env.
- Update `web/content/connect/claude-desktop.md` so the extension-first path is primary for the current first supported macOS `.mcpb`: download/open the bundle, enter sensitive `api_key` through Claude Desktop extension config, enter non-secret athlete/timezone/toolset values, and use the existing manual JSON/keychain setup only as fallback.
- Update `CHANGELOG.md` under `[Unreleased]` with the new MCPB packaging/release/docs behavior.

**Step 4 validation notes**

- Built `/tmp/icuvisor_step4.mcpb` with the pinned MCPB CLI and opened it with Claude Desktop 1.7196.3 via `open -a Claude /tmp/icuvisor_step4.mcpb` (exit 0). Claude installed it as `local.mcpb.icuvisor-maintainers.icuvisor` under `~/Library/Application Support/Claude/Claude Extensions/` and recorded it in `extensions-installations.json` with `server.type: binary`, `entry_point: server/icuvisor`, and sensitive `api_key` metadata. The extension settings file contains Desktop-managed encrypted config only; no plaintext API key was observed or copied.
- Operator validation for R011 confirmed Claude Desktop loaded the icuvisor integration tools and called `icuvisor_list_advanced_capabilities` once; it returned 24 capabilities. No network-dependent tools were called.
- Supplementary stdio smoke used the unpacked bundle binary with dummy non-secret env and successfully completed MCP `initialize`, `tools/list`, and `tools/call` for `icuvisor_list_advanced_capabilities`.

---


### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] Targeted tests passing
- [x] FULL test suite passing: `make test`
- [x] Build passes: `make build`
- [x] Lint passes: `make lint`
- [x] All failures fixed or documented as pre-existing unrelated failures

**Step 5 validation notes**

- Targeted checks passed: `npx --yes @anthropic-ai/mcpb@latest validate packaging/mcpb/manifest.json`; `ICUVISOR_MCPB_OUTPUT=/tmp/tp101-targeted.mcpb scripts/package_mcpb.sh`; archive inspection confirmed only `manifest.json`, `server/icuvisor`, `README.md`, `LICENSE`, `CHANGELOG.md`, and `assets/icon.png` were present.
- Full verification passed: `make test`, `make build`, and `make lint` (`golangci-lint run ./...`, 0 issues).
- No test, build, or lint failures required remediation or pre-existing failure documentation.

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
| R005 | code | 2 | REVISE | `.reviews/R005-code-step2.md` |
| R006 | code | 2 | APPROVE | `.reviews/R006-code-step2.md` |
| R007 | plan | 3 | APPROVE | `.reviews/R007-plan-step3.md` |
| R008 | code | 3 | APPROVE | `.reviews/R008-code-step3.md` |
| R009 | plan | 4 | REVISE | `.reviews/R009-plan-step4.md` |
| R010 | plan | 4 | APPROVE | `.reviews/R010-plan-step4.md` |
| R011 | code | 4 | REVISE | `.reviews/R011-code-step4.md` |
| R012 | code | 4 | REVISE | `.reviews/R012-code-step4.md` |
| R013 | code | 4 | REVISE | `.reviews/R013-code-step4.md` |
| R014 | code | 4 | APPROVE | `.reviews/R014-code-step4.md` |
| R015 | plan | 5 | APPROVE | `.reviews/R015-plan-step5.md` |
| R016 | code | 5 | APPROVE | `.reviews/R016-code-step5.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Claude Desktop install can be confirmed by opening the `.mcpb`, but a true Desktop-mediated tool call may require GUI/operator validation in this environment. | Captured as Step 4 validation evidence and used for R011 closure. | STATUS.md Step 4 validation notes |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 15:23 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 15:23 | Step 0 started | Preflight |
| 2026-05-20 16:14 | Worker iter 1 | done in 3047s, tools: 99 |
| 2026-05-20 16:46 | Worker iter 2 | done in 1910s, tools: 133 |
| 2026-05-20 16:46 | Step 5 started | Testing & Verification |
| 2026-05-20 17:19 | Worker iter 3 | done in 1977s, tools: 113 |
| 2026-05-20 17:19 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

**Step 6 delivery notes**

- Must Update docs modified: `CHANGELOG.md` records the MCPB packaging/release/docs behavior under `[Unreleased]`; `STATUS.md` records validation evidence, review outcomes, and final delivery state.
- Check If Affected docs reviewed: `README.md` has no Claude Desktop/MCPB setup text; `web/content/reference/tools.md` is unaffected because no tool catalog schema/descriptions changed; `docs/prd/PRD-icuvisor.md` was updated from legacy DXT wording to MCPB (formerly DXT) for the Claude Desktop bundle entry.

*Reserved for execution notes*
| 2026-05-20 15:26 | Review R001 | plan Step 1: REVISE |
| 2026-05-20 15:27 | Review R002 | plan Step 1: APPROVE |
| 2026-05-20 15:31 | Review R003 | code Step 1: UNAVAILABLE (reviewer exited with no output; proceeding cautiously) |
| 2026-05-20 15:44 | Review R004 | plan Step 2: REVISE; plan hydrated in Step 2 |
| 2026-05-20 16:21 | Review R005 | code Step 2: REVISE; platform binary validation and manifest tool-name fix required |
| 2026-05-20 16:26 | Review R006 | code Step 2: APPROVE |
| 2026-05-20 16:29 | Review R007 | plan Step 3: APPROVE |
| 2026-05-20 16:34 | Review R008 | code Step 3: APPROVE |
| 2026-05-20 16:38 | Review R009 | plan Step 4: REVISE; install evidence and supplementary tool-call smoke clarified |
| 2026-05-20 16:40 | Review R010 | plan Step 4: APPROVE |
| 2026-05-20 16:36 | Review R009 | plan Step 4: REVISE |
| 2026-05-20 16:37 | Review R010 | plan Step 4: APPROVE |
| 2026-05-20 16:44 | Review R011 | code Step 4: REVISE |
| 2026-05-20 16:58 | Review R012 | code Step 4: REVISE |
| 2026-05-20 17:01 | Review R013 | code Step 4: REVISE |
| 2026-05-20 17:05 | Review R014 | code Step 4: APPROVE |
| 2026-05-20 17:08 | Review R015 | plan Step 5: APPROVE |
| 2026-05-20 17:14 | Review R016 | code Step 5: APPROVE |
