param(
    [int]$Requests = 10,
    [decimal]$Amount = 77.77,
    [switch]$KeepRunning,
    [int]$LedgerPort = 18080,
    [int]$TransactionPort = 18081,
    [int]$FraudPort = 18082
)

. "$PSScriptRoot/common.ps1"

$stack = Start-DDBankStack -LedgerPort $LedgerPort -TransactionPort $TransactionPort -FraudPort $FraudPort

try {
    Wait-DDBankReady -Stack $stack | Out-Null

    $reference = 'ref-concurrency-' + (Get-Date -Format 'yyyyMMddHHmmss')
    $jobs = @()
    $worker = {
        param($port, $requestReference, $amount, $index)

        $payload = @{
            reference = $requestReference
            from_account = 'ACC-001'
            to_account = 'ACC-002'
            amount = $amount
        } | ConvertTo-Json -Compress

        try {
            $response = Invoke-WebRequest `
                -Uri "http://127.0.0.1:$port/transfer" `
                -Method Post `
                -UseBasicParsing `
                -TimeoutSec 20 `
                -ContentType 'application/json' `
                -Headers @{ 'X-Correlation-Id' = "corr-concurrency-$index" } `
                -Body $payload

            $parsed = $response.Content | ConvertFrom-Json
            [pscustomobject]@{
                ok = $true
                status = $parsed.status
                duplicate = [bool]$parsed.data.duplicate
                transaction_id = $parsed.data.transaction_id
            }
        } catch {
            [pscustomobject]@{
                ok = $false
                status = 'error'
                duplicate = $false
                error = $_.Exception.Message
            }
        }
    }

    foreach ($index in 1..$Requests) {
        $jobs += Start-Job -ScriptBlock $worker -ArgumentList $stack.TransactionPort, $reference, $Amount, $index
    }

    Wait-Job -Job $jobs | Out-Null
    $results = @($jobs | Receive-Job)
    $jobs | Remove-Job -Force | Out-Null

    $failures = @($results | Where-Object { -not $_.ok -or $_.status -ne 'success' })
    if ($failures.Count -gt 0) {
        throw "Concurrency test encountered $($failures.Count) failed requests"
    }

    $firstWriters = @($results | Where-Object { -not $_.duplicate })
    $retries = @($results | Where-Object { $_.duplicate })

    $env:PGPASSWORD = '123123'
    $transactionCount = [int](psql -h localhost -U postgres -d ddbank -At -c "select count(*) from ledger_transactions where reference = '$reference';")
    $journalCount = [int](psql -h localhost -U postgres -d ddbank -At -c "select count(*) from journal_entries je join ledger_transactions lt on lt.id = je.transaction_id where lt.reference = '$reference';")

    if ($firstWriters.Count -ne 1) {
        throw "Expected exactly 1 non-duplicate response, got $($firstWriters.Count)"
    }
    if ($retries.Count -ne ($Requests - 1)) {
        throw "Expected $($Requests - 1) duplicate responses, got $($retries.Count)"
    }
    if ($transactionCount -ne 1) {
        throw "Expected 1 persisted transaction for duplicate race, got $transactionCount"
    }
    if ($journalCount -ne 2) {
        throw "Expected 2 journal entries for duplicate race, got $journalCount"
    }

    [pscustomobject]@{
        requests = $Requests
        reference = $reference
        first_writer_count = $firstWriters.Count
        duplicate_count = $retries.Count
        persisted_transactions = $transactionCount
        journal_entries = $journalCount
        transaction_id = $firstWriters[0].transaction_id
    } | ConvertTo-Json -Depth 6
} catch {
    Write-DDBankDiagnostics -Stack $stack
    throw
} finally {
    if (-not $KeepRunning) {
        Stop-DDBankStack -Stack $stack
    }
}
