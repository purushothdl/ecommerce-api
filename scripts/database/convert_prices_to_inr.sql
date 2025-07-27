-- scripts/convert_prices_to_inr.sql
-- This is a one-time script to convert existing product prices from USD to INR.
-- We'll use an approximate exchange rate of 1 USD = 83 INR.

UPDATE products
SET price = price * 86.0;

SELECT id, name, price FROM products ORDER BY id LIMIT 10;