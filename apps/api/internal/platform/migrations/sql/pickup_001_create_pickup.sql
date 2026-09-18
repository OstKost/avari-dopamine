-- +goose Up
CREATE SCHEMA IF NOT EXISTS pickup;

CREATE TABLE IF NOT EXISTS pickup.pickup_points (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    latitude NUMERIC(10, 7) NOT NULL,
    longitude NUMERIC(10, 7) NOT NULL,
    distance_meters NUMERIC(10, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pickup_points_user_id ON pickup.pickup_points (user_id);

-- +goose Down
DROP TABLE IF EXISTS pickup.pickup_points;
DROP SCHEMA IF EXISTS pickup CASCADE;
