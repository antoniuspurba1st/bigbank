$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

. "$PSScriptRoot/common.ps1"

$registryPath = Get-ProcessRegistryPath
if (-not (Test-Path $registryPath)) {
    [pscustomobject]@{
        status = 'no_process_registry'
        registry = $registryPath
    } | ConvertTo-Json -Depth 4
    exit 0
}

$processes = Get-Content $registryPath | ConvertFrom-Json
foreach ($process in $processes) {
    Stop-Process -Id $process.pid -Force -ErrorAction SilentlyContinue
}

Remove-Item $registryPath -Force -ErrorAction SilentlyContinue
[pscustomobject]@{
    status = 'stopped'
    services = $processes
} | ConvertTo-Json -Depth 4
