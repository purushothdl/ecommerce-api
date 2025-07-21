-- migrations/000007_create_cart_tables.up.sql

CREATE TABLE IF NOT EXISTS carts (
    id bigserial PRIMARY KEY,
    -- user_id is nullable because anonymous users can have carts.
    -- ON DELETE SET NULL means if a user is deleted, their cart becomes an anonymous cart.
    -- Or you could use ON DELETE CASCADE if you want the cart to be deleted with the user.
    -- Let's go with CASCADE as it's cleaner.
    user_id bigint REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cart_items (
    id bigserial PRIMARY KEY,
    cart_id bigint NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id bigint NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity integer NOT NULL CHECK (quantity > 0),
    
    -- This constraint prevents having multiple rows for the same product in the same cart.
    -- We'll just update the quantity of the existing row instead.
    UNIQUE (cart_id, product_id)
);

-- Add an index for faster lookups of a user's cart
CREATE INDEX IF NOT EXISTS idx_carts_user_id ON carts(user_id);