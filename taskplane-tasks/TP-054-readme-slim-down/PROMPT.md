# TP-054 — README slim-down and `docs/` cleanup

## Mission

With end-user content now hosted at icuvisor.app (TP-050 → TP-053), rewrite `README.md` to serve **developers and contributors only**, delete the migrated user-facing files from `docs/`, and update every inbound link. After this task, there is no end-user prose in the repo's top-level docs that is also on the website.

Project context: icuvisor is private and pre-launch. No external links to preserve. No need for redirect stubs.

PRD anchors: §7.4 (project principles), §6 KR1 (clear install path — now lives on website).

ROADMAP positioning: docs polish for v0.5 / v1.0.

Complexity: Blast radius 2 (top-level repo docs; affects every GitHub visitor), Pattern novelty 1, Security 1, Reversibility 2 = 6 → Review Level 2. Size: S.

## Dependencies

- **TP-050, TP-051, TP-052, TP-053, TP-055** must all have landed. All blocking. This task is the final cleanup; do not start it while any predecessor is open.

## Context to Read First

- `CLAUDE.md`
- Current `README.md` (every section)
- The migrated content on the live site after TP-052 — confirm every section being deleted from the README has a real, complete destination.
- `CONTRIBUTING.md`, `SECURITY.md`, `ROADMAP.md`, `docs/prd/PRD-icuvisor.md` — link targets for the slimmed README.
- `CHANGELOG.md`

## File Scope

Expected files:

- `README.md` — rewritten.
- `docs/install/macos.md` — **delete**.
- `docs/clients/claude-desktop.md` — **delete**.
- `docs/clients/claude-code.md` — **delete**.
- `docs/coach-mode.md` — **delete**.
- `docs/post-update.md` — **delete**.
- `docs/clients/codex-local.md` — keep (developer content).
- Every other `docs/**` file — keep.
- `web/README.md` — keep, but update if local-preview commands have changed since TP-050.
- Any other file that links to the deleted `docs/*.md` paths — update the link target.
- `taskplane-tasks/TP-054-readme-slim-down/STATUS.md`.
- `CHANGELOG.md` under `[Unreleased]`.

## Target README structure (~60–90 lines)

```markdown
# icuvisor

[badges]

> One-paragraph pitch + link to https://icuvisor.app for end-user docs.

## Status

Pre-alpha. See [ROADMAP.md](ROADMAP.md).

## For users

Install, connect your AI assistant, and read the tool catalog at <https://icuvisor.app>.

## For developers

### Build from source

[the existing `git clone … && make build` block]

### Project layout

[the existing tree, unchanged]

### Development

[`make build`, `make test`, `make test-race`, `make lint`, `make snapshot`, `make docs-tools`, `make help`]

Requires Go 1.23+ and (optionally) `golangci-lint`, `goreleaser`.

The KR5 benchmark harness is documented in [`docs/kr5-benchmark.md`](docs/kr5-benchmark.md).

See [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md), and the [PRD](docs/prd/PRD-icuvisor.md).

## Acknowledgements

[unchanged]

## License

[unchanged]
```

Sections to **delete** from the current README:

- §Features (planned for v1.0) — lives on website landing.
- §"MCP tool catalog" — generated and rendered on website (TP-051).
- §"MCP resources" — lives at `/reference/resources-prompts/`.
- §"MCP prompts" — same.
- §Install (the macOS DMG steps) — lives at `/install/macos/`.
- §Quickstart — lives at `/install/` and `/guides/api-key/`.
- §"Getting an API key" — lives at `/guides/api-key/`.
- §"MCP transport" — split between `/guides/http-transport/` and `/reference/cli/`.
- §"Delete/write safety mode" — split between `/reference/safety-modes/` and `/explain/safety-modes/`.
- §"Toolset tiers" — same two destinations.
- The line "After upgrading icuvisor, open a new AI-client conversation…" — lives at `/guides/after-upgrade/`.

Sections to **keep**:

- Title + badges
- One-paragraph pitch
- Status (pre-alpha, link to ROADMAP)
- For users (one paragraph → icuvisor.app)
- For developers (build from source, project layout, development targets, KR5 benchmark, pointers to CONTRIBUTING/SECURITY/PRD)
- Acknowledgements
- License

