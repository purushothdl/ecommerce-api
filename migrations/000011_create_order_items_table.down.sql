-- 000011_create_order_items_table.down.sql
-- Drop order items table and related indexes

DROP INDEX IF EXISTS idx_order_items_order_id;
DROP INDEX IF EXISTS idx_order_items_product_id;

DROP TABLE IF EXISTS order_items;
