-- migrations/000007_create_cart_tables.down.sql

-- Drop the cart_items table first (due to foreign key dependencies)
DROP TABLE IF EXISTS cart_items;

-- Drop the carts table
DROP TABLE IF EXISTS carts;

-- Drop the index (optional, but cleaner for a full rollback)
DROP INDEX IF EXISTS idx_carts_user_id;