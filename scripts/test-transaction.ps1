$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$root = Split-Path $PSScriptRoot -Parent
$transactionPath = Join-Path $root 'transaction-service'

Set-Location $transactionPath
$env:GOCACHE = Join-Path (Get-Location) '.gocache'
go test ./...
