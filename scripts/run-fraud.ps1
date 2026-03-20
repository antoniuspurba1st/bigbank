param(
    [int]$Port = 8082
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$root = Split-Path $PSScriptRoot -Parent
. "$PSScriptRoot/common.ps1"

$fraudPath = Join-Path $root 'fraud-service'
Import-EnvFile -Path (Join-Path $root '.env')
Import-EnvFile -Path (Join-Path $fraudPath '.env')

Set-Location $fraudPath
$env:PORT = $Port.ToString()
cargo run
