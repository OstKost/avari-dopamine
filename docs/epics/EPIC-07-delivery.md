---
id: EPIC-07
title: Delivery Simulation
module: [delivery]
priority: P0
status: done
depends_on: [EPIC-06]
covers_requirements: [FR-DELIVERY-01]
related_adrs: [ADR-005, ADR-012]
scope:
  paths:
    - apps/api/internal/modules/delivery/**
    - apps/api/internal/contracts/delivery.go
    - apps/api/internal/platform/migrations/sql/delivery_001_create_delivery.sql
    - apps/api/internal/platform/scheduler/**
---

# EPIC-07: Delivery Simulation

## Цель

Timer-driven стейт-машина доставки (ADR-005): `assembling → courier_assigned → in_transit → [опционально delivery_delayed] → delivered`, устойчивая к рестарту сервиса, публикующая статусы через outbox для `notification` (EPIC-08).

## Задачи

1. `internal/platform/scheduler`: разделяемый примитив "запланировать вызов callback в момент `fire_at`, устойчиво к рестарту" — таблица `scheduled_transitions` (per-schema, каждый модуль-потребитель создаёт свою), `SELECT ... FOR UPDATE SKIP LOCKED` polling loop (тик 1с), интерфейс `scheduler.Schedule(ctx, tx, fireAt, payload) error` для регистрации в рамках существующей транзакции + `scheduler.Run(ctx, handler)` для запуска обработчика в `cmd/worker`. Оценить на этапе реализации: если EPIC-06 (mock payment webhook delay) уже реализовал похожий механизм — рефакторить в общий примитив здесь, не дублировать логику поллинга.
2. `domain/delivery.go`: агрегат `Delivery` (id, order_id, status, courier_name, courier_rating, started_at, estimated_completion_at). Метод `TransitionTo` с явной таблицей переходов, включая вероятностную ветку `delivery_delayed` (10% шанс, генерируется через `internal/platform/random`, инжектированный, не `math/rand` напрямую).
3. `usecase/StartDelivery`: вызывается consumer'ом `payment.succeeded.v1`, создаёт `Delivery` в статусе `assembling`, планирует переход в `courier_assigned` через scheduler (случайная задержка 10-30с), генерирует случайное имя курьера (faker) + рейтинг `Uniform(4.2, 5.0)`.
4. `usecase/AdvanceDeliveryState`: вызывается scheduler-обработчиком при срабатывании таймера — выполняет переход, публикует outbox-событие `delivery.status_changed.v1`, планирует следующий переход (с учётом 10% шанса `delivery_delayed`).
5. `usecase/CompleteDelivery`: финальный переход в `delivered`, публикует `delivery.completed.v1`.
6. `adapter/postgres`: схема `delivery`, таблицы `deliveries`, `scheduled_transitions`, `outbox_events`, `processed_events`.
7. `adapter/kafka`: consumer группа `delivery-service-group`, подписка на `payment.succeeded.v1`.
8. Детерминированность для тестов: seed для `delivery_id` прокидывается в `random.Source`, что позволяет юнит-тестам предсказывать длительности и наличие/отсутствие `delivery_delayed` без flaky-тестов.
9. Тесты: unit на `Delivery.TransitionTo` (все переходы, включая delayed-ветку), unit на распределение длительностей (property-based: N прогонов, все в заявленном диапазоне [10,30]/[60,180]/[15,45] секунд), интеграционный тест полного жизненного цикла с ускоренным временем (инжектированные короткие интервалы в тестовом конфиге, не реальные 60-180с в CI).

## Acceptance Criteria

- [x] Полный цикл `assembling(10-30s) → courier_assigned → in_transit(60-180s) → delivered` работает и переживает рестарт сервиса worker в середине цикла (`FR-DELIVERY-01`, `ADR-005`)
- [x] `delivery_delayed` ветка срабатывает примерно в 10% случаев на большой выборке (property-based тест допускает статистический разброс)
- [x] Курьер имеет сгенерированное имя и рейтинг при назначении (`FR-DELIVERY-01`)
- [x] Каждый переход публикует событие для `notification` модуля с задержкой доставки в Kafka < 500мс от коммита транзакции

## Definition of Done

Стандартный DoD. Обязательно включить тест "рестарт посередине доставки" — это ключевая архитектурная гарантия ADR-005, должна быть явно продемонстрирована, а не только заявлена в документации.
