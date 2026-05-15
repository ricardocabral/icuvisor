#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
dist_dir=${ICUVISOR_DIST_DIR:-"$repo_root/dist"}
plist_template=${ICUVISOR_PLIST_TEMPLATE:-"$repo_root/build/macos/Info.plist"}
release_mode=${ICUVISOR_MACOS_RELEASE:-0}

if [[ "$(uname -s)" != "Darwin" ]]; then
    if [[ "$release_mode" == "1" ]]; then
        echo "error: macOS DMG release packaging requires a macOS runner" >&2
        exit 1
    fi
    echo "warning: skipping macOS DMG packaging on non-macOS host" >&2
    exit 0
fi

metadata_value() {
    local key=$1
    python3 - "$dist_dir/metadata.json" "$key" <<'PY'
import json, sys
path, key = sys.argv[1], sys.argv[2]
try:
    with open(path, encoding="utf-8") as f:
        data = json.load(f)
except FileNotFoundError:
    print("")
    raise SystemExit(0)
print(data.get(key, ""))
PY
}

version=${ICUVISOR_VERSION:-$(metadata_value version)}
commit=${ICUVISOR_COMMIT:-$(metadata_value commit)}
version=${version:-dev}
commit=${commit:-local}
bundle_version=${ICUVISOR_BUNDLE_VERSION:-$commit}

binary_path=${ICUVISOR_DARWIN_BINARY:-}
if [[ -z "$binary_path" ]]; then
    while IFS= read -r candidate; do
        binary_path=$candidate
        break
    done < <(find "$dist_dir" -path '*_darwin_all/icuvisor' -type f | sort)
fi

if [[ -z "$binary_path" || ! -f "$binary_path" ]]; then
    echo "error: universal darwin binary not found under $dist_dir" >&2
    exit 1
fi

work_dir="$dist_dir/macos"
app_path="$work_dir/icuvisor.app"
stage_dir="$work_dir/dmg-root"
if [[ "$release_mode" == "1" ]]; then
    dmg_path="$dist_dir/icuvisor_${version}_macos_universal.dmg"
else
    dmg_path="$dist_dir/icuvisor_${version}_macos_universal_unsigned.dmg"
fi

rm -rf "$work_dir"
mkdir -p "$app_path/Contents/MacOS" "$stage_dir"

python3 - "$plist_template" "$app_path/Contents/Info.plist" "$version" "$bundle_version" <<'PY'
from pathlib import Path
import sys
src, dst, version, build = sys.argv[1:5]
text = Path(src).read_text(encoding="utf-8")
text = text.replace("__ICUVISOR_VERSION__", version)
text = text.replace("__ICUVISOR_BUILD__", build)
Path(dst).write_text(text, encoding="utf-8")
PY

cp "$binary_path" "$app_path/Contents/MacOS/icuvisor"
chmod 0755 "$app_path/Contents/MacOS/icuvisor"

if [[ "$release_mode" == "1" ]]; then
    : "${APPLE_TEAM_ID:?APPLE_TEAM_ID is required for release packaging}"
    : "${APPLE_API_KEY_ID:?APPLE_API_KEY_ID is required for notarization}"
    : "${APPLE_API_KEY_ISSUER:?APPLE_API_KEY_ISSUER is required for notarization}"
    : "${APPLE_API_KEY_PATH:?APPLE_API_KEY_PATH must point to the decoded App Store Connect .p8 key}"
    if [[ ! -f "$APPLE_API_KEY_PATH" ]]; then
        echo "error: APPLE_API_KEY_PATH does not point to a file" >&2
        exit 1
    fi

    identity=${APPLE_DEVELOPER_IDENTITY:-}
    if [[ -z "$identity" ]]; then
        identity=$(security find-identity -v -p codesigning | awk -F'"' '/Developer ID Application/ { print $2; exit }')
    fi
    if [[ -z "$identity" ]]; then
        echo "error: no Developer ID Application signing identity found" >&2
        security find-identity -v -p codesigning >&2 || true
        exit 1
    fi

    codesign --force --options runtime --timestamp --sign "$identity" "$app_path"
    codesign --verify --deep --strict --verbose=2 "$app_path"
else
    echo "warning: creating unsigned macOS DMG scaffold; do not publish this artifact" >&2
fi

ln -s /Applications "$stage_dir/Applications"
cp -R "$app_path" "$stage_dir/icuvisor.app"
rm -f "$dmg_path"
hdiutil create -volname "icuvisor" -srcfolder "$stage_dir" -ov -format UDZO "$dmg_path"

if [[ "$release_mode" == "1" ]]; then
    codesign --force --timestamp --sign "$identity" "$dmg_path"
    codesign --verify --verbose=2 "$dmg_path"
    xcrun notarytool submit "$dmg_path" \
        --key "$APPLE_API_KEY_PATH" \
        --key-id "$APPLE_API_KEY_ID" \
        --issuer "$APPLE_API_KEY_ISSUER" \
        --wait
    xcrun stapler staple "$dmg_path"
    xcrun stapler validate "$dmg_path"
fi

echo "$dmg_path"
