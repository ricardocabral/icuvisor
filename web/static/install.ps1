<#
.SYNOPSIS
  icuvisor installer for native Windows / PowerShell.

.DESCRIPTION
  Downloads the matching Windows release archive, verifies the SHA-256 checksum
  (and, if cosign is present, the keyless Sigstore signature), and installs the
  icuvisor.exe binary into a per-user directory on PATH.

.PARAMETER InstallDir
  Directory to install icuvisor.exe into. Default: directory of an existing
  icuvisor.exe on PATH (so re-running updates in place), else
  $env:LOCALAPPDATA\Programs\icuvisor.

.PARAMETER Version
  Release tag to install (e.g. "v0.4.1"). Default: latest.

.PARAMETER RequireCosign
  If set, abort when cosign is missing or signature verification fails.

.PARAMETER DryRun
  Resolve version + asset and print what would happen without downloading.

.PARAMETER Check
  Print whether an update is available and exit. Exit code is 0 when already
  up to date, 1 when an update is available. Does not modify anything.

.PARAMETER Force
  Reinstall even if the installed version already matches the target.

.EXAMPLE
  # One-liner:
  iwr -useb https://icuvisor.app/install.ps1 | iex

.EXAMPLE
  # Inspect first, then run:
  iwr -useb https://icuvisor.app/install.ps1 -OutFile install.ps1
  notepad install.ps1
  .\install.ps1
#>

[CmdletBinding()]
param(
  [string]$InstallDir,
  [string]$Version,
  [switch]$RequireCosign,
  [switch]$DryRun,
  [switch]$Check,
  [switch]$Force
)

$ErrorActionPreference = 'Stop'
$ProgressPreference    = 'SilentlyContinue'

$Repo            = 'ricardocabral/icuvisor'
$ReleasesBase    = "https://github.com/$Repo/releases"
$CosignIdentity  = "https://github.com/$Repo/.github/workflows/release.yml@refs/tags/"
$CosignIssuer    = 'https://token.actions.githubusercontent.com'

function Write-Step([string]$msg) { Write-Host "==> $msg" }
function Write-Warn([string]$msg) { Write-Warning $msg }

function Get-Arch {
  switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { return 'amd64' }
    'ARM64' { return 'arm64' }
    'x86'   { throw 'icuvisor does not ship a 32-bit Windows binary.' }
    default { throw "unsupported PROCESSOR_ARCHITECTURE: $($env:PROCESSOR_ARCHITECTURE)" }
  }
}

function Get-LatestVersion {
  # Follow the /latest redirect; avoids JSON parsing and the GitHub API rate limit.
  $req = [System.Net.WebRequest]::Create("$ReleasesBase/latest")
  $req.AllowAutoRedirect = $false
  $req.Method = 'HEAD'
  $resp = $req.GetResponse()
  try {
    $location = $resp.Headers['Location']
  } finally {
    $resp.Close()
  }
  if (-not $location) { throw 'could not resolve latest release' }
  $tag = ($location -split '/')[-1]
  if ($tag -notmatch '^v\d') { throw "could not parse latest version from '$location'" }
  return $tag
}

$arch = Get-Arch

if (-not $Version) {
  Write-Step 'resolving latest release from GitHub'
  $Version = Get-LatestVersion
} elseif ($Version -notmatch '^v') {
  $Version = "v$Version"
}
$versionBare = $Version.TrimStart('v')

$asset    = "icuvisor_${versionBare}_windows_${arch}.zip"
$baseUrl  = "$ReleasesBase/download/$Version"
$assetUrl = "$baseUrl/$asset"

Write-Step "icuvisor $Version for windows/$arch (asset: $asset)"

# Detect existing install. If InstallDir is unset and icuvisor.exe is on PATH,
# overwrite the existing copy so re-running updates in place. Refuse to clobber
# a Scoop-managed install.
$existingCmd  = Get-Command icuvisor -ErrorAction SilentlyContinue
$existingPath = if ($existingCmd) { $existingCmd.Source } else { $null }
if ($existingPath) {
  if ($existingPath -match '\\scoop\\(apps|shims)\\') {
    throw "icuvisor is managed by Scoop at $existingPath — run 'scoop update icuvisor' instead, or pass -InstallDir to override."
  }
  if (-not $InstallDir) {
    $InstallDir = Split-Path -Parent $existingPath
  }
}

$currentVersion = $null
if ($existingPath -and (Test-Path $existingPath)) {
  try {
    $vOut = & $existingPath version 2>$null
    if (-not $vOut) { $vOut = & $existingPath --version 2>$null }
    if ($vOut) {
      $m = ([string]::Join("`n", $vOut)) | Select-String -Pattern 'v\d+\.\d+\.\d+([-+][A-Za-z0-9.+-]+)?' -AllMatches
      if ($m -and $m.Matches.Count -gt 0) { $currentVersion = $m.Matches[0].Value }
    }
  } catch { }
}

if ($Check) {
  if (-not $existingPath) {
    Write-Step "not installed; latest is $Version"
    exit 1
  }
  if ($currentVersion -eq $Version) {
    Write-Step "icuvisor $currentVersion is up to date"
    exit 0
  }
  $cur = if ($currentVersion) { $currentVersion } else { '<unknown>' }
  Write-Step "update available: $cur -> $Version"
  exit 1
}

if ($currentVersion -and $currentVersion -eq $Version -and -not $Force) {
  Write-Step "icuvisor $currentVersion is already installed at $existingPath; pass -Force to reinstall"
  return
}

if ($DryRun) {
  if ($currentVersion) { Write-Step "dry run: would update $currentVersion -> $Version" }
  Write-Step "dry run: would download $assetUrl"
  Write-Step "dry run: would verify $baseUrl/SHA256SUMS.txt (+ .sig/.pem if present)"
  return
}

