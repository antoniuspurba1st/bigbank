$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Get-DDBankRoot {
    return Split-Path $PSScriptRoot -Parent
}

function Get-ToolPath {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name
    )

    return (Get-Command $Name -ErrorAction Stop).Source
}

function Import-EnvFile {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Path
    )

    if (-not (Test-Path $Path)) {
        return
    }

    foreach ($line in Get-Content $Path) {
        $trimmed = $line.Trim()
        if ($trimmed -eq '' -or $trimmed.StartsWith('#')) {
            continue
        }

        $parts = $trimmed -split '=', 2
        if ($parts.Count -ne 2) {
            continue
        }

        $name = $parts[0].Trim()
        $value = $parts[1].Trim().Trim('"').Trim("'")
        Set-Item -Path "Env:$name" -Value $value
    }
}

function Get-ProcessRegistryPath {
    $runtimeDir = Join-Path (Get-DDBankRoot) '.runtime'
    if (-not (Test-Path $runtimeDir)) {
        New-Item -ItemType Directory -Path $runtimeDir | Out-Null
    }

    return Join-Path $runtimeDir 'ddbank-processes.json'
}

function Wait-Http {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Url,
        [int]$TimeoutSeconds = 90
    )

    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
        try {
            return Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 5
        } catch {
            Start-Sleep -Seconds 2
        }
    }

    throw "Timeout waiting for $Url"
}

function Get-DDBankDiagnostics {
    param(
        [Parameter(Mandatory = $true)]
        [pscustomobject]$Stack
    )

    $jobs = @(
        @{ Name = 'ledger'; Job = $Stack.LedgerJob },
        @{ Name = 'fraud'; Job = $Stack.FraudJob },
        @{ Name = 'transaction'; Job = $Stack.GoJob }
    )

    $diagnostics = foreach ($entry in $jobs) {
        $output = @()
        if ($null -ne $entry.Job) {
            $output = @(Receive-Job -Job $entry.Job -Keep -ErrorAction SilentlyContinue | ForEach-Object { $_.ToString() })
        }

        [pscustomobject]@{
            service = $entry.Name
            state = if ($null -ne $entry.Job) { $entry.Job.State } else { 'unknown' }
            has_more_data = if ($null -ne $entry.Job) { $entry.Job.HasMoreData } else { $false }
            output = $output
        }
    }

    return $diagnostics
}

function Write-DDBankDiagnostics {
    param(
        [Parameter(Mandatory = $true)]
        [pscustomobject]$Stack
    )

    (Get-DDBankDiagnostics -Stack $Stack | ConvertTo-Json -Depth 6) | Write-Host
}

function Start-DDBankStack {
    param(
        [int]$LedgerPort = 18080,
        [int]$TransactionPort = 18081,
        [int]$FraudPort = 18082
    )

    $root = Get-DDBankRoot
    $ledgerPath = Join-Path $root 'ls_springboot'
    $transactionPath = Join-Path $root 'transaction-service'
    $fraudPath = Join-Path $root 'fraud-service'
    $scriptsPath = Join-Path $root 'scripts'
    $powershellExe = Get-ToolPath -Name 'powershell'

    $ledgerJob = Start-Job -ScriptBlock {
        param($exe, $scriptPath, $port)
        & $exe -NoProfile -ExecutionPolicy Bypass -File $scriptPath -Port $port
    } -ArgumentList $powershellExe, (Join-Path $scriptsPath 'run-ledger.ps1'), $LedgerPort

    $fraudJob = Start-Job -ScriptBlock {
        param($exe, $scriptPath, $port)
        & $exe -NoProfile -ExecutionPolicy Bypass -File $scriptPath -Port $port
    } -ArgumentList $powershellExe, (Join-Path $scriptsPath 'run-fraud.ps1'), $FraudPort

    $goJob = Start-Job -ScriptBlock {
        param($exe, $scriptPath, $port, $fraudUrl, $ledgerUrl)
        & $exe -NoProfile -ExecutionPolicy Bypass -File $scriptPath -Port $port -FraudServiceUrl $fraudUrl -LedgerServiceUrl $ledgerUrl
    } -ArgumentList $powershellExe, (Join-Path $scriptsPath 'run-transaction.ps1'), $TransactionPort, "http://127.0.0.1:$FraudPort", "http://127.0.0.1:$LedgerPort"

    return [pscustomobject]@{
        Root            = $root
        LedgerPort      = $LedgerPort
        TransactionPort = $TransactionPort
        FraudPort       = $FraudPort
        LedgerJob       = $ledgerJob
        FraudJob        = $fraudJob
        GoJob           = $goJob
    }
}

function Start-DDBankUiJob {
    param(
        [int]$UiPort = 3000,
        [int]$TransactionPort = 18081
    )

    $root = Get-DDBankRoot
    $powershellExe = Get-ToolPath -Name 'powershell'
    $scriptPath = Join-Path $root 'scripts\run-ui.ps1'

    return Start-Job -ScriptBlock {
        param($exe, $filePath, $port, $transactionServiceUrl)
        & $exe -NoProfile -ExecutionPolicy Bypass -File $filePath -Port $port -TransactionServiceUrl $transactionServiceUrl
    } -ArgumentList $powershellExe, $scriptPath, $UiPort, "http://127.0.0.1:$TransactionPort"
}

function Wait-DDBankReady {
    param(
        [Parameter(Mandatory = $true)]
        [pscustomobject]$Stack
    )

    return [pscustomobject]@{
        Ledger      = Wait-Http -Url "http://127.0.0.1:$($Stack.LedgerPort)/health"
        Fraud       = Wait-Http -Url "http://127.0.0.1:$($Stack.FraudPort)/health"
        Transaction = Wait-Http -Url "http://127.0.0.1:$($Stack.TransactionPort)/health"
    }
}

function Stop-DDBankStack {
    param(
        [Parameter(Mandatory = $true)]
        [pscustomobject]$Stack
    )

    Stop-Job -Job $Stack.LedgerJob, $Stack.FraudJob, $Stack.GoJob -ErrorAction SilentlyContinue | Out-Null
    Remove-Job -Job $Stack.LedgerJob, $Stack.FraudJob, $Stack.GoJob -Force -ErrorAction SilentlyContinue | Out-Null
}

function Invoke-DDBankTransfer {
    param(
        [Parameter(Mandatory = $true)]
        [int]$TransactionPort,
        [Parameter(Mandatory = $true)]
        [string]$Reference,
        [Parameter(Mandatory = $true)]
        [decimal]$Amount,
        [string]$CorrelationId = ''
    )

    $body = @{
        reference    = $Reference
        from_account = 'ACC-001'
        to_account   = 'ACC-002'
        amount       = $Amount
    } | ConvertTo-Json -Compress

    $headers = @{}
    if ($CorrelationId -ne '') {
        $headers['X-Correlation-Id'] = $CorrelationId
    }

    return Invoke-WebRequest `
        -Uri "http://127.0.0.1:$TransactionPort/transfer" `
        -Method Post `
        -UseBasicParsing `
        -ContentType 'application/json' `
        -Headers $headers `
        -Body $body
}
