---
id: EPIC-08
title: Realtime Notifications (SSE)
module: [notification]
priority: P0
status: todo
depends_on: [EPIC-05, EPIC-07]
covers_requirements: [FR-NOTIF-01]
related_adrs: [ADR-009, ADR-003]
scope:
  paths:
    - apps/api/internal/modules/notification/**
    - apps/api/internal/contracts/notification.go
    - docs/api/openapi.yaml   # секция /orders/{id}/events
---

# EPIC-08: Realtime Notifications (SSE)

## Цель

SSE-эндпоинт, транслирующий изменения статуса заказа клиенту в реальном времени (ADR-009), с восстановлением текущего состояния при переподключении.

## Задачи

1. `usecase/StreamOrderEvents`: при подключении клиента — читает текущий статус заказа (через `contracts.OrderLookup`), отправляет `event: snapshot`, затем подписывается на канал обновлений конкретного `order_id` и транслирует `event: status_changed` по мере поступления.
2. Механизм доставки от Kafka-consumer к конкретному открытому SSE-соединению: consumer группа `notification-service-group`, подписка на `order.*.v1`/`delivery.status_changed.v1`/`delivery.completed.v1` → внутренний in-process pub/sub (`internal/platform/pubsub`, простой `map[orderID][]chan Event]` с мьютексом, или через Redis Pub/Sub если нужна поддержка нескольких инстансов `apps/api` — на 100 MAU достаточно in-process, задокументировать как решение с триггером пересмотра при horizontal scaling API).
3. `adapter/httpapi`: `GET /orders/{id}/events` — chi handler с `http.Flusher`, устанавливает `Content-Type: text/event-stream`, держит соединение, отправляет keep-alive комментарии каждые 15с (предотвращает timeout промежуточных proxy).
4. Авторизация: проверка, что `order.user_id == request.user_id` (нельзя подписаться на чужой заказ) — 403 иначе.
5. Graceful cleanup: при разрыве соединения (context cancellation) — отписка от internal pub-sub канала, освобождение ресурсов.
6. Тесты: интеграционный тест — открыть SSE-соединение, опубликовать Kafka-событие, проверить получение на клиенте в течение ожидаемого времени; тест авторизации (чужой заказ → 403); тест восстановления снапшота при новом подключении к уже продвинутому заказу.

## Acceptance Criteria

- [ ] Переподключение клиента получает текущий статус, а не только будущие события (`FR-NOTIF-01`)
- [ ] Задержка от публикации в Kafka до получения клиентом < 500мс (`FR-NOTIF-01`) — замерить в интеграционном тесте
- [ ] Нельзя подписаться на SSE чужого заказа (403)
- [ ] Соединение переживает отсутствие событий 60+ секунд (keep-alive работает, не рвётся по timeout)

## Definition of Done

Стандартный DoD. Явно задокументировать в коде/PR ограничение in-process pub-sub (single-instance) и путь миграции на Redis Pub/Sub при горизонтальном масштабировании `apps/api`.
