param(
    [int]$Port = 8081,
    [string]$FraudServiceUrl = '',
    [string]$LedgerServiceUrl = ''
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$root = Split-Path $PSScriptRoot -Parent
. "$PSScriptRoot/common.ps1"

$transactionPath = Join-Path $root 'transaction-service'
Import-EnvFile -Path (Join-Path $root '.env')
Import-EnvFile -Path (Join-Path $transactionPath '.env')

Set-Location $transactionPath
$env:GOCACHE = Join-Path (Get-Location) '.gocache'
$env:PORT = $Port.ToString()
if ($FraudServiceUrl -ne '') {
    $env:FRAUD_SERVICE_URL = $FraudServiceUrl
}
if ($LedgerServiceUrl -ne '') {
    $env:LEDGER_SERVICE_URL = $LedgerServiceUrl
}
go run ./cmd/main.go
