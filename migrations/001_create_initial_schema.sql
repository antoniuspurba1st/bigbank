-- Phase 7.3: Migration 001 - Initial Database Schema
-- Creates core tables for ledger, accounts, and transactions

-- Create accounts table
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_number VARCHAR(50) UNIQUE NOT NULL,
    owner_name VARCHAR(255) NOT NULL,
    balance DECIMAL(19, 2) DEFAULT 0.00 CHECK (balance >= 0),
    user_id UUID UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create journal_entries table for double-entry accounting
CREATE TABLE IF NOT EXISTS journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    entry_type VARCHAR(10) NOT NULL CHECK (entry_type IN ('DEBIT', 'CREDIT')),
    amount DECIMAL(19, 2) NOT NULL CHECK (amount > 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (transaction_id) REFERENCES ledger_transactions(id) ON DELETE CASCADE
);

-- Create ledger_transactions table
CREATE TABLE IF NOT EXISTS ledger_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_accounts_account_number ON accounts(account_number);
CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_journal_entries_account ON journal_entries(account_id);
CREATE INDEX IF NOT EXISTS idx_journal_entries_transaction ON journal_entries(transaction_id);
CREATE INDEX IF NOT EXISTS idx_ledger_transactions_reference ON ledger_transactions(reference);
CREATE INDEX IF NOT EXISTS idx_ledger_transactions_status ON ledger_transactions(status);
