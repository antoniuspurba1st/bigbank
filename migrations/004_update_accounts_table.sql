-- Migration: Update accounts table for 1:1 relationship with users
-- Add user_id with foreign key to users.id and unique constraint
-- Add balance column

-- Add user_id column
ALTER TABLE accounts ADD COLUMN user_id UUID;

-- Add foreign key constraint
ALTER TABLE accounts ADD CONSTRAINT fk_accounts_user_id FOREIGN KEY (user_id) REFERENCES users(id);

-- Add unique constraint on user_id to ensure 1:1 relationship
ALTER TABLE accounts ADD CONSTRAINT uq_accounts_user_id UNIQUE (user_id);

-- Add balance column with default 0
ALTER TABLE accounts ADD COLUMN balance DECIMAL(19,2) DEFAULT 0.00 NOT NULL;