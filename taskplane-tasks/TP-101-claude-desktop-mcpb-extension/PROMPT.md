# TP-101 — ICUvisor Desktop Extension (.mcpb) for Claude Desktop

**Created:** 2026-05-20
**Size:** L

## Review Level: 3

**Assessment:** New distribution artifact with secret-handling and release-pipeline impact; full review is required.
**Score:** 7/8 — Blast radius: 2, Pattern novelty: 2, Security: 2, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-101-claude-desktop-mcpb-extension/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Package icuvisor as a Claude Desktop Extension (`.mcpb`, formerly `.dxt`) so Claude Desktop users can install the local MCP server with a double-click/install flow instead of editing JSON. The bundle should include the signed icuvisor binary, a valid `manifest.json`, sensitive user configuration handled by Claude/Desktop Extension keychain support, and release/docs integration.

## Dependencies

- **Task:** TP-037 (macOS signed installer/release context exists)
- **Task:** TP-078 (keychain-backed onboarding behavior is aligned)
- **External:** Review Anthropic Desktop Extensions/MCPB spec before implementation

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- Anthropic engineering post: https://www.anthropic.com/engineering/desktop-extensions — `.mcpb` overview, binary server type, user_config, packaging/testing flow.
- MCPB spec/tooling: the relevant upstream project documentation and examples — manifest schema and packaging commands.
- .goreleaser.yaml — current release artifact pipeline.
- web/content/connect/claude-desktop.md — current manual Claude Desktop setup docs.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `packaging/mcpb/manifest.json`
- `packaging/mcpb/README.md`
- `packaging/mcpb/**/*`
- `.goreleaser.yaml`
- `.github/workflows/*`
- `scripts/*`
- `web/content/connect/claude-desktop.md`
- `web/content/install/*`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Research and decide bundle shape

- [ ] Read the MCPB manifest spec and examples; record required fields and binary-server support in STATUS.md.
- [ ] Decide artifact naming (`.mcpb`) and whether to also mention legacy `.dxt` compatibility.
- [ ] Decide how user_config maps to icuvisor config/env/keychain without writing plaintext secrets.

### Step 2: Create MCPB packaging assets

- [ ] Add `packaging/mcpb/manifest.json` for a binary server using the bundled icuvisor executable.
- [ ] Declare user-visible metadata, tools/prompts/resources summary where useful, platform compatibility, icon assets if available, and sensitive config fields.
- [ ] Add local packaging README/script that validates and packs the bundle with `mcpb pack`.

### Step 3: Integrate with releases

- [ ] Update GoReleaser/workflows/scripts to produce per-platform `.mcpb` artifacts or a documented first supported platform slice.
- [ ] Ensure the bundle includes the correct signed binary and no development secrets.
- [ ] Add smoke/validation step for manifest schema.

### Step 4: Test install and document

- [ ] Test local installation in Claude Desktop by dragging/opening the `.mcpb` and confirming stdio tool call works.
- [ ] Update Claude Desktop install docs with extension-first path plus manual fallback.
- [ ] Update CHANGELOG.md.

### Step 5: Testing & Verification

- [ ] Run targeted tests added/affected by this task
- [ ] Run FULL test suite: `make test`
- [ ] Build passes: `make build`
- [ ] Lint passes: `make lint`
- [ ] Fix all failures or document pre-existing unrelated failures in STATUS.md

### Step 6: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] `STATUS.md` discoveries and final notes updated
- [ ] Commit at step boundary with the task ID in the message

## Documentation Requirements

**Must Update:**
- CHANGELOG.md — record user-visible behavior under [Unreleased] if code or docs behavior changes.
- STATUS.md — keep execution state current.

**Check If Affected:**
- README.md — update if public setup/tool behavior changes.
- web/content/reference/tools.md — update if tool catalog descriptions or generated docs are affected.
- docs/prd/PRD-icuvisor.md — check only if behavior intentionally diverges from product scope.

## Completion Criteria

- A valid MCPB/Desktop Extension bundle can be built locally.
- Claude Desktop can install the bundle and launch icuvisor over stdio.
- Sensitive API key/user_config is stored securely and not written plaintext by icuvisor packaging.
- Release pipeline can produce or validate the bundle.
- `make test`, `make build`, `make lint`, and MCPB validation pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-101` for traceability. Examples:

- `feat(TP-101): complete step 1 — scope current behavior`
- `fix(TP-101): repair regression found during analyzer tests`
- `test(TP-101): add golden coverage for roadmap behavior`
- `hydrate: TP-101 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not call the new artifact `.dxt` only; Anthropic recommends `.mcpb` for new extensions.
- Do not bundle API keys or generated local config into the extension archive.
- Do not drop manual Claude Desktop config docs.

---

## Amendments

_Add amendments below this line only._
