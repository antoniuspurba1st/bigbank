# Phase 7.3: Database Migration Runner (PowerShell)
# Applies all versioned SQL migrations to the database
# Usage: .\run-migrations.ps1 -environment development
# Example: .\run-migrations.ps1 -environment production

param(
    [string]$environment = "development",
    [string]$databaseUrl = ""
)

Write-Host "=====================================" -ForegroundColor Green
Write-Host "Database Migrations - Environment: $environment" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green

# Load .env file if it exists
if (Test-Path ".env") {
    Write-Host "Loading environment from .env file..." -ForegroundColor Cyan
    Get-Content ".env" | ForEach-Object {
        if ($_ -like "*=*" -and $_ -notlike "#*") {
            $name, $value = $_.Split('=', 2)
            [Environment]::SetEnvironmentVariable($name.Trim(), $value.Trim())
        }
    }
}

# Get DATABASE_URL from parameter or environment
if ([string]::IsNullOrEmpty($databaseUrl)) {
    $databaseUrl = $env:DATABASE_URL
}

# Validate DATABASE_URL is set
if ([string]::IsNullOrEmpty($databaseUrl)) {
    Write-Host "Error: DATABASE_URL not provided or not set" -ForegroundColor Red
    Write-Host "Usage: .\run-migrations.ps1 -databaseUrl 'postgres://user:pass@host:5432/db'" -ForegroundColor Yellow
    exit 1
}

# Find all migration files in ascending order
$migrations = Get-ChildItem "migrations" -Filter "*.sql" | Sort-Object Name

if ($migrations.Count -eq 0) {
    Write-Host "No migration files found in migrations/ directory" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "Found migrations to apply:" -ForegroundColor Cyan
foreach ($migration in $migrations) {
    Write-Host "  - $($migration.Name)" -ForegroundColor Gray
}
Write-Host ""

$migrationCount = 0

# Apply each migration
foreach ($migration in $migrations) {
    $migrationName = $migration.Name
    Write-Host "Applying migration: $migrationName" -ForegroundColor Cyan
    
    try {
        # Read migration SQL file
        $sqlContent = Get-Content $migration.FullName -Raw
        
        # Use psql to execute migration
        # Note: Requires psql.exe in PATH
        $env:PGPASSWORD = ""
        & psql $databaseUrl -f $migration.FullName
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ Migration $migrationName applied successfully" -ForegroundColor Green
            $migrationCount++
        }
        else {
            Write-Host "✗ Migration $migrationName failed (exit code: $LASTEXITCODE)" -ForegroundColor Red
            exit 1
        }
    }
    catch {
        Write-Host "✗ Migration $migrationName failed: $_" -ForegroundColor Red
        exit 1
    }
    
    Write-Host ""
}

Write-Host "=====================================" -ForegroundColor Green
Write-Host "Migrations Complete" -ForegroundColor Green
Write-Host "Total migrations applied: $migrationCount" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green
