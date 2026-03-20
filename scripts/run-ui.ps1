$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

param(
    [int]$Port = 3000,
    [string]$TransactionServiceUrl = ''
)

$root = Split-Path $PSScriptRoot -Parent
. "$PSScriptRoot/common.ps1"

$uiPath = Join-Path $root 'ui'
Import-EnvFile -Path (Join-Path $root '.env')
Import-EnvFile -Path (Join-Path $uiPath '.env.local')

Set-Location $uiPath
if ($TransactionServiceUrl -ne '') {
    $env:TRANSACTION_SERVICE_URL = $TransactionServiceUrl
} elseif (-not $env:TRANSACTION_SERVICE_URL) {
    $env:TRANSACTION_SERVICE_URL = 'http://127.0.0.1:8081'
}

npm run dev -- --hostname 127.0.0.1 --port $Port
