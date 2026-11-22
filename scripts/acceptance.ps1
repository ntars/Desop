# Desop Windows acceptance helper. Run on a Windows 10/11 machine.
param(
  [Parameter(Mandatory = $true)][string]$Exe,
  [switch]$SkipProcessCheck,
  [switch]$SkipLocalAPI,
  [switch]$GenerateConfig
)

$ErrorActionPreference = 'Stop'
function Record([string]$name, [bool]$ok, [string]$detail = '') {
  $script:results += [pscustomobject]@{ Check = $name; Result = if ($ok) { 'PASS' } else { 'FAIL' }; Detail = $detail }
}

if (-not (Test-Path $Exe)) { throw "Executable not found: $Exe" }
$exePath = (Resolve-Path $Exe).Path
$config = Join-Path $env:APPDATA 'Desop\config.json'
$results = @()
$originalConfig = $null
if (Test-Path $config) { $originalConfig = Get-Content $config -Raw }

if ($GenerateConfig -and -not (Test-Path $config)) {
  & $exePath status 2>&1 | Out-Null
}

if (Test-Path $config) {
  Record 'config pre-existing' $true $config
} else {
  Record 'config generated on first run' $false 'Run desop.exe once, or rerun with -GenerateConfig.'
}

try {
  $json = Get-Content $config -Raw -ErrorAction Stop | ConvertFrom-Json
  Record 'config JSON valid' $true "version=$($json.version)"
} catch {
  Record 'config JSON valid' $false $_.Exception.Message
}

$status = & $exePath status 2>&1
Record 'status command' ($LASTEXITCODE -eq 0 -and "$status" -match 'config:') ("$status" -replace "`r?`n", ' | ')

$enabledBefore = & $exePath status 2>&1
& $exePath disable island 2>&1 | Out-Null
$disabled = & $exePath status 2>&1
& $exePath enable island 2>&1 | Out-Null
$reenabled = & $exePath status 2>&1
Record 'component disable/enable persists' ("$disabled" -ne "$enabledBefore" -and "$reenabled" -eq "$enabledBefore") ''

if (-not $SkipProcessCheck) {
  $p = Start-Process -FilePath $exePath -PassThru
  Start-Sleep -Seconds 3
  $running = Get-Process -Id $p.Id -ErrorAction SilentlyContinue
  Record 'process starts and stays alive' ($null -ne $running) "pid=$($p.Id)"
  Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue
  Start-Sleep -Milliseconds 500
  Record 'process stops cleanly' ($null -eq (Get-Process -Id $p.Id -ErrorAction SilentlyContinue)) ''
}

if (-not $SkipLocalAPI) {
  $health = $null
  try {
    $p = Start-Process -FilePath $exePath -PassThru
    Start-Sleep -Seconds 3
    $health = Invoke-RestMethod -Method Get -Uri 'http://127.0.0.1:17890/health' -TimeoutSec 5
    Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue
  } catch {
    Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue
  }
  Record 'island loopback health API' ($null -ne $health -and $health.ok -eq $true) ''
}

$selfTest = & $exePath selftest 2>&1
Record 'idle CPU and working-set self-test' ($LASTEXITCODE -eq 0 -and "$selfTest" -match 'cpu: PASS' -and "$selfTest" -match 'memory: PASS') ("$selfTest" -replace "`r?`n", ' | ')

$results | Format-Table -AutoSize
$originalConfig | Set-Content -Path $config -NoNewline -ErrorAction SilentlyContinue
$failed = @($results | Where-Object Result -eq 'FAIL')
Write-Host ("Summary: {0} passed, {1} failed" -f ($results.Count - $failed.Count), $failed.Count)
if ($failed.Count -gt 0) { exit 1 }
