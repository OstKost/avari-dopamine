-- +goose Up
-- SQL миграция для модуля Payment (EPIC-06, ADR-006, ADR-011)

CREATE SCHEMA IF NOT EXISTS payment;

CREATE TABLE IF NOT EXISTS payment.payments (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL UNIQUE,
    provider VARCHAR(64) NOT NULL,
    provider_payment_id VARCHAR(255) NOT NULL UNIQUE,
    amount_rub NUMERIC(10, 2) NOT NULL DEFAULT 10.00,
    status VARCHAR(32) NOT NULL,
    confirmation_url TEXT,
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payment.payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_provider_payment_id ON payment.payments(provider_payment_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payment.payments(status);

-- Таблица outbox_events для модуля Payment (ADR-003, ADR-011)
CREATE TABLE IF NOT EXISTS payment.outbox_events (
    id UUID PRIMARY KEY,
    topic VARCHAR(255) NOT NULL,
    event_key VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_payment_outbox_events_unpublished
    ON payment.outbox_events (created_at ASC)
    WHERE published_at IS NULL;

-- Таблица processed_events для идемпотентности обработки входящих событий и вебхуков
CREATE TABLE IF NOT EXISTS payment.processed_events (
    event_id VARCHAR(255) NOT NULL,
    consumer_group VARCHAR(255) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, consumer_group)
);

-- +goose Down
DROP TABLE IF EXISTS payment.processed_events;
DROP TABLE IF EXISTS payment.outbox_events;
DROP TABLE IF EXISTS payment.payments;
DROP SCHEMA IF EXISTS payment CASCADE;
