# Wade installer for Windows (PowerShell)
# Usage (copy-paste into PowerShell):
#   irm https://github.com/wadefengx/wade/releases/latest/download/install.ps1 | iex
# Requires: PowerShell 5.1+ (built into Windows). No admin needed.

# NOTE: never use `exit` in this script — under `irm | iex` it kills the
# whole PowerShell session (window closes, error invisible). We throw and
# let the outer catch pause with Read-Host instead.

$ErrorActionPreference = 'Stop'

function Fail([string]$msg) {
  Write-Host "❌ $msg" -ForegroundColor Red
  throw $msg
}

try {
  $repo = 'wadefengx/wade'
  $installDir = Join-Path $env:LOCALAPPDATA 'wade'
  $shimDir = Join-Path $installDir 'shims'

  Write-Host "`n🏄 Wade Installer`n" -ForegroundColor Cyan

  # --- Detect architecture ---
  switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { $arch = 'amd64' }
    'ARM64' { $arch = 'arm64' }
    default { Fail "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
  }

  # --- Resolve latest release (dual-channel: API for CN, redirect fallback) ---
  Write-Host "Fetching latest release..." -ForegroundColor Gray
  $version = $null

  # Channel 1: GitHub API (reachable from CN networks; rate-limited but usually works)
  try {
    $rel = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest" -Headers @{ 'User-Agent' = 'wade-installer' } -TimeoutSec 10 -ErrorAction Stop
    if ($rel.tag_name) { $version = $rel.tag_name }
  } catch {
    Write-Host "  (api channel failed: $($_.Exception.Message))" -ForegroundColor DarkGray
  }

  # Channel 2: HTTP redirect https://github.com/$repo/releases/latest → /releases/tag/vX.Y.Z
  if (-not $version) {
    try {
      $resp = Invoke-WebRequest -Uri "https://github.com/$repo/releases/latest" -MaximumRedirection 0 -UseBasicParsing -TimeoutSec 10 -ErrorAction Stop
    } catch {
      $resp = $_.Exception.Response
    }
    if ($resp -and $resp.Headers) {
      $loc = $resp.Headers['Location']
      if ($loc -match '/tag/([^/]+)') { $version = $Matches[1] }
    }
  }

  if (-not $version) {
    Fail "Could not determine latest version. Check network/proxy, or download manually from https://github.com/$repo/releases/latest"
  }

  $fileName = "wade-windows-$arch.zip"
  $shaName = "wade-windows-$arch.sha256"
  # Prefer API asset endpoint (api.github.com + release-assets CDN reachable from CN),
  # fall back to github.com direct when the API channel failed.
  $apiAsset = $null
  if ($rel -and $rel.assets) { $apiAsset = $rel.assets | Where-Object { $_.name -eq $fileName } | Select-Object -First 1 }
  $apiSha = $null
  if ($rel -and $rel.assets) { $apiSha = $rel.assets | Where-Object { $_.name -eq $shaName } | Select-Object -First 1 }
  Write-Host "Latest version: $version" -ForegroundColor Green

  # --- Download + verify checksum ---
  $tmp = Join-Path $env:TEMP "wade-$([guid]::NewGuid())"
  New-Item -ItemType Directory -Path $tmp | Out-Null
  $zipPath = Join-Path $tmp $fileName

  Write-Host "Downloading $fileName ..." -ForegroundColor Gray
  $downloaded = $false
  $lastErr = $null

  # Attempt 1-2: API asset endpoint (api.github.com + release-assets CDN reachable from CN)
  if ($apiAsset) {
    for ($attempt = 1; $attempt -le 2; $attempt++) {
      try {
        Invoke-WebRequest -Uri $apiAsset.url -Headers @{ 'Accept' = 'application/octet-stream'; 'User-Agent' = 'wade-installer' } -OutFile $zipPath -TimeoutSec 60 -UseBasicParsing -ErrorAction Stop
        $downloaded = $true
        break
      } catch {
        $lastErr = $_.Exception.Message
        Write-Host "  (api attempt $attempt/2 failed: $lastErr)" -ForegroundColor DarkGray
        Start-Sleep -Seconds 2
      }
    }
  }

  # Attempt 3: github.com direct (works outside CN / when API is down)
  if (-not $downloaded) {
    try {
      Invoke-WebRequest -Uri "https://github.com/$repo/releases/download/$version/$fileName" -OutFile $zipPath -TimeoutSec 60 -UseBasicParsing -ErrorAction Stop
      $downloaded = $true
    } catch {
      $lastErr = $_.Exception.Message
    }
  }

  if (-not $downloaded) {
    Fail "Download failed after 3 attempts (last: $lastErr).`n       If behind a proxy, run:`n       `$env:HTTP_PROXY='http://127.0.0.1:7890'; `$env:HTTPS_PROXY='http://127.0.0.1:7890';`n       then re-run the installer."
  }

  # Checksum: API asset first, github.com fallback
  try {
    if ($apiSha) {
      $expected = (Invoke-WebRequest -Uri $apiSha.url -Headers @{ 'Accept' = 'application/octet-stream'; 'User-Agent' = 'wade-installer' } -UseBasicParsing -TimeoutSec 30 -ErrorAction Stop).Content.Trim() -split '\s+' | Select-Object -First 1
    } else {
      $expected = (Invoke-WebRequest -Uri "https://github.com/$repo/releases/download/$version/$shaName" -UseBasicParsing -TimeoutSec 30 -ErrorAction Stop).Content.Trim() -split '\s+' | Select-Object -First 1
    }
    $actual = (Get-FileHash -Algorithm SHA256 -Path $zipPath).Hash.ToLower()
    if ($actual -ne $expected.ToLower()) {
      Remove-Item -Recurse -Force $tmp
      Fail "Checksum mismatch! Expected $expected, got $actual"
    }
    Write-Host "✓ Checksum verified" -ForegroundColor Green
  } catch {
    Write-Host "⚠ Checksum unavailable ($($_.Exception.Message)) — continuing without verification" -ForegroundColor Yellow
  }

  # --- Extract ---
  Write-Host "Extracting..." -ForegroundColor Gray
  Expand-Archive -Path $zipPath -DestinationPath $tmp -Force
  $exe = Get-ChildItem -Path $tmp -Filter 'wade.exe' -Recurse | Select-Object -First 1
  if (-not $exe) { Remove-Item -Recurse -Force $tmp; Fail "wade.exe not found in archive" }

  # --- Install ---
  New-Item -ItemType Directory -Path $installDir -Force | Out-Null
  Copy-Item $exe.FullName -Destination (Join-Path $installDir 'wade.exe') -Force
  Remove-Item -Recurse -Force $tmp

  # --- Add to user PATH (permanent) ---
  $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
  if ($userPath -notlike "*$installDir*") {
    $newPath = if ([string]::IsNullOrEmpty($userPath)) { $installDir } else { "$userPath;$installDir" }
    [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
    Write-Host "✓ Added $installDir to user PATH" -ForegroundColor Green
  } else {
    Write-Host "✓ $installDir already in PATH" -ForegroundColor Green
  }

  # --- Post-install hint ---
  Write-Host "`n✅ wade $version installed to $installDir\wade.exe" -ForegroundColor Green
  Write-Host ""
  Write-Host "⚠️  PATH updated — open a NEW terminal, then run:" -ForegroundColor Yellow
  Write-Host "   wade -i" -ForegroundColor White
  Write-Host ""

  # Try to make it work in the current session too
  $env:Path = "$installDir;$env:Path"
  try { & (Join-Path $installDir 'wade.exe') version } catch { }

} catch {
  # Keep the window open so the user can read the error (irm|iex + exit = flash-close)
  Write-Host ""
  Write-Host "⚠️  Installer failed. Press Enter to close this window..." -ForegroundColor Yellow
  Read-Host | Out-Null
}
