param(
    [string]$TransactionServiceUrl = "http://localhost:8081",
    [string]$Reference = "test-fraud-$(Get-Date -Format 'yyyyMMddHHmmss')"
)

$ErrorActionPreference = 'Stop'

$payload = @{
    reference    = $Reference
    from_account = "ACC-001"
    to_account   = "ACC-002"
    amount       = 1500000.00
} | ConvertTo-Json

Write-Host "Testing FRAUD REJECTION scenario..." -ForegroundColor Cyan
Write-Host "Reference: $Reference"
Write-Host "Amount: 1500000.00 (exceeds fraud limit of 1,000,000)"
Write-Host ""

try {
    $response = Invoke-WebRequest `
        -Uri "$TransactionServiceUrl/transfer" `
        -Method POST `
        -Headers @{ "Content-Type" = "application/json" } `
        -Body $payload `
        -TimeoutSec 10 `
        -UseBasicParsing

    $body = $response.Content | ConvertFrom-Json
    $correlationId = $response.Headers['X-Correlation-Id'][0]

    Write-Host "Status Code: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Correlation ID: $correlationId"
    Write-Host ""
    Write-Host "Response:" -ForegroundColor Yellow
    $body | ConvertTo-Json -Depth 5 | Write-Host

    if ($body.status -eq "rejected") {
        Write-Host ""
        Write-Host "✅ FRAUD REJECTION TEST PASSED" -ForegroundColor Green
        exit 0
    }
    else {
        Write-Host ""
        Write-Host "❌ Expected status=rejected but got status=$($body.status)" -ForegroundColor Red
        exit 1
    }
}
catch {
    Write-Host "❌ TEST FAILED" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit 1
}
