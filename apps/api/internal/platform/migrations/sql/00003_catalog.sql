-- +goose Up
CREATE SCHEMA IF NOT EXISTS catalog;

CREATE TABLE IF NOT EXISTS catalog.categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS catalog.products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES catalog.categories(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price_rub NUMERIC(10, 2) NOT NULL DEFAULT 10.00,
    image_seed VARCHAR(255) NOT NULL DEFAULT '',
    tsv TSVECTOR GENERATED ALWAYS AS (
        setweight(to_tsvector('russian', coalesce(name, '')), 'A') ||
        setweight(to_tsvector('russian', coalesce(description, '')), 'B')
    ) STORED,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_category_id ON catalog.products (category_id);
CREATE INDEX IF NOT EXISTS idx_products_tsv ON catalog.products USING GIN(tsv);

-- +goose Down
DROP TABLE IF EXISTS catalog.products;
DROP TABLE IF EXISTS catalog.categories;
DROP SCHEMA IF EXISTS catalog CASCADE;
