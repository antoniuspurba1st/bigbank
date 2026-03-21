Write-Host "Starting DD Bank Microservices..." -ForegroundColor Cyan

# Get the root directory (this is where start-all.ps1 lives)
$root = $PSScriptRoot

# Function to start a service in a new PowerShell window
function Start-Component {
    param(
        [string]$Name,
        [string]$Path,
        [string]$Command
    )
    Write-Host "Starting $Name..." -ForegroundColor Green
    
    # Resolve full path
    $fullPath = Join-Path $root $Path
    if (-not (Test-Path $fullPath)) {
        Write-Host "ERROR: Path not found: $fullPath" -ForegroundColor Red
        return
    }
    
    # Try to find PowerShell executable
    $pwsh = Get-Command pwsh -ErrorAction SilentlyContinue
    if (-not $pwsh) {
        $pwsh = Get-Command powershell -ErrorAction SilentlyContinue
    }
    
    if (-not $pwsh) {
        Write-Host "ERROR: Could not find PowerShell executable" -ForegroundColor Red
        return
    }
    
    $pshExe = $pwsh.Source
    $processInfo = New-Object System.Diagnostics.ProcessStartInfo
    $processInfo.FileName = $pshExe
    $processInfo.Arguments = "-NoExit -Command `"cd '$fullPath'; $Command`""
    $processInfo.UseShellExecute = $true
    
    try {
        [System.Diagnostics.Process]::Start($processInfo) | Out-Null
    }
    catch {
        Write-Host "ERROR starting $Name : $_" -ForegroundColor Red
    }
}

# 1. Start Ledger Service (Kotlin)
Start-Component -Name "Ledger Service (Port 8080)" -Path "ls_springboot" -Command ".\gradlew.bat bootRun"

# 2. Start Fraud Service (Rust)
Start-Component -Name "Fraud Service (Port 8082)" -Path "fraud-service" -Command "cargo run --release"

# 3. Start Transaction Service (Go)
Start-Component -Name "Transaction Service (Port 8081)" -Path "transaction-service" -Command "go run ./cmd/main.go"

# 4. Start UI (Next.js)
Start-Component -Name "Web UI (Port 3000)" -Path "ui" -Command "npm run dev"

Write-Host "All services started! Please wait ~30s for Java/Rust compilation." -ForegroundColor Yellow