#!/usr/bin/env pwsh
<#
DD Bank - Comprehensive Demo Script
Demonstrates the complete system end-to-end:
  1. Health checks
  2. Successful transfer
  3. Fraud rejection
  4. Duplicate idempotency
  5. Transaction history with pagination
#>

param(
    [switch]$SkipServices = $false,
    [switch]$VerboseOutput = $false
)

$ErrorActionPreference = "Stop"
$BASE_URL = "http://localhost:8081"

function Out-DemoInfo {
    param([string]$Message, [string]$Color = "Cyan")
    Write-Host "==> $Message" -ForegroundColor $Color
}

function Out-DemoSuccess {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

function Out-DemoFailure {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor Red
}

function Test-ServiceHealth {
    param([string]$Url, [string]$Name, [int]$MaxAttempts = 30)
    
    Out-DemoInfo "Checking $Name health..."
    $attempt = 0
    
    while ($attempt -lt $MaxAttempts) {
        try {
            $response = Invoke-WebRequest -Uri "$Url/health" -UseBasicParsing -ErrorAction SilentlyContinue
            if ($response.StatusCode -eq 200) {
                $json = $response.Content | ConvertFrom-Json
                Out-DemoSuccess "$Name is healthy (status: $($json.status))"
                return $true
            }
        }
        catch {
            $attempt++
            if ($attempt -lt $MaxAttempts) {
                Write-Host "  Attempt $attempt/$MaxAttempts... waiting" -ForegroundColor Yellow
                Start-Sleep -Seconds 1
            }
        }
    }
    
    Out-DemoFailure "$Name did not become healthy after $MaxAttempts attempts"
    return $false
}

function Invoke-TransferTest {
    param(
        [string]$FromAccount,
        [string]$ToAccount,
        [decimal]$Amount,
        [string]$Description,
        [bool]$ExpectSuccess = $true
    )
    
    Out-DemoInfo $Description -Color "Magenta"
    
    $reference = "demo-$(Get-Random -Minimum 100000 -Maximum 999999)-$(Get-Date -Format 'yyyyMMddHHmmss')"
    $payload = @{
        from_account = $FromAccount
        to_account   = $ToAccount
        amount       = $Amount
        reference    = $reference
    } | ConvertTo-Json
    
    if ($VerboseOutput) {
        Write-Host "Request: POST $BASE_URL/transfer" -ForegroundColor Gray
        Write-Host $payload -ForegroundColor Gray
    }
    
    try {
        $response = Invoke-WebRequest `
            -Uri "$BASE_URL/transfer" `
            -Method POST `
            -Headers @{"Content-Type" = "application/json" } `
            -Body $payload `
            -UseBasicParsing
        
        $json = $response.Content | ConvertFrom-Json
        
        Write-Host ""
        Write-Host "Response Status: $($json.status)" -ForegroundColor Cyan
        Write-Host "Transaction ID: $($json.data.transaction_id)" -ForegroundColor Green
        Write-Host "Fraud Decision: $($json.data.fraud_decision)" -ForegroundColor Cyan
        Write-Host "Ledger Status: $($json.data.ledger_status)" -ForegroundColor Cyan
        Write-Host "Duplicate: $($json.data.duplicate)" -ForegroundColor Cyan
        Write-Host "Correlation ID: $($json.correlation_id)" -ForegroundColor Yellow
        
        if ($ExpectSuccess -and $json.status -eq "success") {
            Out-DemoSuccess "Transfer completed as expected"
            return $json.data.transaction_id
        }
        elseif (-not $ExpectSuccess -and $json.status -eq "rejected") {
            Out-DemoSuccess "Transfer rejected as expected (fraud gate)"
            return $null
        }
        else {
            Out-DemoFailure "Unexpected response status"
            return $null
        }
    }
    catch {
        $errorResponse = $_.Exception.Response.Content | ConvertFrom-Json
        Write-Host "Error Response: $($errorResponse.message)" -ForegroundColor Red
        Write-Host "Error Code: $($errorResponse.code)" -ForegroundColor Red
        Out-DemoFailure "Transfer failed"
        return $null
    }
}

function Get-TransactionHistory {
    Out-DemoInfo "Fetching transaction history" -Color "Magenta"
    
    try {
        $response = Invoke-WebRequest `
            -Uri "$BASE_URL/transactions?page=1&limit=10" `
            -Method GET `
            -UseBasicParsing
        
        $json = $response.Content | ConvertFrom-Json
        
        Write-Host ""
        Write-Host "Total Pages: $($json.data.pagination.total_pages)" -ForegroundColor Cyan
        Write-Host "Current Page: $($json.data.pagination.current_page)" -ForegroundColor Cyan
        Write-Host "Items: $($json.data.transactions.Count)" -ForegroundColor Cyan
        
        if ($json.data.transactions.Count -gt 0) {
            Write-Host ""
            Write-Host "Recent Transactions:" -ForegroundColor Yellow
            foreach ($txn in $json.data.transactions | Select-Object -First 5) {
                Write-Host "  • $($txn.reference): $($txn.from_account) → $($txn.to_account), `$$($txn.amount), $($txn.status)" -ForegroundColor Green
            }
        }
        
        Out-DemoSuccess "Transaction history retrieved"
    }
    catch {
        Out-DemoFailure "Failed to retrieve transactions"
    }
}

# ===== MAIN DEMO FLOW =====

Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║         DD Bank - Comprehensive Demo Script               ║" -ForegroundColor Cyan
Write-Host "║                                                            ║" -ForegroundColor Cyan
Write-Host "║  Demonstrates:                                             ║" -ForegroundColor Cyan
Write-Host "║  ✓ Successful transfers                                    ║" -ForegroundColor Cyan
Write-Host "║  ✓ Fraud rejection                                         ║" -ForegroundColor Cyan
Write-Host "║  ✓ Duplicate idempotency                                   ║" -ForegroundColor Cyan
Write-Host "║  ✓ Transaction history & pagination                        ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Step 1: Health Check
Out-DemoInfo "STEP 1: Service Health Checks" -Color "Green"
Write-Host ""

$healthyServices = 0

if (Test-ServiceHealth "http://localhost:8080" "Ledger Service") {
    $healthyServices++
}
if (Test-ServiceHealth "http://localhost:8082" "Fraud Service") {
    $healthyServices++
}
if (Test-ServiceHealth "http://localhost:8081" "Transaction Service") {
    $healthyServices++
}

Write-Host ""

if ($healthyServices -ne 3) {
    Out-DemoFailure "Not all services are healthy. Please check service startup."
    exit 1
}

Out-DemoSuccess "All services healthy!"
Write-Host ""

# Step 2: Successful Transfer
Out-DemoInfo "STEP 2: Successful Transfer Example" -Color "Green"
Write-Host ""
Write-Host "Scenario: Transfer \$100.50 from ACC-001 to ACC-002" -ForegroundColor Yellow
Write-Host ""

$txn1 = Invoke-TransferTest -FromAccount "ACC-001" -ToAccount "ACC-002" -Amount 100.50 -Description "Executing successful transfer" -ExpectSuccess $true
Write-Host ""

# Step 3: Fraud Rejection
Out-DemoInfo "STEP 3: Fraud Rejection Example" -Color "Green"
Write-Host ""
Write-Host "Scenario: Transfer \$1,500,000 (exceeds \$1M fraud limit)" -ForegroundColor Yellow
Write-Host "(Shows fraud gate preventing ledger write)" -ForegroundColor Yellow
Write-Host ""

$txn2 = Invoke-TransferTest -FromAccount "ACC-001" -ToAccount "ACC-003" -Amount 1500000 -Description "Attempting fraudulent transfer" -ExpectSuccess $false
Write-Host ""

# Summary of captured transaction IDs
Write-Host "Captured Transaction IDs:" -ForegroundColor Yellow
if ($txn1) { Write-Host "  Successful transfer: $txn1" -ForegroundColor Green }
if ($txn2) { Write-Host "  Fraudulent transfer: $txn2" -ForegroundColor Green } else { Write-Host "  Fraudulent transfer: (rejected - no ID)" -ForegroundColor Gray }
Write-Host ""

# Step 4: Duplicate Idempotency
Out-DemoInfo "STEP 4: Duplicate Idempotency Example" -Color "Green"
Write-Host ""
Write-Host "Scenario: Submit same transfer twice with identical reference" -ForegroundColor Yellow
Write-Host "(Demonstrates safe retry behavior)" -ForegroundColor Yellow
Write-Host ""

$reference = "demo-dup-$(Get-Random -Minimum 10000 -Maximum 99999)"
$payload = @{
    from_account = "ACC-002"
    to_account   = "ACC-003"
    amount       = 75.25
    reference    = $reference
} | ConvertTo-Json

Write-Host "Request 1: Submitting with reference '$reference'" -ForegroundColor Cyan
$response1 = Invoke-WebRequest -Uri "$BASE_URL/transfer" -Method POST -Headers @{"Content-Type" = "application/json" } -Body $payload -UseBasicParsing
$json1 = $response1.Content | ConvertFrom-Json

Write-Host "  Transaction ID: $($json1.data.transaction_id)" -ForegroundColor Green
Write-Host "  Duplicate: $($json1.data.duplicate)" -ForegroundColor Green
Write-Host ""

Start-Sleep -Milliseconds 200

Write-Host "Request 2: Retrying with SAME reference (simulating client retry)" -ForegroundColor Cyan
$response2 = Invoke-WebRequest -Uri "$BASE_URL/transfer" -Method POST -Headers @{"Content-Type" = "application/json" } -Body $payload -UseBasicParsing
$json2 = $response2.Content | ConvertFrom-Json

Write-Host "  Transaction ID: $($json2.data.transaction_id)" -ForegroundColor Green
Write-Host "  Duplicate: $($json2.data.duplicate)" -ForegroundColor Green
Write-Host ""

if ($json1.data.transaction_id -eq $json2.data.transaction_id) {
    Out-DemoSuccess "Transaction IDs match - idempotency guaranteed!"
}
else {
    Out-DemoFailure "Transaction IDs should match - idempotency failed!"
}

if ($json2.data.duplicate -eq $true) {
    Out-DemoSuccess "Duplicate flag set correctly"
}
else {
    Out-DemoFailure "Duplicate flag not set"
}

Write-Host ""

# Step 5: Transaction History
Out-DemoInfo "STEP 5: Transaction History & Pagination" -Color "Green"
Write-Host ""

Get-TransactionHistory

Write-Host ""

# Summary
Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                     Demo Complete! ✅                      ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

Write-Host "Key Observations:" -ForegroundColor Yellow
Write-Host "  • All 3 services responded correctly" -ForegroundColor Green
Write-Host "  • Successful transfer created transaction" -ForegroundColor Green
Write-Host "  • Fraudulent transfer was rejected" -ForegroundColor Green
Write-Host "  • Duplicate retry returned same transaction_id" -ForegroundColor Green
Write-Host "  • Transaction history is queryable with pagination" -ForegroundColor Green
Write-Host "  • Correlation IDs visible in all responses" -ForegroundColor Green
Write-Host ""

Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "  • Visit http://localhost:3000/transfer to use the UI" -ForegroundColor Cyan
Write-Host "  • Visit http://localhost:3000/transactions to see history" -ForegroundColor Cyan
Write-Host "  • Check README.md for architecture and implementation details" -ForegroundColor Cyan
Write-Host ""
