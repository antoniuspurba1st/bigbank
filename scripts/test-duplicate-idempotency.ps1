param(
    [string]$TransactionServiceUrl = "http://localhost:8081",
    [string]$Reference = "test-dup-$(Get-Date -Format 'yyyyMMddHHmmss')"
)

$ErrorActionPreference = 'Stop'

$payload = @{
    reference    = $Reference
    from_account = "ACC-001"
    to_account   = "ACC-003"
    amount       = 75.25
} | ConvertTo-Json

Write-Host "Testing DUPLICATE IDEMPOTENCY scenario..." -ForegroundColor Cyan
Write-Host "Reference: $Reference"
Write-Host ""

Write-Host "STEP 1: Submit initial transfer..." -ForegroundColor Yellow

try {
    $response1 = Invoke-WebRequest `
        -Uri "$TransactionServiceUrl/transfer" `
        -Method POST `
        -Headers @{ "Content-Type" = "application/json" } `
        -Body $payload `
        -TimeoutSec 10 `
        -UseBasicParsing

    $body1 = $response1.Content | ConvertFrom-Json
    $correlationId1 = $response1.Headers['X-Correlation-Id'][0]

    Write-Host "Status Code: $($response1.StatusCode)" -ForegroundColor Green
    Write-Host "First Response:" -ForegroundColor Yellow
    $body1 | ConvertTo-Json -Depth 5 | Write-Host

    $transactionId1 = $body1.data.transaction_id
    Write-Host ""
    Write-Host "STEP 2: Submit SAME reference again (duplicate)..." -ForegroundColor Yellow

    # Small delay to avoid timestamp collisions
    Start-Sleep -Milliseconds 500

    $response2 = Invoke-WebRequest `
        -Uri "$TransactionServiceUrl/transfer" `
        -Method POST `
        -Headers @{ "Content-Type" = "application/json" } `
        -Body $payload `
        -TimeoutSec 10 `
        -UseBasicParsing

    $body2 = $response2.Content | ConvertFrom-Json
    $correlationId2 = $response2.Headers['X-Correlation-Id'][0]

    Write-Host "Status Code: $($response2.StatusCode)" -ForegroundColor Green
    Write-Host "Second Response (Duplicate):" -ForegroundColor Yellow
    $body2 | ConvertTo-Json -Depth 5 | Write-Host

    $transactionId2 = $body2.data.transaction_id
    $isDuplicate = $body2.data.duplicate

    Write-Host ""
    Write-Host "VALIDATION:" -ForegroundColor Cyan
    Write-Host "Transaction ID (first):  $transactionId1"
    Write-Host "Transaction ID (second): $transactionId2"
    Write-Host "IDs match: $(if ($transactionId1 -eq $transactionId2) { 'YES ✓' } else { 'NO ✗' })"
    Write-Host "Marked as duplicate: $(if ($isDuplicate) { 'YES ✓' } else { 'NO ✗' })"

    if ($transactionId1 -eq $transactionId2 -and $isDuplicate -eq $true) {
        Write-Host ""
        Write-Host "✅ DUPLICATE IDEMPOTENCY TEST PASSED" -ForegroundColor Green
        exit 0
    }
    else {
        Write-Host ""
        Write-Host "❌ Idempotency check failed" -ForegroundColor Red
        exit 1
    }
}
catch {
    Write-Host "❌ TEST FAILED" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit 1
}
