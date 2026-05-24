#!/bin/sh
# icuvisor installer — POSIX sh.
#
# Usage:
#   curl -fsSL https://icuvisor.app/install.sh | sh
#   curl -fsSL https://icuvisor.app/install.sh -o install.sh && less install.sh && sh install.sh
#
# Environment variables:
#   INSTALL_DIR              Directory to install icuvisor into. Default: directory of
#                            an existing 'icuvisor' on PATH (so re-running updates in
#                            place), else /usr/local/bin if writable (uses sudo
#                            otherwise), else $HOME/.local/bin.
#   INSTALL_VERSION          Version to install, e.g. "v0.4.1". Default: latest release.
#   INSTALL_REQUIRE_COSIGN   If "1", abort when cosign is missing or signature check
#                            fails. Default: best-effort (verify if cosign is present,
#                            fall back to SHA-256-only with a warning otherwise).
#   INSTALL_NO_MODIFY_PATH   If "1", skip the PATH advisory at the end.
#
# Flags:
#   --dry-run                Resolve version and asset, print what would happen, exit 0.
#   --check                  Print whether an update is available and exit. Exit code
#                            is 0 when already up to date, 1 when an update is
#                            available, other on error. Does not modify anything.
#   --force                  Reinstall even if the installed version already matches.
#   --help                   Print usage.
#
# Exit codes:
#   0  success / up to date (--check)
#   1  generic failure, or update available (--check)
#   2  unsupported OS/arch, or refusing to overwrite a package-manager-managed install
#   3  download / network failure
#   4  signature or checksum verification failed

set -eu

REPO="ricardocabral/icuvisor"
RELEASES_BASE="https://github.com/${REPO}/releases"
# OIDC identity pinned to the release workflow. Update both fields together if the
# workflow file moves.
COSIGN_IDENTITY="https://github.com/${REPO}/.github/workflows/release.yml@refs/tags/"
COSIGN_ISSUER="https://token.actions.githubusercontent.com"

DRY_RUN=0
CHECK_ONLY=0
FORCE=0
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=1 ;;
    --check)   CHECK_ONLY=1 ;;
    --force)   FORCE=1 ;;
    --help|-h)
      sed -n '2,36p' "$0" | sed 's/^# \{0,1\}//'
      exit 0
      ;;
    *)
      echo "install.sh: unknown argument: $arg" >&2
      exit 1
      ;;
  esac
done

# ---------- helpers ----------

log()  { printf '==> %s\n' "$*"; }
warn() { printf 'install.sh: warning: %s\n' "$*" >&2; }
die()  { code="${2:-1}"; printf 'install.sh: error: %s\n' "$1" >&2; exit "$code"; }

need() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

http_get() {
  # http_get URL OUT
  if command -v curl >/dev/null 2>&1; then
    curl --fail --silent --show-error --location --output "$2" "$1"
  elif command -v wget >/dev/null 2>&1; then
    wget --quiet --output-document="$2" "$1"
  else
    die "neither curl nor wget is available" 3
  fi
}

# ---------- detect OS/arch ----------

uname_s=$(uname -s 2>/dev/null || echo unknown)
uname_m=$(uname -m 2>/dev/null || echo unknown)

case "$uname_s" in
  Linux)                            os=linux  ;;
  Darwin)                           os=darwin ;;
  MINGW*|MSYS*|CYGWIN*|Windows_NT)  os=windows ;;
  *) die "unsupported OS: $uname_s" 2 ;;
esac

case "$uname_m" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) die "unsupported architecture: $uname_m" 2 ;;
esac

# Asset naming follows .goreleaser.yaml. macOS ships a universal tarball; Linux and
# Windows are per-arch.
case "$os" in
  linux)   asset_suffix="linux_${arch}.tar.gz"     ; archive_kind=tar ;;
  darwin)  asset_suffix="macos_universal.tar.gz"   ; archive_kind=tar ;;
  windows) asset_suffix="windows_${arch}.zip"      ; archive_kind=zip ;;
esac

# ---------- resolve version ----------

if [ -n "${INSTALL_VERSION:-}" ]; then
  version="$INSTALL_VERSION"
  case "$version" in v*) ;; *) version="v$version" ;; esac
else
  log "resolving latest release from GitHub"
  # Follow the /latest redirect to read the tag without a JSON parser.
  if command -v curl >/dev/null 2>&1; then
    redirect=$(curl --silent --head --location -o /dev/null -w '%{url_effective}' \
      "${RELEASES_BASE}/latest") || die "could not resolve latest release" 3
  else
    redirect=$(wget --max-redirect=0 -S --spider "${RELEASES_BASE}/latest" 2>&1 \
      | awk '/^[[:space:]]*Location: / { print $2 }' | tail -n 1) \
      || die "could not resolve latest release" 3
  fi
  version=${redirect##*/}
  case "$version" in
    v[0-9]*) ;;
    *) die "could not parse latest version from '$redirect'" 3 ;;
  esac
