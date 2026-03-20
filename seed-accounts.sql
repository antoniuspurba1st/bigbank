-- Seed accounts for DD Bank
-- Run with: psql -h localhost -U postgres -d ddbank -f scripts/seed-accounts.sql

INSERT INTO accounts (id, account_number, owner_name, created_at) VALUES
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ACC-001', 'Alice', NOW()),
('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'ACC-002', 'Bob', NOW()),
('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'ACC-003', 'Charlie', NOW())
ON CONFLICT (account_number) DO NOTHING;