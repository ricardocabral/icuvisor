#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MANIFEST_TEMPLATE="$ROOT_DIR/packaging/mcpb/manifest.json"
BUNDLE_README="$ROOT_DIR/packaging/mcpb/README.md"
ICON_SOURCE="$ROOT_DIR/packaging/mcpb/assets/icon.png"
BINARY_PATH="${ICUVISOR_MCPB_BINARY:-$ROOT_DIR/bin/icuvisor}"
PLATFORM="${ICUVISOR_MCPB_PLATFORM:-darwin}"
CLI_PACKAGE="${ICUVISOR_MCPB_CLI_PACKAGE:-@anthropic-ai/mcpb@latest}"

case "$PLATFORM" in
  darwin|linux|win32) ;;
  *)
    echo "unsupported ICUVISOR_MCPB_PLATFORM=$PLATFORM; use darwin, linux, or win32" >&2
    exit 2
    ;;
esac

if [ ! -f "$MANIFEST_TEMPLATE" ]; then
  echo "missing manifest template: $MANIFEST_TEMPLATE" >&2
  exit 1
fi

if [ ! -f "$BUNDLE_README" ]; then
  echo "missing bundle README: $BUNDLE_README" >&2
  exit 1
fi

if [ ! -f "$ICON_SOURCE" ]; then
  echo "missing bundle icon: $ICON_SOURCE" >&2
  exit 1
fi

if [ ! -f "$BINARY_PATH" ]; then
  echo "missing icuvisor binary: $BINARY_PATH" >&2
  echo "build one first for local testing, e.g. make build, or set ICUVISOR_MCPB_BINARY=/path/to/signed/icuvisor" >&2
  exit 1
fi

if [ "$PLATFORM" != "win32" ] && [ ! -x "$BINARY_PATH" ]; then
  echo "icuvisor binary is not executable: $BINARY_PATH" >&2
  exit 1
fi

BINARY_SIZE="$(wc -c < "$BINARY_PATH" | tr -d ' ')"
if [ "$BINARY_SIZE" -lt 1048576 ]; then
  echo "icuvisor binary is unexpectedly small ($BINARY_SIZE bytes): $BINARY_PATH" >&2
  echo "refusing to package a placeholder or invalid development file" >&2
  exit 1
fi

if ! command -v file >/dev/null 2>&1; then
  echo "file(1) is required to verify ICUVISOR_MCPB_PLATFORM matches the binary format" >&2
  exit 1
fi

BINARY_FORMAT="$(file -b "$BINARY_PATH" 2>/dev/null || true)"
case "$PLATFORM" in
  darwin)
    case "$BINARY_FORMAT" in
      *Mach-O*) ;;
      *)
        echo "ICUVISOR_MCPB_PLATFORM=darwin requires a Mach-O icuvisor binary, got: $BINARY_FORMAT" >&2
        exit 1
        ;;
    esac
    ;;
  linux)
    case "$BINARY_FORMAT" in
      *ELF*) ;;
      *)
        echo "ICUVISOR_MCPB_PLATFORM=linux requires an ELF icuvisor binary, got: $BINARY_FORMAT" >&2
        exit 1
        ;;
    esac
    ;;
  win32)
    case "$BINARY_FORMAT" in
      *PE32*|*MS\ Windows*) ;;
      *)
        echo "ICUVISOR_MCPB_PLATFORM=win32 requires a PE/Windows icuvisor binary, got: $BINARY_FORMAT" >&2
        exit 1
        ;;
    esac
    ;;
esac

if ! command -v npx >/dev/null 2>&1; then
  echo "npx is required to run $CLI_PACKAGE validate/pack; install Node.js/npm or set up the MCPB CLI first" >&2
  exit 1
fi

raw_version="${ICUVISOR_MCPB_VERSION:-}"
if [ -z "$raw_version" ]; then
  raw_version="$(cd "$ROOT_DIR" && git describe --tags --dirty --always 2>/dev/null || true)"
fi
VERSION="$(python3 - "$raw_version" <<'PY'
import re, sys
raw = (sys.argv[1] or "").strip()
if raw.startswith("v"):
    raw = raw[1:]
if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)?", raw):
    raw = "0.0.0-dev"
print(raw)
PY
)"

artifact_platform="$PLATFORM"
if [ "$PLATFORM" = "darwin" ]; then
  artifact_platform="darwin_universal"
fi
OUTPUT_PATH="${ICUVISOR_MCPB_OUTPUT:-$ROOT_DIR/dist/icuvisor_${VERSION}_${artifact_platform}.mcpb}"

STAGE_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$STAGE_DIR"
}
trap cleanup EXIT

mkdir -p "$STAGE_DIR/server" "$STAGE_DIR/assets" "$(dirname "$OUTPUT_PATH")"
cp "$MANIFEST_TEMPLATE" "$STAGE_DIR/manifest.json"
cp "$BINARY_PATH" "$STAGE_DIR/server/icuvisor"
chmod 0755 "$STAGE_DIR/server/icuvisor" || true
cp "$ICON_SOURCE" "$STAGE_DIR/assets/icon.png"
cp "$ROOT_DIR/LICENSE" "$STAGE_DIR/LICENSE"
cp "$ROOT_DIR/CHANGELOG.md" "$STAGE_DIR/CHANGELOG.md"
cp "$BUNDLE_README" "$STAGE_DIR/README.md"

python3 - "$STAGE_DIR/manifest.json" "$VERSION" "$PLATFORM" <<'PY'
import json
import sys
from pathlib import Path

manifest_path = Path(sys.argv[1])
version = sys.argv[2]
platform = sys.argv[3]

manifest = json.loads(manifest_path.read_text())
manifest["version"] = version
manifest.setdefault("compatibility", {})["platforms"] = [platform]

binary_name = "icuvisor.exe" if platform == "win32" else "icuvisor"
server_path = f"server/{binary_name}"
manifest["server"]["entry_point"] = server_path
manifest["server"].setdefault("mcp_config", {})["command"] = "${__dirname}/" + server_path

manifest_path.write_text(json.dumps(manifest, indent=2) + "\n")
PY

if [ "$PLATFORM" = "win32" ]; then
  mv "$STAGE_DIR/server/icuvisor" "$STAGE_DIR/server/icuvisor.exe"
fi

if ! npx --yes "$CLI_PACKAGE" validate "$STAGE_DIR/manifest.json"; then
  echo "MCPB manifest validation failed; check packaging/mcpb/manifest.json and $CLI_PACKAGE availability" >&2
  exit 1
fi

if ! npx --yes "$CLI_PACKAGE" pack "$STAGE_DIR" "$OUTPUT_PATH"; then
  echo "MCPB packing failed; check the staged bundle inputs and $CLI_PACKAGE availability" >&2
  exit 1
fi

echo "Wrote $OUTPUT_PATH"
