$ErrorActionPreference = 'Stop'
$env:PGPASSWORD = '123123'

$tables = psql -h localhost -U postgres -d ddbank -At -c "select tablename from pg_tables where schemaname='public' and tablename in ('accounts','ledger_transactions','journal_entries') order by tablename;"
$columns = psql -h localhost -U postgres -d ddbank -At -F '|' -c "select table_name, column_name, data_type from information_schema.columns where table_schema='public' and table_name in ('accounts','ledger_transactions','journal_entries') order by table_name, ordinal_position;"
$indexes = psql -h localhost -U postgres -d ddbank -At -F '|' -c "select indexname, indexdef from pg_indexes where schemaname='public' and tablename='ledger_transactions' order by indexname;"
$constraints = psql -h localhost -U postgres -d ddbank -At -F '|' -c "select tc.table_name, tc.constraint_type, coalesce(kcu.column_name, ''), coalesce(ccu.table_name, '') from information_schema.table_constraints tc left join information_schema.key_column_usage kcu on tc.constraint_name = kcu.constraint_name and tc.table_schema = kcu.table_schema left join information_schema.constraint_column_usage ccu on tc.constraint_name = ccu.constraint_name and tc.table_schema = ccu.table_schema where tc.table_schema='public' and tc.table_name in ('accounts','ledger_transactions','journal_entries') order by tc.table_name, tc.constraint_type, kcu.column_name;"

$requiredTables = @('accounts', 'journal_entries', 'ledger_transactions')
foreach ($table in $requiredTables) {
    if ($table -notin $tables) {
        throw "Missing required table: $table"
    }
}

if (-not ($indexes | Where-Object { $_ -match 'UNIQUE INDEX .* \(reference\)' })) {
    throw 'Missing unique index or constraint on ledger_transactions.reference'
}

if (-not ($constraints | Where-Object { $_ -match '^journal_entries\|FOREIGN KEY\|account_id\|accounts$' })) {
    throw 'Missing foreign key from journal_entries.account_id to accounts'
}

if (-not ($constraints | Where-Object { $_ -match '^journal_entries\|FOREIGN KEY\|transaction_id\|ledger_transactions$' })) {
    throw 'Missing foreign key from journal_entries.transaction_id to ledger_transactions'
}

[pscustomobject]@{
    tables = @($tables)
    columns = @($columns)
    indexes = @($indexes)
    constraints = @($constraints)
} | ConvertTo-Json -Depth 6
