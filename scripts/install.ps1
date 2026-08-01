# Wade installer for Windows (PowerShell)
# Usage (copy-paste into PowerShell):
#   irm https://github.com/wadefengx/wade/releases/latest/download/install.ps1 | iex
# Requires: PowerShell 5.1+ (built into Windows). No admin needed.

$ErrorActionPreference = 'Stop'

$repo = 'wadefengx/wade'
$installDir = Join-Path $env:LOCALAPPDATA 'wade'
$shimDir = Join-Path $installDir 'shims'

Write-Host "`n🏄 Wade Installer`n" -ForegroundColor Cyan

# --- Detect architecture ---
switch ($env:PROCESSOR_ARCHITECTURE) {
  'AMD64' { $arch = 'amd64' }
  'ARM64' { $arch = 'arm64' }
  default { Write-Host "❌ Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" -ForegroundColor Red; exit 1 }
}

# --- Resolve latest release ---
Write-Host "Fetching latest release..." -ForegroundColor Gray
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest" -Headers @{ 'User-Agent' = 'wade-installer' }
$version = $release.tag_name
$fileName = "wade-windows-$arch.zip"
$url = "https://github.com/$repo/releases/download/$version/$fileName"
Write-Host "Latest version: $version" -ForegroundColor Green

# --- Download + verify checksum ---
$tmp = Join-Path $env:TEMP "wade-$([guid]::NewGuid())"
New-Item -ItemType Directory -Path $tmp | Out-Null
$zipPath = Join-Path $tmp $fileName

Write-Host "Downloading $fileName ..." -ForegroundColor Gray
Invoke-WebRequest -Uri $url -OutFile $zipPath

$shaAsset = $release.assets | Where-Object { $_.name -eq "$fileName.sha256" }
if ($shaAsset) {
  $expected = (Invoke-WebRequest -Uri $shaAsset.browser_download_url -UseBasicParsing).Content.Trim() -split '\s+' | Select-Object -First 1
  $actual = (Get-FileHash -Algorithm SHA256 -Path $zipPath).Hash.ToLower()
  if ($actual -ne $expected.ToLower()) {
    Write-Host "❌ Checksum mismatch! Expected $expected, got $actual" -ForegroundColor Red
    Remove-Item -Recurse -Force $tmp
    exit 1
  }
  Write-Host "✓ Checksum verified" -ForegroundColor Green
}

# --- Extract ---
Write-Host "Extracting..." -ForegroundColor Gray
Expand-Archive -Path $zipPath -DestinationPath $tmp -Force
$exe = Get-ChildItem -Path $tmp -Filter 'wade.exe' -Recurse | Select-Object -First 1
if (-not $exe) { Write-Host "❌ wade.exe not found in archive" -ForegroundColor Red; Remove-Item -Recurse -Force $tmp; exit 1 }

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
