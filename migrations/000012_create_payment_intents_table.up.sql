-- 000012_create_payment_intents.sql
-- Create payment intents table for Stripe integration

CREATE TABLE payment_intents (
    id VARCHAR(255) PRIMARY KEY, -- Stripe payment intent ID
    order_id BIGINT NOT NULL REFERENCES orders(id),
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL,
    client_secret VARCHAR(500) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_payment_intents_order_id ON payment_intents(order_id);
CREATE INDEX idx_payment_intents_status ON payment_intents(status);
