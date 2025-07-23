-- 000009_create_user_addresses_table.up.sql
-- Create user addresses table for address book functionality

CREATE TABLE user_addresses (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    street1 VARCHAR(255) NOT NULL,
    street2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL DEFAULT 'India',
    is_default_shipping BOOLEAN DEFAULT FALSE,
    is_default_billing BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_user_addresses_user_id ON user_addresses(user_id);
CREATE INDEX idx_user_addresses_default_shipping ON user_addresses(user_id, is_default_shipping) WHERE is_default_shipping = true;
CREATE INDEX idx_user_addresses_default_billing ON user_addresses(user_id, is_default_billing) WHERE is_default_billing = true;

-- Ensure only one default shipping/billing per user
CREATE UNIQUE INDEX idx_user_addresses_unique_default_shipping ON user_addresses(user_id) WHERE is_default_shipping = true;
CREATE UNIQUE INDEX idx_user_addresses_unique_default_billing ON user_addresses(user_id) WHERE is_default_billing = true;
