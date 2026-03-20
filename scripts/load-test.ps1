param(
    [int]$Requests = 40,
    [int]$Concurrency = 8,
    [decimal]$Amount = 12.34,
    [switch]$KeepRunning,
    [int]$LedgerPort = 18080,
    [int]$TransactionPort = 18081,
    [int]$FraudPort = 18082
)

. "$PSScriptRoot/common.ps1"

function Get-Percentile {
    param(
        [int[]]$Values,
        [double]$Percentile
    )

    if ($Values.Count -eq 0) {
        return 0
    }

    $sorted = $Values | Sort-Object
    $index = [math]::Ceiling(($Percentile / 100) * $sorted.Count) - 1
    if ($index -lt 0) {
        $index = 0
    }
    if ($index -ge $sorted.Count) {
        $index = $sorted.Count - 1
    }

    return $sorted[$index]
}

$stack = Start-DDBankStack -LedgerPort $LedgerPort -TransactionPort $TransactionPort -FraudPort $FraudPort

try {
    Wait-DDBankReady -Stack $stack | Out-Null

    $prefix = 'ref-load-' + (Get-Date -Format 'yyyyMMddHHmmss')
    $overall = [System.Diagnostics.Stopwatch]::StartNew()
    $jobs = @()

    $worker = {
        param($port, $reference, $amount)

        $payload = @{
            reference    = $reference
            from_account = 'ACC-001'
            to_account   = 'ACC-002'
            amount       = $amount
        } | ConvertTo-Json -Compress

        $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
        try {
            $response = Invoke-WebRequest -Uri "http://127.0.0.1:$port/transfer" -Method Post -UseBasicParsing -TimeoutSec 15 -ContentType 'application/json' -Body $payload
            $parsed = $response.Content | ConvertFrom-Json
            [pscustomobject]@{
                reference = $reference
                ok = $true
                status = $parsed.status
                duplicate = [bool]$parsed.data.duplicate
                duration_ms = $stopwatch.ElapsedMilliseconds
            }
        } catch {
            [pscustomobject]@{
                reference = $reference
                ok = $false
                status = 'error'
                duplicate = $false
                duration_ms = $stopwatch.ElapsedMilliseconds
                error = $_.Exception.Message
            }
        }
    }

    foreach ($index in 1..$Requests) {
        while ((@($jobs | Where-Object { $_.State -eq 'Running' }).Count) -ge $Concurrency) {
            Start-Sleep -Milliseconds 200
        }

        $jobs += Start-Job -ScriptBlock $worker -ArgumentList $stack.TransactionPort, "$prefix-$index", $Amount
    }

    Wait-Job -Job $jobs | Out-Null
    $results = $jobs | Receive-Job
    $jobs | Remove-Job -Force | Out-Null
    $overall.Stop()

    $successes = @($results | Where-Object { $_.ok -and $_.status -eq 'success' })
    $failures = @($results | Where-Object { -not $_.ok -or $_.status -ne 'success' })
    $durations = @($successes | ForEach-Object { [int]$_.duration_ms })

    $env:PGPASSWORD = '123123'
    $dbCount = [int](psql -h localhost -U postgres -d ddbank -At -c "select count(*) from ledger_transactions where reference like '$prefix-%';")

    if ($failures.Count -gt 0) {
        throw "Load test encountered $($failures.Count) failed requests"
    }
    if ($dbCount -ne $Requests) {
        throw "Expected $Requests persisted transactions from load test, got $dbCount"
    }

    [pscustomobject]@{
        requests = $Requests
        concurrency = $Concurrency
        total_duration_ms = $overall.ElapsedMilliseconds
        throughput_rps = [math]::Round(($Requests / [math]::Max(($overall.ElapsedMilliseconds / 1000.0), 0.001)), 2)
        success_count = $successes.Count
        persisted_transactions = $dbCount
        min_ms = ($durations | Measure-Object -Minimum).Minimum
        avg_ms = [math]::Round(($durations | Measure-Object -Average).Average, 2)
        max_ms = ($durations | Measure-Object -Maximum).Maximum
        p95_ms = Get-Percentile -Values $durations -Percentile 95
        p99_ms = Get-Percentile -Values $durations -Percentile 99
    } | ConvertTo-Json -Depth 6
} catch {
    Write-DDBankDiagnostics -Stack $stack
    throw
} finally {
    if (-not $KeepRunning) {
        Stop-DDBankStack -Stack $stack
    }
}
