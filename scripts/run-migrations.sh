#!/bin/bash
# Phase 7.3: Database Migration Runner
# Applies all versioned SQL migrations to the database
# Usage: ./run-migrations.sh [environment]
# Example: ./run-migrations.sh production

set -e

# Get environment from parameter or use development
ENV="${1:-development}"

# Load environment variables
if [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Validate DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL environment variable not set"
    echo "Please set DATABASE_URL or source .env file"
    exit 1
fi

echo "====================================="
echo "Database Migrations - Environment: $ENV"
echo "====================================="

# Find all migration files in ascending order
MIGRATIONS=$(find migrations -maxdepth 1 -name "*.sql" -type f | sort)

if [ -z "$MIGRATIONS" ]; then
    echo "No migration files found in migrations/ directory"
    exit 1
fi

# Counter for applied migrations
MIGRATION_COUNT=0

echo ""
echo "Found migrations to apply:"
for migration in $MIGRATIONS; do
    echo "  - $(basename $migration)"
done
echo ""

# Apply each migration
for migration in $MIGRATIONS; do
    MIGRATION_NAME=$(basename "$migration")
    echo "Applying migration: $MIGRATION_NAME"
    
    # Extract migration version from filename (e.g., "001" from "001_*.sql")
    VERSION=$(echo "$MIGRATION_NAME" | cut -d'_' -f1)
    
    # Use psql to execute migration
    # The database URL already contains credentials
    psql "$DATABASE_URL" -f "$migration"
    
    if [ $? -eq 0 ]; then
        echo "✓ Migration $MIGRATION_NAME applied successfully"
        ((MIGRATION_COUNT++))
    else
        echo "✗ Migration $MIGRATION_NAME failed"
        exit 1
    fi
    echo ""
done

echo "====================================="
echo "Migrations Complete"
echo "Total migrations applied: $MIGRATION_COUNT"
echo "====================================="
