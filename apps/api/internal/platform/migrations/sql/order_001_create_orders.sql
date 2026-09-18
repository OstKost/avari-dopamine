-- +goose Up
CREATE SCHEMA IF NOT EXISTS "order";

CREATE TABLE IF NOT EXISTS "order".orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    total_amount_rub NUMERIC(10, 2) NOT NULL DEFAULT 10.00,
    pickup_point_id UUID,
    pickup_point_name VARCHAR(255),
    pickup_point_latitude NUMERIC(10, 7),
    pickup_point_longitude NUMERIC(10, 7),
    pickup_point_distance_meters NUMERIC(10, 2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_created ON "order".orders (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_user_status ON "order".orders (user_id, status);

CREATE TABLE IF NOT EXISTS "order".order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES "order".orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    category_name VARCHAR(255) NOT NULL DEFAULT '',
    price_rub NUMERIC(10, 2) NOT NULL DEFAULT 10.00,
    image_seed VARCHAR(255) NOT NULL DEFAULT '',
    quantity INT NOT NULL,
    subtotal_rub NUMERIC(10, 2) NOT NULL DEFAULT 10.00,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON "order".order_items (order_id);

CREATE TABLE IF NOT EXISTS "order".order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES "order".orders(id) ON DELETE CASCADE,
    from_status VARCHAR(50) NOT NULL,
    to_status VARCHAR(50) NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_status_history_order ON "order".order_status_history (order_id, created_at);

CREATE TABLE IF NOT EXISTS "order".outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic VARCHAR(255) NOT NULL,
    event_key VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbox_events_unpublished ON "order".outbox_events (created_at) WHERE published_at IS NULL;

CREATE TABLE IF NOT EXISTS "order".processed_events (
    event_id VARCHAR(255) PRIMARY KEY,
    consumer_name VARCHAR(255) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS "order".processed_events;
DROP TABLE IF EXISTS "order".outbox_events;
DROP TABLE IF EXISTS "order".order_status_history;
DROP TABLE IF EXISTS "order".order_items;
DROP TABLE IF EXISTS "order".orders;
DROP SCHEMA IF EXISTS "order" CASCADE;
