# icuvisor MCPB packaging

This directory contains the Claude Desktop Extension (MCPB) manifest and bundle-only README for icuvisor.

## Bundle contents

`scripts/package_mcpb.sh` stages a temporary bundle with only these files:

- `manifest.json` with a release/version-specific binary server entry
- `server/icuvisor` or `server/icuvisor.exe`
- `assets/icon.png`
- `README.md`
- `LICENSE`
- `CHANGELOG.md`

The script does not copy source files, `.git`, taskplane state, `.env`, `icuvisor.json`, keychain exports, local configuration, generated secrets, or release credentials.

## Build a local bundle

Build or supply an icuvisor binary first. For local smoke testing:

```sh
make build
scripts/package_mcpb.sh
```

For release packaging, point the script at the signed/notarized release binary instead of rebuilding:

```sh
ICUVISOR_MCPB_BINARY=/path/to/signed/icuvisor \
ICUVISOR_MCPB_VERSION=0.1.0 \
ICUVISOR_MCPB_PLATFORM=darwin \
ICUVISOR_MCPB_OUTPUT=dist/icuvisor_0.1.0_darwin_universal.mcpb \
scripts/package_mcpb.sh
```

Supported `ICUVISOR_MCPB_PLATFORM` values are `darwin`, `linux`, and `win32`. The first documented release slice is the macOS universal bundle (`darwin_universal` artifact name) using the signed macOS binary produced by the release pipeline.

The script validates the staged manifest and then packs the archive with:

```sh
npx --yes @anthropic-ai/mcpb@latest validate <staged-manifest>
npx --yes @anthropic-ai/mcpb@latest pack <staging-dir> <output.mcpb>
```

Override the CLI package with `ICUVISOR_MCPB_CLI_PACKAGE` only for debugging or pinned release automation.

## Secret handling

The bundle never contains an intervals.icu API key and does not generate plaintext icuvisor config. `manifest.json` marks `api_key` as `sensitive: true`, so Claude Desktop stores the value in its secure extension configuration and passes it to icuvisor as `INTERVALS_ICU_API_KEY` only when launching the local stdio server.

Other user configuration maps to environment variables at launch:

| User field | Environment variable | Notes |
| --- | --- | --- |
| `api_key` | `INTERVALS_ICU_API_KEY` | Required and sensitive |
| `athlete_id` | `INTERVALS_ICU_ATHLETE_ID` | Required; accepts `12345` or `i12345` |
| `timezone` | `ICUVISOR_TIMEZONE` | Optional; defaults to `UTC` |
| `toolset` | `ICUVISOR_TOOLSET` | Optional; defaults to `core` |

`ICUVISOR_TRANSPORT=stdio` is set by the manifest. Delete-mode configuration is intentionally not exposed in this Desktop Extension UI.

## Local install smoke test

After creating a bundle, open it with Claude Desktop or drag it into the Extensions UI. Enter the intervals.icu API key and athlete ID when prompted, then confirm Claude can list the icuvisor tools or run a simple athlete-profile request.

The extension-first user documentation lives in `web/content/connect/claude-desktop.md`. Keep the manual JSON/keychain setup path there as the fallback for users who do not install the MCPB bundle.
