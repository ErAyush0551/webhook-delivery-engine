-- 1. Create Subscriptions Table
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY,
    target_url VARCHAR(255) NOT NULL,
    secret VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Create Webhooks Table
CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY,
    subscription_id UUID REFERENCES subscriptions(id) ON DELETE CASCADE,
    payload JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Create Delivery Attempts / Logs Table
CREATE TABLE IF NOT EXISTS delivery_attempts (
    id UUID PRIMARY KEY,
    webhook_id UUID REFERENCES webhooks(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscriptions(id) ON DELETE CASCADE,
    target_url VARCHAR(255) NOT NULL,
    attempt INT NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL, -- e.g., 'success', 'failed'
    http_status INT,             -- e.g., 200, 500, 404
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Create Indexes for Faster Lookups
CREATE INDEX IF NOT EXISTS idx_webhooks_subscription_id ON webhooks(subscription_id);
CREATE INDEX IF NOT EXISTS idx_delivery_attempts_webhook_id ON delivery_attempts(webhook_id);
CREATE INDEX IF NOT EXISTS idx_delivery_attempts_subscription_id ON delivery_attempts(subscription_id);

-- We have made this file to create the initial database schema for the HookFlow application.
-- This migration script creates three tables: subscriptions, webhooks, and delivery_attempts (logs).
-- The subscriptions table stores information about each subscription, including the target URL and secret.
-- The webhooks table stores information about each webhook, including the payload and status.
-- The delivery_attempts table stores logs of each delivery attempt, including the status, HTTP status code, and any error messages.
