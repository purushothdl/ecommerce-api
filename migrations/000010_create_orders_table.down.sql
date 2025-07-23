-- 000010_create_orders_table.down.sql
-- Drop orders table, related indexes, and custom types

DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_payment_status;
DROP INDEX IF EXISTS idx_orders_order_number;
DROP INDEX IF EXISTS idx_orders_created_at;

DROP TABLE IF EXISTS orders;

DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS payment_status;
