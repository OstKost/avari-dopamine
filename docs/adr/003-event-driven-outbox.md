# ADR-003: Kafka + Outbox pattern для событий заказа

- Статус: accepted
- Дата: 2026-09-18
- Контекст PRD: FR-ORDER-02, NFR-REL-01, NFR-REL-02, INV-05

## Контекст

Жизненный цикл заказа пересекает несколько модулей (`order` → `payment` → `delivery` → `notification`), которые по ADR-001 не имеют права напрямую импортировать друг друга. Нужен надёжный асинхронный канал распространения фактов о смене состояния, который переживает рестарт сервиса и не теряет события при сбое между записью в БД и отправкой в шину (classic dual-write problem).

## Решение

**Transactional Outbox + Kafka:**

1. Каждый модуль, публикующий событие, в ОДНОЙ Postgres-транзакции со своим бизнес-изменением пишет строку в таблицу `{schema}.outbox_events` (event_id UUID, aggregate_type, aggregate_id, event_type, payload JSONB, created_at, published_at NULL).
2. Отдельный фоновый процесс `outbox-relay` (часть `cmd/worker`) поллит необработанные записи (`published_at IS NULL`) через `SELECT ... FOR UPDATE SKIP LOCKED`, публикует в Kafka, помечает `published_at`.
3. Kafka topics именуются `{domain}.{event}.v1`, например `order.created.v1`, `payment.succeeded.v1`, `delivery.status_changed.v1`. Ключ партиции — `aggregate_id` (order_id), что гарантирует упорядоченность событий одного заказа.
4. Каждый consumer — идемпотентен: хранит обработанные `event_id` в таблице `{schema}.processed_events` (unique constraint), обрабатывает событие в транзакции с insert в эту таблицу — при повторной доставке insert конфликтует и обработка пропускается (INV-05: exactly-once на уровне эффекта, at-least-once на уровне транспорта).
5. Consumer group на модуль (`payment-service-group`, `delivery-service-group`), что позволяет их независимо масштабировать (на бумаге, см. ADR-002).

## Альтернативы

| Вариант | Почему отклонён |
|---|---|
| Прямой dual-write (запись в БД + продюсер Kafka без outbox) | Классическая проблема потери события при падении между двумя операциями — недопустимо по NFR-REL-01 |
| Debezium CDC (Postgres logical replication → Kafka Connect) | Более "production-grade" паттерн, но избыточная инфраструктурная сложность (отдельный Kafka Connect кластер) для 100 MAU и портфолио — усложняет demo без пропорциональной образовательной ценности |
| Синхронные HTTP-вызовы между модулями вместо событий | Создаёт временную связанность (order ждёт ответа payment синхронно), хуже отражает event-driven архитектуру, которую хотим показать |
| RabbitMQ вместо Kafka | Kafka явно запрошен в стеке проекта; также Kafka лучше подходит для будущего replay/аудита событий заказа |

## Последствия

- (+) Гарантия at-least-once доставки без потери событий при падении сервиса между записью в БД и отправкой в шину.
- (+) Полный event log заказа хранится в Postgres (outbox) и в Kafka — пригоден для аудита/дебага/replay.
- (+) Явная демонстрация зрелого паттерна распределённых систем в портфолио.
- (-) Дополнительная задержка публикации = интервал поллинга outbox-relay (целимся в 200-500мс, приемлемо против NFR-OBS/UX требований).
- (-) Требует фонового процесса и мониторинга лага outbox (метрика `outbox_unpublished_count`).

## Схема Kafka-топиков (нормативная)

```yaml
topics:
  - name: order.created.v1
    producer: order module
    consumers: [payment module]
  - name: payment.succeeded.v1
    producer: payment module
    consumers: [order module, delivery module]
  - name: payment.failed.v1
    producer: payment module
    consumers: [order module]
  - name: delivery.status_changed.v1
    producer: delivery module
    consumers: [order module, notification module]
  - name: delivery.completed.v1
    producer: delivery module
    consumers: [order module, notification module]
partitions_per_topic: 3   # запас на будущее масштабирование, избыточно для 100 MAU, но нулевая цена сейчас
replication_factor: 1     # single-broker (ADR-002); зафиксировать переход на 3 при multi-broker
```
