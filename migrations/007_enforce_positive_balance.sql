-- Migration: Enforce non-negative account balances
-- Add CHECK constraint to prevent negative balances
-- This is a data integrity safeguard for Phase 4

-- Add CHECK constraint to ensure balance never goes negative
ALTER TABLE accounts ADD CONSTRAINT chk_accounts_balance_nonnegative CHECK (balance >= 0);

-- Add CHECK constraint to journal entries to ensure amounts are positive
ALTER TABLE journal_entries ADD CONSTRAINT chk_journal_entry_amount_positive CHECK (amount > 0);
