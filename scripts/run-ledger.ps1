$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$root = Split-Path $PSScriptRoot -Parent
$ledgerPath = Join-Path $root 'ls_springboot'

Set-Location $ledgerPath
$env:GRADLE_USER_HOME = Join-Path (Get-Location) '.gradle'
& .\gradlew.bat bootRun --no-daemon --console=plain
