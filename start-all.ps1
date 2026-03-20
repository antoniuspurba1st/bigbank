Write-Host "Starting DD Bank Microservices..." -ForegroundColor Cyan

$root = Resolve-Path "$PSScriptRoot/.."

# Function to start a service in a new PowerShell window
function Start-Component {
    param(
        [string]$Name,
        [string]$Path,
        [string]$Command
    )
    Write-Host "Starting $Name..." -ForegroundColor Green
    $processInfo = New-Object System.Diagnostics.ProcessStartInfo
    $processInfo.FileName = "pwsh"
    $processInfo.Arguments = "-NoExit -Command ""cd '$Path'; $Command"""
    $processInfo.UseShellExecute = $true
    [System.Diagnostics.Process]::Start($processInfo) | Out-Null
}

# 1. Start Ledger Service (Kotlin)
Start-Component -Name "Ledger Service (Port 8080)" -Path "$root/ls_springboot" -Command "./gradlew.bat bootRun"

# 2. Start Fraud Service (Rust)
Start-Component -Name "Fraud Service (Port 8082)" -Path "$root/fraud-service" -Command "cargo run --release"

# 3. Start Transaction Service (Go)
Start-Component -Name "Transaction Service (Port 8081)" -Path "$root/transaction-service" -Command "go run ./cmd/main.go"

# 4. Start UI (Next.js)
Start-Component -Name "Web UI (Port 3000)" -Path "$root/ui" -Command "npm run dev"

Write-Host "All services started! Please wait ~30s for Java/Rust compilation." -ForegroundColor Yellow