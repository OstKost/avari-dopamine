---
id: EPIC-05
title: Order Lifecycle
module: [order]
priority: P0
status: todo
depends_on: [EPIC-04]
covers_requirements: [FR-ORDER-01, FR-ORDER-02, FR-HISTORY-01, INV-01, INV-02, INV-04, INV-05]
related_adrs: [ADR-003, ADR-004, ADR-011]
scope:
  paths:
    - apps/api/internal/modules/order/**
    - apps/api/internal/contracts/order.go
    - apps/api/migrations/order/**
    - docs/api/openapi.yaml   # секция /orders/*
---

# EPIC-05: Order Lifecycle

## Цель

Центральный агрегат системы. Создание заказа из корзины, стейт-машина статусов (`created → payment_pending → paid → assembling → courier_assigned → in_transit → delivered`, ветки `payment_failed`/`cancelled`), история заказов, публикация событий через outbox (ADR-003).

## Задачи

1. `domain/order.go`: агрегат `Order` (id, user_id, status, total_amount_rub — всегда 10.00, pickup_point_snapshot, created_at, items []OrderItem). Метод `TransitionTo(newStatus) error`, валидирующий допустимые переходы согласно явно закодированной таблице переходов (не набор `if`, а explicit map `allowedTransitions[currentStatus][]Status` — легко читается и тестируется).
2. `domain/order_item.go`: `OrderItem` (product_id, name_snapshot, price_snapshot, quantity) — снапшот на момент заказа (ADR-011).
3. `domain/errors.go`: `ErrCartEmpty`, `ErrInvalidStatusTransition`, `ErrDuplicatePendingOrder` (`INV-02`).
4. `port/`: `OrderRepository`, `OutboxWriter` (общий паттерн, может быть в `internal/platform/outbox`, реюзается всеми модулями, публикующими события — оценить вынос в platform при реализации, если больше одного модуля нуждается — да, payment/delivery тоже нуждаются, поэтому `internal/platform/outbox` реализуется здесь как разделяемый примитив с транзакционной записью, но таблица `outbox_events` — per-schema, см. ниже).
5. `usecase/CreateOrder`: читает `CartStore` (через `contracts.CartLookup`), `ProductLookup.GetByIDs` для снапшота цен, проверяет `INV-02` (нет другого `payment_pending` заказа у пользователя), в одной транзакции пишет `orders`+`order_items`+`outbox_events` (`order.created.v1`), очищает корзину (после commit — см. ADR-008).
6. `usecase/TransitionOrderStatus`: вызывается consumer'ами событий от `payment`/`delivery` модулей (через Kafka, не прямой вызов — ADR-001), либо HTTP (для admin/debug сценариев, если появятся). Каждый вызов — в транзакции с outbox-записью нового статуса для `notification`.
7. `usecase/ListOrderHistory`: курсорная пагинация по `(created_at, id)`, фильтр по статусу/дате (`FR-HISTORY-01`).
8. `usecase/RepeatOrder`: берёт состав старого заказа, создаёт новую корзину с тем же набором товаров (не создаёт заказ напрямую — пользователь всё равно проходит через checkout).
9. `adapter/postgres`: схема `order`, таблицы `orders`, `order_items`, `outbox_events`, `order_status_history` (append-only лог переходов для аудита — `FR-ORDER-02` "все переходы логируются с timestamp").
10. `adapter/kafka`: consumer группа `order-service-group`, подписка на `payment.succeeded.v1`, `payment.failed.v1`, `delivery.status_changed.v1`, `delivery.completed.v1` — каждый вызывает `TransitionOrderStatus` идемпотентно (ADR-003 `processed_events` таблица).
11. `adapter/httpapi`: `POST /orders` (checkout), `GET /orders` (история), `GET /orders/{id}`, `POST /orders/{id}/repeat`.
12. `internal/platform/outbox`: реализовать разделяемый relay-воркер (`cmd/worker` подкоманда), который умеет поллить `outbox_events` в ЛЮБОЙ схеме (параметризуется списком схем) — единый процесс обслуживает outbox всех модулей, чтобы не поднимать N воркеров на 100 MAU (соответствует ADR-002 экономии ресурсов).
13. Тесты: unit на `Order.TransitionTo` (все валидные и невалидные переходы — таблица тест-кейсов), usecase-тесты (в т.ч. `INV-02` конфликт), интеграционный тест полного flow create→transition через реальную Kafka+Postgres (docker-compose test profile), тест идемпотентности consumer (повторная доставка события не создаёт дублирующий эффект).

## Acceptance Criteria

- [ ] Пустая корзина не может быть оформлена (`FR-ORDER-01`)
- [ ] Сумма заказа всегда 10.00 RUB независимо от состава корзины (`INV-01`)
- [ ] Заказ создаётся в статусе `created`, событие `order.created.v1` публикуется через outbox в той же транзакции (`FR-ORDER-01`, `NFR-REL-01`)
- [ ] Невалидные переходы статуса отклоняются на уровне домена (`FR-ORDER-02`)
- [ ] Все переходы статуса записаны в `order_status_history` с timestamp (`FR-ORDER-02`)
- [ ] Пользователь не может иметь два `payment_pending` заказа одновременно (`INV-02`)
- [ ] Заказ неизменяем после перехода в `paid` (`INV-04`) — попытка изменить состав возвращает ошибку домена
- [ ] Повторная доставка Kafka-события не приводит к повторному переходу статуса (`NFR-REL-02`)
- [ ] История заказов пагинируется курсором, поддерживает фильтр по статусу/дате (`FR-HISTORY-01`)

## Definition of Done

Стандартный DoD. Дополнительно: диаграмма стейт-машины (Mermaid) в комментарии к `domain/order.go` или отдельном `docs/order-state-machine.md`, синхронизированная с фактической `allowedTransitions` картой.
