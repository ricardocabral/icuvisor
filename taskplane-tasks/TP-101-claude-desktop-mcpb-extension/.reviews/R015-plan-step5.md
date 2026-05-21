# R015 Plan Review — Step 5

Verdict: REVISE

## Findings

### 1. Step 5 plan is too generic for the MCPB/release artifact risk

The current Step 5 section only repeats the broad checklist (`targeted tests`, `make test`, `make build`, `make lint`). For this task, that does not verify the highest-risk changes: MCPB manifest validity, package script behavior, archive contents/secret exclusion, release config integration, and the web/docs updates. The task completion criteria explicitly require MCPB validation to pass, and previous steps changed `packaging/mcpb/manifest.json`, `scripts/package_mcpb.sh`, `.github/workflows/release.yml`, `.goreleaser.yaml`, and Claude Desktop docs.

Please hydrate the Step 5 plan in `STATUS.md` before executing it with an explicit targeted verification matrix. At minimum it should include:

- `git diff --check` because this task already had whitespace regressions in Step 4 reviews.
- MCPB manifest validation using the same pinned CLI as release automation, e.g. `ICUVISOR_MCPB_CLI_PACKAGE=@anthropic-ai/mcpb@2.1.2 npx --yes "$ICUVISOR_MCPB_CLI_PACKAGE" validate packaging/mcpb/manifest.json` or an equivalent pinned invocation.
- Shell/script sanity for `scripts/package_mcpb.sh` (`bash -n`; `shellcheck` if available, otherwise note unavailable).
- Local bundle exercise from a freshly built binary: `make build`, then `ICUVISOR_MCPB_CLI_PACKAGE=@anthropic-ai/mcpb@2.1.2 ICUVISOR_MCPB_OUTPUT=/tmp/icuvisor_step5.mcpb scripts/package_mcpb.sh`.
- Archive inspection of the generated `.mcpb`: confirm the approved files are present (`manifest.json`, `server/icuvisor`, `assets/icon.png`, `README.md`, `LICENSE`, `CHANGELOG.md`) and forbidden files are absent (`.env`, `icuvisor.json`, `.git`, taskplane state, keychain/config exports, secrets). Also confirm the packaged manifest has the expected platform/version/server command after script substitution.
- A no-network packaged-binary MCP stdio smoke after extracting the bundle, at least `initialize`, `tools/list`, and `tools/call` for `icuvisor_list_advanced_capabilities`, using dummy non-secret env.
- Release/config checks for the files touched in Step 3, at least `goreleaser check` (or document if the binary is unavailable) and any available workflow/YAML validation.
- Docs/site verification because Step 4 changed web docs, e.g. `make web-build` or `cd web && hugo`.
- The required repo gates: `make test`, `make build`, and `make lint`.

### 2. Failure documentation criteria need to be explicit

The plan should state that a checkbox is only marked complete when the exact command passed. If a command cannot run because a local dependency is missing or because the failure is believed pre-existing/unrelated, record the command, relevant output, environment reason, and disposition in `STATUS.md` without marking it as passed. This is especially important for `npx`/MCPB CLI, `golangci-lint`, `goreleaser`, and Hugo, which may not be installed in every worker environment.

## Suggested Step 5 status hydration

Add a short subsection like `Step 5 implementation plan` with the commands above, grouped as:

1. formatting/static smoke (`git diff --check`, `bash -n`, optional `shellcheck`),
2. MCPB validation/package/archive inspection,
3. packaged stdio smoke,
4. release/docs validation (`goreleaser check`, `make web-build`),
5. full repo gates (`make test`, `make build`, `make lint`).

After that hydration, the plan should be sufficient to execute Step 5.
