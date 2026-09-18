# ADR-010: Структурные логи → Graylog, трейсинг через OTel

- Статус: accepted
- Дата: 2026-09-18
- Контекст PRD: NFR-OBS-01, NFR-OBS-02

## Контекст

Graylog явно запрошен в стеке. Наблюдаемость особенно важна в event-driven системе (ADR-003) — без трейсинга сложно отследить путь заказа через order → payment → delivery → notification, особенно при дебаге агентами (AI-агент должен уметь диагностировать проблему по логам, а не только по коду — это часть заявленной компетенции пользователя "подробно диагностирует ошибки через логи").

## Решение

### Логирование

- Единый логгер — стандартный `log/slog` (Go 1.21+), формат JSON.
- Каждая запись обязательно включает: `timestamp`, `level`, `msg`, `service` (`api`/`worker`), `module` (identity/catalog/order/...), `trace_id`, `span_id`, и контекстные поля (`order_id`, `user_id` где применимо).
- Транспорт в Graylog: GELF UDP через `slog.Handler`-адаптер (`internal/platform/logger/gelf_handler.go`), направлено на Graylog GELF UDP input (порт 12201). UDP выбран для минимальной задержки/накладных расходов на critical path; при потере отдельных лог-сообщений при пиковой нагрузке — приемлемо (логи не бизнес-критичны, в отличие от событий Kafka).
- Локальная разработка: тот же JSON пишется в stdout человекочитаемо через `slog.TextHandler` при `ENV=local` (переключение по конфигу, не по коду).

### Трейсинг

- OpenTelemetry SDK для Go: авто-инструментация HTTP (chi middleware `otelchi`), Postgres (`otelsql`), Kafka producer/consumer (кастомный span wrapper вокруг `Initiate`/`Handle`).
- Экспорт трейсов — OTLP → в MVP локально в консоль/Jaeger-контейнер (опционально включаемый профиль в docker-compose, не обязательный для базового запуска); связь логов и трейсов через общий `trace_id`, инжектируемый в контекст запроса middleware и пробрасываемый через Kafka-заголовки сообщений (propagation через `traceparent` header в Kafka message headers).

### Метрики

- Prometheus `/metrics` эндпоинт на `apps/api` и `cmd/worker`: стандартные Go runtime метрики + бизнес-метрики (`orders_created_total`, `orders_completed_total`, `outbox_unpublished_count`, `payment_provider_errors_total{provider=}`, `delivery_active_count`).

## Альтернативы

| Вариант | Почему отклонён |
|---|---|
| ELK stack вместо Graylog | Graylog явно запрошен пользователем в стеке |
| Только stdout + `docker logs` без Graylog | Не демонстрирует навык централизованного логирования; хуже для дебага event-driven потоков |
| GELF TCP вместо UDP | TCP гарантирует доставку, но добавляет backpressure риск на critical path при проблемах с Graylog; для портфолио-масштаба UDP достаточен, задокументировано как осознанный трейд-офф |
| Полноценный APM (Datadog/New Relic) | Платный SaaS, не соответствует требованию self-hosted инфраструктуры на одной VPS (ADR-002) |

## Последствия

- (+) Единая точка расследования инцидентов — Graylog UI с фильтрами по `trace_id`/`order_id`/`module`.
- (+) `trace_id`, пробрасываемый через Kafka headers, позволяет восстановить полный путь заказа через все модули в одном трейсе — сильный демонстрационный элемент.
- (-) UDP GELF может терять логи под нагрузкой — задокументированный, принятый риск на данном масштабе.
- (-) Требует поддержания корректного propagation контекста через все async-границы (HTTP → outbox → Kafka → consumer) — источник потенциальных ошибок, покрывается интеграционными тестами Epic OBSERVABILITY.
