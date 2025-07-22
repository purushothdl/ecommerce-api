-- Remove the indexes first
DROP INDEX IF EXISTS idx_cart_items_cart_id_created_at;
DROP INDEX IF EXISTS idx_cart_items_created_at;

-- Remove the columns
ALTER TABLE cart_items 
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS created_at;
