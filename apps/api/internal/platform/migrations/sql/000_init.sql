-- +goose Up
-- Базовые расширения Postgres, нужные нескольким модулям.
-- pgcrypto: gen_random_uuid() для генерации UUID первичных ключей без
-- зависимости от uuid-ossp (pgcrypto — рекомендуемый современный выбор).
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose Down
DROP EXTENSION IF EXISTS pgcrypto;
