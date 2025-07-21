-- migrations/000005_create_products_table.down.sql

-- Drop the products table first (due to foreign key dependency)
DROP TABLE IF EXISTS products;

-- Then drop the categories table
DROP TABLE IF EXISTS categories;