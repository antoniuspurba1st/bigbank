$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

param(
    [switch]$SkipUi,
    [int]$LedgerPort = 8080,
    [int]$TransactionPort = 8081,
    [int]$FraudPort = 8082,
    [int]$UiPort = 3000
)

. "$PSScriptRoot/common.ps1"

$registryPath = Get-ProcessRegistryPath
if (Test-Path $registryPath) {
    throw "A DD Bank process registry already exists at $registryPath. Run stop-all first if needed."
}

$root = Get-DDBankRoot
$powershellExe = Get-ToolPath -Name 'powershell'
$services = @(
    @{
        name = 'ledger'
        script = Join-Path $root 'scripts\run-ledger.ps1'
        args = @('-Port', $LedgerPort.ToString())
        health = "http://127.0.0.1:$LedgerPort/health"
    },
    @{
        name = 'fraud'
        script = Join-Path $root 'scripts\run-fraud.ps1'
        args = @('-Port', $FraudPort.ToString())
        health = "http://127.0.0.1:$FraudPort/health"
    },
    @{
        name = 'transaction'
        script = Join-Path $root 'scripts\run-transaction.ps1'
        args = @(
            '-Port', $TransactionPort.ToString(),
            '-FraudServiceUrl', "http://127.0.0.1:$FraudPort",
            '-LedgerServiceUrl', "http://127.0.0.1:$LedgerPort"
        )
        health = "http://127.0.0.1:$TransactionPort/health"
    }
)

if (-not $SkipUi) {
    $services += @{
        name = 'ui'
        script = Join-Path $root 'scripts\run-ui.ps1'
        args = @(
            '-Port', $UiPort.ToString(),
            '-TransactionServiceUrl', "http://127.0.0.1:$TransactionPort"
        )
        health = "http://127.0.0.1:$UiPort/transfer"
    }
}

$processes = foreach ($service in $services) {
    $argumentList = @(
        '-NoProfile',
        '-ExecutionPolicy', 'Bypass',
        '-File', $service.script
    ) + $service.args

    $process = Start-Process -FilePath $powershellExe -ArgumentList $argumentList -PassThru -WindowStyle Hidden
    [pscustomobject]@{
        name = $service.name
        pid = $process.Id
        health = $service.health
    }
}

try {
    foreach ($process in $processes) {
        Wait-Http -Url $process.health -TimeoutSeconds 120 | Out-Null
    }

    $processes | ConvertTo-Json -Depth 4 | Set-Content -Path $registryPath
    [pscustomobject]@{
        status = 'started'
        registry = $registryPath
        services = $processes
    } | ConvertTo-Json -Depth 5
} catch {
    foreach ($process in $processes) {
        Stop-Process -Id $process.pid -Force -ErrorAction SilentlyContinue
    }
    Remove-Item $registryPath -Force -ErrorAction SilentlyContinue
    throw
}