fi

version_bare=${version#v}
asset="icuvisor_${version_bare}_${asset_suffix}"
base_url="${RELEASES_BASE}/download/${version}"

log "icuvisor ${version} for ${os}/${arch} (asset: ${asset})"

# ---------- detect existing install ----------
#
# If INSTALL_DIR is unset and an 'icuvisor' is already on PATH, prefer overwriting
# it in place so repeated `curl … | sh` invocations update rather than scatter
# binaries across PATH entries. Refuse to clobber binaries managed by Homebrew or
# Scoop — those should be updated through their own tooling.

existing_path=""
if command -v icuvisor >/dev/null 2>&1; then
  existing_path=$(command -v icuvisor)
  # Resolve symlinks so we detect the real on-disk location (Homebrew shims, …).
  if command -v readlink >/dev/null 2>&1; then
    resolved=$(readlink -f "$existing_path" 2>/dev/null || readlink "$existing_path" 2>/dev/null || echo "$existing_path")
  else
    resolved=$existing_path
  fi

  case "$resolved" in
    */Cellar/*|*/homebrew/*|*/linuxbrew/*)
      die "icuvisor is managed by Homebrew at $existing_path — run 'brew upgrade icuvisor' instead, or set INSTALL_DIR to override." 2
      ;;
    */scoop/apps/*|*/scoop/shims/*)
      die "icuvisor is managed by Scoop at $existing_path — run 'scoop update icuvisor' instead, or set INSTALL_DIR to override." 2
      ;;
  esac

  if [ -z "${INSTALL_DIR:-}" ]; then
    INSTALL_DIR=$(dirname "$resolved")
  fi
fi

# Best-effort read of the currently installed version. We do not fail if this
# doesn't parse — an unparseable older binary just means we'll reinstall.
current_version=""
if [ -n "$existing_path" ] && [ -x "$existing_path" ]; then
  current_output=$("$existing_path" version 2>/dev/null || "$existing_path" --version 2>/dev/null || true)
  current_version=$(printf '%s' "$current_output" \
    | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+([-+][A-Za-z0-9.+-]+)?' \
    | head -n 1 || true)
fi

if [ "$CHECK_ONLY" -eq 1 ]; then
  if [ -z "$existing_path" ]; then
    log "not installed; latest is ${version}"
    exit 1
  fi
  if [ "$current_version" = "$version" ]; then
    log "icuvisor ${current_version} is up to date"
    exit 0
  fi
  log "update available: ${current_version:-<unknown>} -> ${version}"
  exit 1
fi

if [ -n "$current_version" ] && [ "$current_version" = "$version" ] && [ "$FORCE" -ne 1 ]; then
  log "icuvisor ${current_version} is already installed at ${existing_path}; pass --force to reinstall"
  exit 0
fi

if [ "$DRY_RUN" -eq 1 ]; then
  if [ -n "$current_version" ]; then
    log "dry run: would update ${current_version} -> ${version}"
  fi
  log "dry run: would download ${base_url}/${asset}"
  log "dry run: would verify ${base_url}/SHA256SUMS.txt (+ .sig/.pem if present)"
  exit 0
fi

# ---------- choose install dir ----------

if [ -z "${INSTALL_DIR:-}" ]; then
  if [ -w /usr/local/bin ] 2>/dev/null; then
    INSTALL_DIR=/usr/local/bin
    sudo_cmd=""
  elif command -v sudo >/dev/null 2>&1 && [ -d /usr/local/bin ]; then
    INSTALL_DIR=/usr/local/bin
    sudo_cmd="sudo"
  else
    INSTALL_DIR="$HOME/.local/bin"
    sudo_cmd=""
  fi
else
  if [ -w "$INSTALL_DIR" ] 2>/dev/null; then
    sudo_cmd=""
  elif command -v sudo >/dev/null 2>&1; then
    sudo_cmd="sudo"
  else
    sudo_cmd=""
  fi
fi

mkdir -p "$INSTALL_DIR" 2>/dev/null || ${sudo_cmd:-true} mkdir -p "$INSTALL_DIR" \
  || die "cannot create $INSTALL_DIR" 1

# ---------- download ----------

tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t icuvisor-install)
trap 'rm -rf "$tmpdir"' EXIT INT TERM

log "downloading $asset"
http_get "${base_url}/${asset}"             "${tmpdir}/${asset}"      || die "download failed" 3
log "downloading SHA256SUMS.txt"
http_get "${base_url}/SHA256SUMS.txt"       "${tmpdir}/SHA256SUMS.txt" || die "checksum download failed" 3

sig_url="${base_url}/SHA256SUMS.txt.sig"
pem_url="${base_url}/SHA256SUMS.txt.pem"
have_sig=0
if http_get "$sig_url" "${tmpdir}/SHA256SUMS.txt.sig" 2>/dev/null \
   && http_get "$pem_url" "${tmpdir}/SHA256SUMS.txt.pem" 2>/dev/null; then
  have_sig=1
fi

# ---------- verify cosign signature ----------

if [ "$have_sig" -eq 1 ] && command -v cosign >/dev/null 2>&1; then
  log "verifying cosign signature (keyless, OIDC)"
  if ! cosign verify-blob \
        --certificate "${tmpdir}/SHA256SUMS.txt.pem" \
        --signature   "${tmpdir}/SHA256SUMS.txt.sig" \
        --certificate-identity-regexp "^${COSIGN_IDENTITY}.*$" \
        --certificate-oidc-issuer "$COSIGN_ISSUER" \
        "${tmpdir}/SHA256SUMS.txt" >/dev/null 2>&1; then
    die "cosign signature verification FAILED — refusing to install" 4
  fi
  log "cosign signature: OK"
elif [ "${INSTALL_REQUIRE_COSIGN:-0}" = "1" ]; then
  if [ "$have_sig" -eq 0 ]; then
    die "INSTALL_REQUIRE_COSIGN=1 but the release has no .sig/.pem artifacts" 4
  fi
  die "INSTALL_REQUIRE_COSIGN=1 but cosign is not installed (https://docs.sigstore.dev/cosign/installation/)" 4
else
  if [ "$have_sig" -eq 0 ]; then
    warn "no cosign signature found for this release — falling back to SHA-256 only"
  else
    warn "cosign not installed — falling back to SHA-256 only (set INSTALL_REQUIRE_COSIGN=1 to require it; see https://docs.sigstore.dev/cosign/installation/)"
  fi
fi

# ---------- verify sha256 ----------

log "verifying SHA-256"
expected=$(awk -v a="$asset" '$2==a || $2=="*"a { print $1 }' "${tmpdir}/SHA256SUMS.txt" | head -n 1)
if [ -z "$expected" ]; then
  die "$asset not listed in SHA256SUMS.txt" 4
fi

if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "${tmpdir}/${asset}" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
  actual=$(shasum -a 256 "${tmpdir}/${asset}" | awk '{print $1}')
else
  die "no sha256 tool found (need sha256sum or shasum)" 1
fi

if [ "$expected" != "$actual" ]; then
  die "SHA-256 mismatch for ${asset} (expected ${expected}, got ${actual})" 4
fi
log "SHA-256: OK"

# ---------- extract ----------

extract_dir="${tmpdir}/extract"
mkdir -p "$extract_dir"
case "$archive_kind" in
  tar)
    tar -xzf "${tmpdir}/${asset}" -C "$extract_dir" || die "tar extraction failed" 1
    ;;
  zip)
    need unzip
    unzip -q "${tmpdir}/${asset}" -d "$extract_dir" || die "unzip failed" 1
    ;;
esac

binary_name=icuvisor
case "$os" in windows) binary_name=icuvisor.exe ;; esac

binary_path=$(find "$extract_dir" -type f -name "$binary_name" -print | head -n 1)
[ -n "$binary_path" ] || die "binary $binary_name not found inside $asset" 1

# ---------- install (atomic replace) ----------
#
# Copy the new binary to "<dest>.new" inside INSTALL_DIR, then rename over the
# destination. Same-filesystem rename is atomic on POSIX, and on Linux the kernel
# keeps the old inode alive for any process currently executing the binary, so a
# running icuvisor keeps running off the old inode while the new file takes its
# place.

dest="${INSTALL_DIR}/${binary_name}"
dest_tmp="${dest}.new.$$"

if [ -n "$current_version" ] && [ "$current_version" != "$version" ]; then
  log "updating ${current_version} -> ${version} at ${dest}"
else
  log "installing ${version} to ${dest}"
fi

${sudo_cmd:-} cp "$binary_path" "$dest_tmp" \
  || die "could not write ${dest_tmp}" 1
${sudo_cmd:-} chmod 0755 "$dest_tmp" \
  || { ${sudo_cmd:-} rm -f "$dest_tmp"; die "chmod failed on ${dest_tmp}" 1; }
${sudo_cmd:-} mv -f "$dest_tmp" "$dest" \
  || { ${sudo_cmd:-} rm -f "$dest_tmp"; die "could not move ${dest_tmp} into place" 1; }

# ---------- post-install ----------

installed_version=$("$dest" version 2>/dev/null \
                    || "$dest" --version 2>/dev/null \
                    || echo "$version")
log "installed ${installed_version} at ${dest}"

case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    if [ "${INSTALL_NO_MODIFY_PATH:-0}" != "1" ]; then
      warn "$INSTALL_DIR is not on your PATH"
      printf '\n  Add this to your shell profile:\n    export PATH="%s:$PATH"\n\n' "$INSTALL_DIR"
    fi
    ;;
esac

cat <<EOF

Next steps:
  1. Run 'icuvisor setup' once to store your intervals.icu API key in the OS keychain.
  2. Point your MCP client (Claude Desktop, Cursor, …) at:
       ${dest}
  3. Docs: https://icuvisor.app

EOF
