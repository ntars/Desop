# Desop portable uninstall helper. No admin rights, no system files.
param(
  [string]$Action = 'uninstall',
  [string]$InstallDir = "$env:LOCALAPPDATA\Desop",
  [switch]$KeepConfig,
  [switch]$RemoveConfig,
  [switch]$NonInteractive
)

$ErrorActionPreference = 'Stop'
$runKey = 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Run'
if (-not (Test-Path $runKey)) { New-Item -Path $runKey -Force | Out-Null }

Remove-ItemProperty -Path $runKey -Name 'Desop' -ErrorAction SilentlyContinue
$procs = Get-Process -Name desop -ErrorAction SilentlyContinue
if ($procs) { $procs | Stop-Process -Force }
$keep = if ($RemoveConfig) { 'n' } elseif ($KeepConfig) { 'y' } elseif ($NonInteractive) { 'n' } else { Read-Host 'Keep configuration at %APPDATA%\Desop? (y/N)' }
if ($keep -notmatch '^[yY]') {
  Remove-Item -Path (Join-Path $env:APPDATA 'Desop') -Recurse -Force -ErrorAction SilentlyContinue
}
Remove-Item -Path $InstallDir -Recurse -Force -ErrorAction SilentlyContinue
Write-Host 'Desop removed. HKCU Run entry cleared; no system files were touched.'
