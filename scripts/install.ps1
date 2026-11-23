# Desop portable install helper. No admin rights, no system files.
param(
  [Parameter(Mandatory = $true)][ValidateSet('install', 'uninstall', 'autostart-on', 'autostart-off')][string]$Action,
  [string]$InstallDir = "$env:LOCALAPPDATA\Desop",
  [switch]$KeepConfig,
  [switch]$RemoveConfig
)

$ErrorActionPreference = 'Stop'
$runKey = 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Run'
if (-not (Test-Path $runKey)) { New-Item -Path $runKey -Force | Out-Null }

function Get-DesopExe {
  $local = Join-Path $PSScriptRoot 'desop.exe'
  if (Test-Path $local) { return $local }
  $installed = Join-Path $InstallDir 'desop.exe'
  if (Test-Path $installed) { return $installed }
  throw 'desop.exe not found next to this script or in the install directory.'
}

switch ($Action) {
  'install' {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Copy-Item -Path (Join-Path $PSScriptRoot '*') -Destination $InstallDir -Recurse -Force
    Write-Host "Desop installed to $InstallDir (portable, user mode)."
    Write-Host "Run: `"$InstallDir\desop.exe`""
  }
  'uninstall' {
    Remove-ItemProperty -Path $runKey -Name 'Desop' -ErrorAction SilentlyContinue
    $procs = Get-Process -Name desop -ErrorAction SilentlyContinue
    if ($procs) { $procs | Stop-Process -Force }
    $keep = if ($KeepConfig) { 'y' } elseif ($RemoveConfig) { 'n' } else { Read-Host 'Keep configuration at %APPDATA%\Desop? (y/N)' }
    if ($keep -notmatch '^[yY]') {
      Remove-Item -Path (Join-Path $env:APPDATA 'Desop') -Recurse -Force -ErrorAction SilentlyContinue
    }
    Remove-Item -Path $InstallDir -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host 'Desop removed. HKCU Run entry cleared; no system files were touched.'
  }
  'autostart-on' {
    # Prefer the installed copy so the entry remains valid if the source
    # bundle is moved or deleted after installation.
    $installed = Join-Path $InstallDir 'desop.exe'
    if (-not (Test-Path $installed)) { throw "Not installed: $InstallDir" }
    New-ItemProperty -Path $runKey -Name 'Desop' -Value "`"$installed`"" -PropertyType String -Force | Out-Null
    Write-Host 'Desop autostart enabled (HKCU only).'
  }
  'autostart-off' {
    Remove-ItemProperty -Path $runKey -Name 'Desop' -ErrorAction SilentlyContinue
    Write-Host 'Desop autostart disabled.'
  }
}