## Steps

### Step 1: Pre-flight verification

- [ ] Confirm every section being deleted has a complete website destination. For each deletion, paste the destination URL into `STATUS.md` Step 1.
- [ ] Confirm TP-051 has the tool-catalog generator + website page live. If not, hold this task.
- [ ] Confirm TP-052 has migrated every end-user `docs/*.md` page. If not, hold this task.
- [ ] Confirm TP-053 has shipped the ChatGPT tutorial. If not, hold this task.
- [ ] Confirm TP-055 has landed. If not, **stop** — do not slim the README while the three known PRD/ROADMAP/README conflicts are unresolved.

### Step 2: Inbound link sweep

- [ ] `git grep -nE 'docs/(install/macos|clients/claude-desktop|clients/claude-code|coach-mode|post-update)\.md'` — fix every hit. Most live in `README.md` itself; some may be in `CONTRIBUTING.md`, `SECURITY.md`, `CHANGELOG.md`, `ROADMAP.md`, and `docs/prd/PRD-icuvisor.md`.
- [ ] Replace each match with either (a) the new icuvisor.app URL if the inbound context is end-user, or (b) deletion if the link was already redundant.

### Step 3: Rewrite README

- [ ] Replace the file with the target structure above.
- [ ] Keep all existing badges.
- [ ] Keep the Project layout block verbatim — it is developer-facing reference.
- [ ] Verify Conventional Commits / no-emoji / etc. style requirements from CLAUDE.md.

### Step 4: Delete migrated `docs/*.md` files

- [ ] `rm docs/install/macos.md docs/clients/claude-desktop.md docs/clients/claude-code.md docs/coach-mode.md docs/post-update.md`
- [ ] `rmdir docs/install docs/clients` only if those directories are now empty. `docs/clients/codex-local.md` should still be present — do not remove that directory.

### Step 5: Verify

- [ ] `git grep -n 'docs/install/macos\|docs/clients/claude-desktop\|docs/clients/claude-code\|docs/coach-mode\|docs/post-update'` — zero hits across the whole repo.
- [ ] `make build`, `make test`, `make lint` — no regressions (Go code untouched, but CI may include doc-link checks).
- [ ] `cd web && hugo --minify --gc` — the site still builds.
- [ ] Open the rewritten README in a Markdown previewer; verify badges render and links work.

## Acceptance Criteria

- README is ≤ ~100 lines (give or take), developer-focused, with a clear "users go to icuvisor.app" pointer at the top.
- All end-user `docs/*.md` files now duplicated on the website are deleted.
- Zero inbound links remain pointing to the deleted `docs/*.md` paths anywhere in the repo (including PRD, ROADMAP, CHANGELOG, CONTRIBUTING).
- The Hugo site still builds (`hugo --minify --gc`).
- Go test/lint pipeline is unaffected.
- `STATUS.md` includes the pre-flight verification table (one row per deleted section → website destination URL).
- `CHANGELOG.md` `[Unreleased]` notes the docs reorganization under "Changed".

## Do NOT

- Do not delete `docs/clients/codex-local.md`, `docs/prd/`, `docs/dogfood/`, `docs/internal-beta/`, `docs/safety/`, `docs/threat-models/`, `docs/upstream-gaps/`, `docs/kr5-benchmark.md`. They are developer content.
- Do not delete top-level `CONTRIBUTING.md`, `SECURITY.md`, `ROADMAP.md`, `CHANGELOG.md`, `LICENSE`.
- Do not leave broken inbound links anywhere in the repo. If unsure, search.
- Do not introduce emojis (per CLAUDE.md).
- Do not amend the PRD or ROADMAP factually in this task — TP-055 owns that.
- Do not "tidy up" unrelated docs while you are here. Keep the change focused.

## Documentation

Must update:

- `STATUS.md` — pre-flight verification table; final link-sweep grep output.
- `CHANGELOG.md` `[Unreleased]`.

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-054`, e.g., `TP-054 rewrite readme for developer audience`.

---

## Amendments

_Add amendments below this line only._
