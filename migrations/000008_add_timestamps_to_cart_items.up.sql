-- Add created_at and updated_at columns to cart_items
ALTER TABLE cart_items 
ADD COLUMN created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
ADD COLUMN updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW();

-- Add index for efficient queries by creation date
CREATE INDEX IF NOT EXISTS idx_cart_items_created_at ON cart_items(created_at);
CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id_created_at ON cart_items(cart_id, created_at DESC);
