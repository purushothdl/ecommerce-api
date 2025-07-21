-- migrations/000005_create_products_tables.up.sql

-- First, create the categories table
CREATE TABLE IF NOT EXISTS categories (
    id bigserial PRIMARY KEY,
    name text NOT NULL UNIQUE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

-- Then, create the products table that references categories
CREATE TABLE IF NOT EXISTS products (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    description text NOT NULL,
    price decimal(10, 2) NOT NULL,
    stock_quantity integer NOT NULL,
    category_id bigint NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    
    brand text,
    sku text UNIQUE,
    images text[] NOT NULL, 
    thumbnail text,
    
    -- Storing extra structured data as JSONB is a good practice
    dimensions jsonb,
    warranty_information text,
    
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    version integer NOT NULL DEFAULT 1
);

-- Add an index on category_id for faster filtering
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);