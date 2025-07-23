-- 000012_create_payment_intents_table.down.sql
-- Drop payment intents table and related indexes

DROP INDEX IF EXISTS idx_payment_intents_order_id;
DROP INDEX IF EXISTS idx_payment_intents_status;

DROP TABLE IF EXISTS payment_intents;
