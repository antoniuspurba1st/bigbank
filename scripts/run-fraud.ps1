$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$root = Split-Path $PSScriptRoot -Parent
$fraudPath = Join-Path $root 'fraud-service'

Set-Location $fraudPath
cargo run
