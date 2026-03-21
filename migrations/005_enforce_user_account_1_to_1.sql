-- Migration: Enforce one-to-one relationship between users and accounts
-- Use operations that are safe for existing data.

BEGIN;

-- Ensure accounts has user_id and balance and created_at columns
ALTER TABLE IF EXISTS accounts
  ADD COLUMN IF NOT EXISTS user_id UUID;

ALTER TABLE IF EXISTS accounts
  ADD COLUMN IF NOT EXISTS balance DECIMAL(19,2) DEFAULT 0.00 NOT NULL;

ALTER TABLE IF EXISTS accounts
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL;

-- Ensure user_id references users(id)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
    WHERE tc.table_name = 'accounts' AND tc.constraint_type = 'FOREIGN KEY' AND kcu.column_name = 'user_id'
  ) THEN
    ALTER TABLE accounts
      ADD CONSTRAINT fk_accounts_user_id FOREIGN KEY (user_id) REFERENCES users(id);
  END IF;
END$$;

-- Ensure one account per user
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.table_constraints tc
    WHERE tc.table_name = 'accounts' AND tc.constraint_type = 'UNIQUE' AND tc.constraint_name = 'uq_accounts_user_id'
  ) THEN
    ALTER TABLE accounts
      ADD CONSTRAINT uq_accounts_user_id UNIQUE (user_id);
  END IF;
END$$;

COMMIT;
