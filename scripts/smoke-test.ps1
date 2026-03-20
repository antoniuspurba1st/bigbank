param(
    [switch]$KeepRunning,
    [int]$LedgerPort = 18080,
    [int]$TransactionPort = 18081,
    [int]$FraudPort = 18082
)

. "$PSScriptRoot/common.ps1"

$stack = Start-DDBankStack -LedgerPort $LedgerPort -TransactionPort $TransactionPort -FraudPort $FraudPort

try {
    $health = Wait-DDBankReady -Stack $stack
    $stamp = Get-Date -Format 'yyyyMMddHHmmss'
    $smallReference = "ref-smoke-$stamp-small"
    $largeReference = "ref-smoke-$stamp-large"
    $duplicateReference = "ref-smoke-$stamp-duplicate"

    $smallTransfer = (Invoke-DDBankTransfer -TransactionPort $stack.TransactionPort -Reference $smallReference -Amount 150.25 -CorrelationId 'corr-smoke-success').Content | ConvertFrom-Json
    $largeTransfer = (Invoke-DDBankTransfer -TransactionPort $stack.TransactionPort -Reference $largeReference -Amount 2000000.00 -CorrelationId 'corr-smoke-reject').Content | ConvertFrom-Json
    $duplicateFirst = (Invoke-DDBankTransfer -TransactionPort $stack.TransactionPort -Reference $duplicateReference -Amount 88.00).Content | ConvertFrom-Json
    $duplicateSecond = (Invoke-DDBankTransfer -TransactionPort $stack.TransactionPort -Reference $duplicateReference -Amount 88.00).Content | ConvertFrom-Json
    $transactions = (Invoke-WebRequest -Uri "http://127.0.0.1:$($stack.LedgerPort)/ledger/transactions?limit=10" -UseBasicParsing -TimeoutSec 10).Content | ConvertFrom-Json

    if ($smallTransfer.status -ne 'success') {
        throw 'Expected small transfer to succeed'
    }
    if ($largeTransfer.status -ne 'rejected') {
        throw 'Expected large transfer to be rejected'
    }
    if (-not $duplicateSecond.data.duplicate) {
        throw 'Expected duplicate retry to return duplicate=true'
    }

    $env:PGPASSWORD = '123123'
    $dbCount = [int](psql -h localhost -U postgres -d ddbank -At -c "select count(*) from ledger_transactions where reference in ('$smallReference', '$duplicateReference', '$largeReference');")

    if ($dbCount -ne 2) {
        throw "Expected 2 persisted transactions from smoke test, got $dbCount"
    }

    [pscustomobject]@{
        health = [pscustomobject]@{
            ledger = ($health.Ledger.Content | ConvertFrom-Json)
            fraud = ($health.Fraud.Content | ConvertFrom-Json)
            transaction = ($health.Transaction.Content | ConvertFrom-Json)
        }
        small_transfer = $smallTransfer
        large_transfer = $largeTransfer
        duplicate_first = $duplicateFirst
        duplicate_second = $duplicateSecond
        transactions = $transactions
        persisted_transactions = $dbCount
    } | ConvertTo-Json -Depth 8
} catch {
    Write-DDBankDiagnostics -Stack $stack
    throw
} finally {
    if (-not $KeepRunning) {
        Stop-DDBankStack -Stack $stack
    }
}
