---
id: EPIC-06
title: Payment Abstraction
module: [payment]
priority: P0
status: todo
depends_on: [EPIC-05]
covers_requirements: [FR-PAY-01, INV-01]
related_adrs: [ADR-006, ADR-003]
scope:
  paths:
    - apps/api/internal/modules/payment/**
    - apps/api/internal/contracts/payment.go
    - apps/api/migrations/payment/**
    - docs/api/openapi.yaml   # секция /payments/* (webhook эндпоинты)
---

# EPIC-06: Payment Abstraction

## Цель

Реализовать `contracts.PaymentProvider` (ADR-006) с двумя адаптерами — `MockProvider` и `YooKassaProvider` — переключаемыми конфигурацией, без изменения домена `order`.

## Задачи

1. `internal/contracts/payment.go`: финализировать интерфейс `PaymentProvider` (Initiate, VerifyWebhook), типы `InitiatePaymentRequest/Result`, `WebhookResult` (нормализованный статус: succeeded/failed/pending).
2. `domain/payment.go`: сущность `Payment` (id, order_id, provider, provider_payment_id, amount_rub, status, created_at) — модуль хранит СВОЙ учёт платежей отдельно от `order` (разделение ответственности: order знает "оплачен/не оплачен", payment знает детали транзакции у провайдера).
3. `usecase/InitiatePayment`: вызывается consumer'ом `order.created.v1`, вызывает `PaymentProvider.Initiate`, пишет `Payment` + outbox (нет прямого события "успеха" здесь — это придёт через webhook).
4. `usecase/HandleWebhook`: вызывается HTTP хендлером вебхука, вызывает `PaymentProvider.VerifyWebhook`, обновляет `Payment.status`, публикует `payment.succeeded.v1` ИЛИ `payment.failed.v1` через outbox — идемпотентно по `provider_payment_id` (повторный вебхук с тем же ID не публикует событие дважды).
5. `adapter/mock/provider.go`: `Initiate` возвращает `pending`, планирует webhook самовызов через `delivery`-подобный scheduler механизм (реюз `internal/platform/scheduler`, вынесенного как разделяемый примитив — оценить при реализации: и payment mock, и delivery (EPIC-07) нуждаются в "запланировать действие через N секунд, устойчиво к рестарту" — это платформенный примитив, не специфичный для delivery, несмотря на то что ADR-005 писан в контексте delivery). Success rate 95%, конфигурируемый через env для тестов (`MOCK_PAYMENT_SUCCESS_RATE`).
6. `adapter/yookassa/provider.go`: реальный HTTP-клиент к YooKassa API (`https://api.yookassa.ru/v3/payments`), `Idempotence-Key` header = order_id, `VerifyWebhook` — HMAC/IP-проверка согласно текущей документации YooKassa (проверить актуальный метод верификации на момент реализации — YooKassa могла обновить механизм, не полагаться на память, свериться с официальной документацией).
7. `adapter/postgres`: схема `payment`, таблицы `payments`, `outbox_events`, `processed_events` (для идемпотентности вебхуков).
8. `adapter/httpapi`: `POST /payments/webhook/mock` (только доступен если `PAYMENT_PROVIDER=mock`, иначе 404), `POST /payments/webhook/yookassa`.
9. Composition root (`cmd/server/main.go`): выбор адаптера по `PAYMENT_PROVIDER=mock|yookassa` env, fail-fast если `yookassa` выбран без секретов.
10. Contract test suite: один набор тестов, параметризованный по обеим реализациям `PaymentProvider` — гарантирует поведенческую эквивалентность интерфейса (ADR-006 последствие "оба адаптера нужно покрыть contract-тестами").
11. Тесты: unit на usecase с mock-портами, contract-тесты обоих адаптеров (YooKassa — против sandbox или через записанные HTTP-фикстуры/VCR-подход, чтобы не требовать реальной сети в CI по умолчанию), тест идемпотентности вебхука.

## Acceptance Criteria

- [ ] Сумма платежа всегда 10.00 RUB (`INV-01`)
- [ ] MockProvider эмулирует webhook с задержкой 1-3с, success rate ~95% (`FR-PAY-01`)
- [ ] YooKassaProvider инициирует реальный платёж в тестовом магазине и корректно проверяет подпись вебхука (`FR-PAY-01`)
- [ ] Оба адаптера проходят один и тот же contract test suite (`ADR-006`)
- [ ] Повторный вебхук с тем же `provider_payment_id` не публикует повторное событие (`NFR-REL-02`)
- [ ] Переключение `PAYMENT_PROVIDER` меняет поведение без изменения кода `order`/`payment` domain/usecase слоёв

## Definition of Done

Стандартный DoD. В PR description явно указать: как локально протестировать YooKassa-адаптер (какие env переменные/sandbox credentials нужны получить и куда положить, не коммитить секреты).
