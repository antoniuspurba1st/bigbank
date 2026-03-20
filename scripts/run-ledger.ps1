param(
    [int]$Port = 8080
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$root = Split-Path $PSScriptRoot -Parent
. "$PSScriptRoot/common.ps1"

$ledgerPath = Join-Path $root 'ls_springboot'
Import-EnvFile -Path (Join-Path $root '.env')
Import-EnvFile -Path (Join-Path $ledgerPath '.env')

Set-Location $ledgerPath
$env:GRADLE_USER_HOME = Join-Path (Get-Location) '.gradle'
$env:SERVER_PORT = $Port.ToString()
& .\gradlew.bat bootRun --no-daemon --console=plain
