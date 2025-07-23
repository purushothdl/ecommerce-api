-- 000009_create_user_addresses_table.down.sql
-- Drop user addresses table and related indexes

DROP INDEX IF EXISTS idx_user_addresses_user_id;
DROP INDEX IF EXISTS idx_user_addresses_default_shipping;
DROP INDEX IF EXISTS idx_user_addresses_default_billing;
DROP INDEX IF EXISTS idx_user_addresses_unique_default_shipping;
DROP INDEX IF EXISTS idx_user_addresses_unique_default_billing;

DROP TABLE IF EXISTS user_addresses;