if (-not $InstallDir) {
  $InstallDir = Join-Path $env:LOCALAPPDATA 'Programs\icuvisor'
}
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("icuvisor-install-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp -Force | Out-Null

try {
  $archivePath  = Join-Path $tmp $asset
  $checksumPath = Join-Path $tmp 'SHA256SUMS.txt'
  $sigPath      = Join-Path $tmp 'SHA256SUMS.txt.sig'
  $pemPath      = Join-Path $tmp 'SHA256SUMS.txt.pem'

  Write-Step "downloading $asset"
  Invoke-WebRequest -Uri $assetUrl -OutFile $archivePath -UseBasicParsing
  Write-Step 'downloading SHA256SUMS.txt'
  Invoke-WebRequest -Uri "$baseUrl/SHA256SUMS.txt" -OutFile $checksumPath -UseBasicParsing

  $haveSig = $false
  try {
    Invoke-WebRequest -Uri "$baseUrl/SHA256SUMS.txt.sig" -OutFile $sigPath -UseBasicParsing
    Invoke-WebRequest -Uri "$baseUrl/SHA256SUMS.txt.pem" -OutFile $pemPath -UseBasicParsing
    $haveSig = $true
  } catch {
    $haveSig = $false
  }

  $cosign = Get-Command cosign -ErrorAction SilentlyContinue
  if ($haveSig -and $cosign) {
    Write-Step 'verifying cosign signature (keyless, OIDC)'
    $cosignArgs = @(
      'verify-blob',
      '--certificate',                $pemPath,
      '--signature',                  $sigPath,
      '--certificate-identity-regexp', "^$([regex]::Escape($CosignIdentity)).*$",
      '--certificate-oidc-issuer',    $CosignIssuer,
      $checksumPath
    )
    & cosign @cosignArgs *> $null
    if ($LASTEXITCODE -ne 0) {
      throw 'cosign signature verification FAILED — refusing to install'
    }
    Write-Step 'cosign signature: OK'
  } elseif ($RequireCosign) {
    if (-not $haveSig) {
      throw '-RequireCosign was set but the release has no .sig/.pem artifacts'
    }
    throw '-RequireCosign was set but cosign is not installed (https://docs.sigstore.dev/cosign/installation/)'
  } else {
    if (-not $haveSig) {
      Write-Warn 'no cosign signature found for this release — falling back to SHA-256 only'
    } else {
      Write-Warn 'cosign not installed — falling back to SHA-256 only (pass -RequireCosign to require it; see https://docs.sigstore.dev/cosign/installation/)'
    }
  }

  Write-Step 'verifying SHA-256'
  $expectedLine = Get-Content $checksumPath | Where-Object {
    $parts = $_ -split '\s+', 2
    $parts.Length -eq 2 -and ($parts[1] -eq $asset -or $parts[1] -eq "*$asset")
  } | Select-Object -First 1
  if (-not $expectedLine) { throw "$asset not listed in SHA256SUMS.txt" }
  $expected = ($expectedLine -split '\s+', 2)[0].ToLowerInvariant()
  $actual   = (Get-FileHash -Algorithm SHA256 -Path $archivePath).Hash.ToLowerInvariant()
  if ($expected -ne $actual) {
    throw "SHA-256 mismatch for $asset (expected $expected, got $actual)"
  }
  Write-Step 'SHA-256: OK'

  $extractDir = Join-Path $tmp 'extract'
  Expand-Archive -Path $archivePath -DestinationPath $extractDir -Force

  $binary = Get-ChildItem -Path $extractDir -Recurse -Filter 'icuvisor.exe' | Select-Object -First 1
  if (-not $binary) { throw 'icuvisor.exe not found inside archive' }

  # Atomic-ish replace for a possibly-running .exe: rename the old file to .old
  # (allowed on Windows even while it's executing), then move the new file into
  # place, then best-effort delete the .old. If deletion fails, leave it — the
  # next install will retry the cleanup.
  $destPath = Join-Path $InstallDir 'icuvisor.exe'
  $oldPath  = "$destPath.old"
  if (Test-Path $oldPath) {
    Remove-Item -Force $oldPath -ErrorAction SilentlyContinue
  }
  if (Test-Path $destPath) {
    if ($currentVersion -and $currentVersion -ne $Version) {
      Write-Step "updating $currentVersion -> $Version at $destPath"
    } else {
      Write-Step "reinstalling $Version at $destPath"
    }
    Rename-Item -Path $destPath -NewName 'icuvisor.exe.old' -Force
  } else {
    Write-Step "installing $Version to $destPath"
  }
  Copy-Item -Path $binary.FullName -Destination $destPath -Force
  if (Test-Path $oldPath) {
    Remove-Item -Force $oldPath -ErrorAction SilentlyContinue
  }
  Write-Step "installed: $destPath"

  # Add InstallDir to the user PATH if missing.
  $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
  $pathEntries = if ($userPath) { $userPath -split ';' } else { @() }
  if ($pathEntries -notcontains $InstallDir) {
    $newPath = if ($userPath) { "$userPath;$InstallDir" } else { $InstallDir }
    [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
    Write-Warn "$InstallDir was added to your user PATH. Open a new shell to pick it up."
  }

  Write-Host ''
  Write-Host 'Next steps:'
  Write-Host "  1. Run 'icuvisor setup' once to store your intervals.icu API key in Windows Credential Manager."
  Write-Host '  2. Point your MCP client (Claude Desktop, Cursor, …) at the icuvisor.exe binary.'
  Write-Host '  3. Docs: https://icuvisor.app'
  Write-Host ''
}
finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
