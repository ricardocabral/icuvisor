# Review R004 — Plan Review for Step 2

**Verdict:** Changes requested

Step 1 now records useful decisions, but the Step 2 plan itself is still only the three high-level task checkboxes. For this task, that is not specific enough before creating packaging assets: the manifest, user configuration, and local pack script are the parts most likely to introduce install failures or secret-handling regressions.

## Blocking gaps

1. **Checked-in manifest strategy is not concrete enough.**
   - Decide whether `packaging/mcpb/manifest.json` is a directly valid manifest or a template expanded by the packaging script. The task explicitly names `manifest.json`; if it contains release placeholders like `{{ .Version }}`, local validation will fail unless the script/templates are part of the plan.
   - Record the exact dev/default version strategy, e.g. `0.0.0-dev` plus script-time replacement, or `manifest.json.tmpl` plus generated staging `manifest.json`.
   - Record the first `compatibility.platforms` value for Step 2/local packaging. Step 1 says Step 3 starts with `darwin_universal`, but Step 2 must still know what platform a local bundle is validating against.

2. **Manifest field list needs schema-aware confirmation.**
   - The plan should list the exact top-level fields to add and whether each is known to be accepted by the MCPB schema. In particular, fields like `tools_generated`, `prompts_generated`, `privacy_policies`, `compatibility`, and any resource summary should be verified against the manifest spec before being committed.
   - If the schema does not support a first-class resources array or arbitrary extra fields, put resource information in `README.md`/description instead of risking a manifest that `mcpb validate` rejects.

3. **`user_config` interpolation needs exact syntax and constraints.**
   - Record the exact `user_config` definitions for `api_key`, `athlete_id`, `timezone`, and `toolset`: type, title/description, required/default, enum or allowed values where applicable, and `sensitive: true` only for `api_key`.
   - Record the exact `server.mcp_config.env` interpolation syntax that will be used, e.g. mapping to `INTERVALS_ICU_API_KEY`, `INTERVALS_ICU_ATHLETE_ID`, `ICUVISOR_TIMEZONE`, `ICUVISOR_TOOLSET`, and `ICUVISOR_TRANSPORT=stdio`.
   - Preserve the Step 1 decision not to expose `ICUVISOR_DELETE_MODE` in the Desktop Extension UI.

4. **Packaging script behavior must be planned before implementation.**
   - Name the script/path to be added, likely under `scripts/`, and define its inputs (`--binary`, `--version`, `--platform`, `--arch` or equivalent) and output artifact path.
   - The script should stage into a temporary directory, copy only the approved bundle files (`manifest.json`, `server/icuvisor` or `.exe`, `README.md`, `LICENSE`, `CHANGELOG.md`, optional owned icon), set executable permissions, validate the manifest, then run `mcpb pack`.
   - It must fail closed if the binary path is missing, not executable, or looks like a development placeholder; it must not rebuild icuvisor for release packaging and must not package `.env`, `icuvisor.json`, taskplane state, `.git`, local config, or generated secrets.
   - Add a clear fallback/error message if `npx @anthropic-ai/mcpb@latest` or the chosen MCPB CLI is unavailable.

5. **Local README acceptance criteria are not specified.**
   - The Step 2 plan should say what `packaging/mcpb/README.md` will contain: local pack command, required signed/release binary input, no-secret policy, how Claude Desktop stores sensitive `user_config`, and a minimal smoke command/installation note.
   - The README should not replace the existing manual Claude Desktop/keychain docs; that broader user-facing docs update belongs in Step 4.

6. **Icon/assets decision is still open.**
   - Either identify an existing project-owned permissively licensed icon to include, or explicitly plan to omit icons for Step 2. Do not add a downloaded/unverified asset just to satisfy the metadata checkbox.

## Recommended additions to `STATUS.md` before coding

Add a short "Step 2 implementation plan" block with:

- The exact files to create (`packaging/mcpb/manifest.json`, `packaging/mcpb/README.md`, script path, optional template path if used).
- Direct-valid-manifest vs template decision and version/platform substitution behavior.
- Exact manifest field inventory and schema validation command.
- Exact `user_config` and `mcp_config.env` mapping.
- Packaging script staging layout, inputs, exclusions, and failure behavior.
- Icon decision.
- Minimal validation commands to run after Step 2, even if full release integration remains Step 3.

Once these details are recorded, Step 2 should be safe to implement without re-opening the credential and packaging-shape decisions during code review.
