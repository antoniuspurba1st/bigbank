$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$root = Split-Path $PSScriptRoot -Parent
. "$PSScriptRoot/common.ps1"

$uiPath = Join-Path $root 'ui'
Import-EnvFile -Path (Join-Path $root '.env')
Import-EnvFile -Path (Join-Path $uiPath '.env.local')

Set-Location $uiPath
if (-not $env:TRANSACTION_SERVICE_URL) {
    $env:TRANSACTION_SERVICE_URL = 'http://127.0.0.1:8081'
}

npm run lint
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

npm run build
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
