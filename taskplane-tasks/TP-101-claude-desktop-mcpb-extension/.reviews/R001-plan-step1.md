# Review R001 — Plan Review for Step 1

**Verdict:** Changes requested

The current Step 1 plan is only the three task checkboxes repeated in `STATUS.md`; it does not yet state the actual research outputs or decisions that Step 1 is supposed to lock before packaging work starts. Because this task affects credential handling and release artifacts, Step 2 should not begin until the Step 1 section in `STATUS.md` records concrete decisions.

## Blocking gaps

1. **Bundle shape is not decided.**
   - Record the planned archive layout, e.g. `manifest.json`, `server/icuvisor`/`server/icuvisor.exe`, `README.md`, `LICENSE`, and explicitly excluded files (`.env`, local config JSON, generated secrets, dev artifacts).
   - Decide whether Step 3 will produce one first supported artifact, such as macOS universal, or per-platform artifacts. This matters because MCPB `binary` servers are platform-specific unless the bundle contains platform overrides/multiple binaries.

2. **Manifest requirements need to be captured in STATUS before implementation.**
   The MCPB manifest spec currently requires at least:
   - `manifest_version`
   - `name`
   - `version`
   - `description`
   - `author.name`
   - `server`

   For this project, the plan should also explicitly decide:
   - `server.type: "binary"`
   - `server.entry_point`
   - `server.mcp_config.command` pathing, including whether to use `${__dirname}` and whether Windows needs `platform_overrides` or relies on automatic `.exe` handling.
   - `compatibility.platforms` values (`darwin`, `win32`, `linux`) matching the selected artifact slice.
   - `privacy_policies`, because the extension connects to intervals.icu and the spec says these are required when external services process user data.

3. **Credential mapping is not specific enough.**
   The plan must decide exactly how Desktop Extension `user_config` maps into icuvisor runtime configuration without writing plaintext secrets:
   - Which `user_config` keys exist (`api_key`, `athlete_id`, `timezone`, possibly `toolset`/`delete_mode` if intentionally exposed).
   - Which ones are sensitive and required.
   - Which environment variables they map to (`INTERVALS_ICU_API_KEY`, `INTERVALS_ICU_ATHLETE_ID`, `ICUVISOR_TIMEZONE`, `ICUVISOR_TRANSPORT=stdio`, etc.).
   - Why passing a sensitive Desktop-managed `user_config` value via environment is acceptable relative to the repository rule that API keys live in keychain and are never written to plaintext config. If the decision is to rely on Claude Desktop's secure user_config/keychain storage rather than icuvisor's own OS keychain setup path, that exception/interpretation should be written down clearly.
   - Confirm the bundle will not create or include an `ICUVISOR_CONFIG` file containing `api_key`.

4. **Artifact naming and legacy `.dxt` messaging need a concrete decision.**
   The plan should state the naming pattern, e.g. `icuvisor_<version>_<platform>_<arch>.mcpb`, and whether docs mention that `.dxt` is the former/legacy name while release artifacts use `.mcpb` only. This is explicitly required by the task prompt.

5. **Dynamic catalog behavior should be planned.**
   icuvisor's tool catalog varies by `ICUVISOR_TOOLSET` and delete-mode gating. The Step 1 decision should say whether the manifest will use `tools_generated: true` / `prompts_generated: true`, a static curated summary, or a generated list from the registry. Avoid hard-coding a stale full tool list unless the plan includes a catalog-generation/validation strategy.

## Recommended additions to Step 1 output

Add a short "Step 1 research notes / decisions" block to `STATUS.md` with at least:

- MCPB spec version selected, and required/optional fields relevant to icuvisor.
- Bundle layout and supported platform slice for the first implementation.
- Manifest server config sketch for the binary.
- `user_config` table with key, type, sensitivity, required/default, and target env var.
- Artifact naming convention and `.dxt` compatibility wording.
- Release-signing assumption: the MCPB must package the already signed binary, not an unsigned rebuild.
- Validation command/tooling to be used in later steps (`mcpb pack`/manifest validation), plus any availability caveat.

Once those decisions are recorded, Step 1 will be reviewable and Step 2 can proceed with a much lower risk of rework or accidental secret handling regressions.
