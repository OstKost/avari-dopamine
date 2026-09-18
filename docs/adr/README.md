# Architecture Decision Records

Формат: [MADR-lite](https://adr.github.io/madr/). Каждый ADR неизменяем после статуса `accepted` — новое решение, меняющее старое, оформляется отдельным ADR со ссылкой `supersedes`.

| ID | Заголовок | Статус |
|---|---|---|
| [ADR-001](001-modular-monolith.md) | Модульный монолит вместо микросервисов | accepted |
| [ADR-002](002-single-node-infra.md) | Single-node инфраструктура для 100 MAU | accepted |
| [ADR-003](003-event-driven-outbox.md) | Kafka + Outbox pattern для событий заказа | accepted |
| [ADR-004](004-module-boundaries.md) | Границы модулей и контракты между ними | accepted |
| [ADR-005](005-delivery-state-machine.md) | Timer-driven стейт-машина доставки | accepted |
| [ADR-006](006-payment-abstraction.md) | Абстракция PaymentProvider (Mock + YooKassa) | accepted |
| [ADR-007](007-auth-jwt-sessions.md) | JWT access + refresh с ротацией в Redis | accepted |
| [ADR-008](008-cart-in-redis.md) | Корзина в Redis, а не в Postgres | accepted |
| [ADR-009](009-nextjs-ssr-realtime.md) | Next.js SSR + SSE для real-time статусов | accepted |
| [ADR-010](010-observability-graylog.md) | Структурные логи → Graylog, трейсинг через OTel | accepted |
| [ADR-011](011-db-per-module-schema.md) | Postgres: отдельная схема на модуль, единая БД | accepted |
| [ADR-012](012-synthetic-data-generation.md) | Детерминированная генерация синтетических данных | accepted |
