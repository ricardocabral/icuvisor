# R005 code review — Step 2: Create MCPB packaging assets

Verdict: REVISE

I reviewed the diff from `6734aab..HEAD`, read the changed packaging files, and smoke-tested the local pack script.

Validation run:

- `python3 -m json.tool packaging/mcpb/manifest.json` passes.
- `ICUVISOR_MCPB_OUTPUT=/tmp/icuvisor_test.mcpb scripts/package_mcpb.sh` passes MCPB validation/packing and produces an archive containing only `manifest.json`, `server/icuvisor`, `assets/icon.png`, `README.md`, `LICENSE`, and `CHANGELOG.md`.

Findings:

## 1. Platform override can silently package the wrong binary

**Severity:** High  
**Files:** `scripts/package_mcpb.sh:8-13`, `scripts/package_mcpb.sh:41-51`, `scripts/package_mcpb.sh:103-110`

The script accepts `ICUVISOR_MCPB_PLATFORM=linux` or `win32` and rewrites the manifest compatibility/entry point, but it never checks whether `ICUVISOR_MCPB_BINARY` is actually for that target platform. The only binary validation is existence, executable bit for non-Windows, and size. On this macOS worktree I was able to produce a Linux-declared bundle using the default Mach-O `bin/icuvisor`:

```sh
ICUVISOR_MCPB_PLATFORM=linux \
ICUVISOR_MCPB_OUTPUT=/tmp/icuvisor_linux_bad.mcpb \
  scripts/package_mcpb.sh
# exits 0; manifest says compatibility.platforms=["linux"] but server/icuvisor is the macOS Mach-O binary
```

That violates the Step 2 acceptance note that the script should “fail closed for invalid binary input” and creates a release foot-gun for Step 3: a mis-set platform variable can publish an installable but non-launchable Desktop Extension.

Suggested fix: add an explicit platform/binary-format check before staging or before packing. For example, inspect `file "$BINARY_PATH"` and require Mach-O for `darwin`, ELF for `linux`, and PE/Windows executable for `win32`; or require a target-specific binary path convention and fail if the declared platform does not match. If universal macOS is required, also consider checking the architecture/slice in the signed binary when Step 3 wires release artifacts.

## 2. Manifest advertises a non-existent tool name

**Severity:** Medium  
**File:** `packaging/mcpb/manifest.json:82-85`

The manifest lists:

```json
{"name": "list_advanced_capabilities", ...}
```

but the registered/canonical tool name is `icuvisor_list_advanced_capabilities` (`internal/toolcatalog/catalog.go`, `internal/tools/list_advanced_capabilities.go`). This makes the Desktop Extension metadata inconsistent with the runtime catalog. If Claude Desktop surfaces manifest tools during install/search, users will see a tool name that cannot be called.

Suggested fix: change the manifest entry to `icuvisor_list_advanced_capabilities`, or generate/check the manifest summary against `internal/toolcatalog` so this cannot drift.

