-- +goose Up
-- SQL миграция для модуля Delivery (EPIC-07, ADR-005, ADR-011)

CREATE SCHEMA IF NOT EXISTS delivery;

CREATE TABLE IF NOT EXISTS delivery.deliveries (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL UNIQUE,
    status VARCHAR(32) NOT NULL,
    courier_name VARCHAR(128) NOT NULL,
    courier_rating NUMERIC(3, 2) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    estimated_completion_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deliveries_order_id ON delivery.deliveries(order_id);
CREATE INDEX IF NOT EXISTS idx_deliveries_status ON delivery.deliveries(status);

-- Таблица scheduled_transitions для таймер-стейт-машины (ADR-005)
CREATE TABLE IF NOT EXISTS delivery.scheduled_transitions (
    id UUID PRIMARY KEY,
    delivery_id UUID NOT NULL REFERENCES delivery.deliveries(id) ON DELETE CASCADE,
    target_status VARCHAR(32) NOT NULL,
    fire_at TIMESTAMPTZ NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_delivery_scheduled_pending
    ON delivery.scheduled_transitions (fire_at ASC)
    WHERE processed_at IS NULL;

-- Таблица outbox_events для модуля Delivery (ADR-003, ADR-011)
CREATE TABLE IF NOT EXISTS delivery.outbox_events (
    id UUID PRIMARY KEY,
    topic VARCHAR(255) NOT NULL,
    event_key VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_delivery_outbox_events_unpublished
    ON delivery.outbox_events (created_at ASC)
    WHERE published_at IS NULL;

-- Таблица processed_events для идемпотентности входящих Kafka-событий
CREATE TABLE IF NOT EXISTS delivery.processed_events (
    event_id VARCHAR(255) NOT NULL,
    consumer_group VARCHAR(255) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, consumer_group)
);

-- +goose Down
DROP TABLE IF EXISTS delivery.processed_events;
DROP TABLE IF EXISTS delivery.outbox_events;
DROP TABLE IF EXISTS delivery.scheduled_transitions;
DROP TABLE IF EXISTS delivery.deliveries;
DROP SCHEMA IF EXISTS delivery CASCADE;
