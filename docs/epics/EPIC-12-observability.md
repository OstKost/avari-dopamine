---
id: EPIC-12
title: Observability
module: [platform]
priority: P1
status: todo
depends_on: [EPIC-00]
covers_requirements: [NFR-OBS-01, NFR-OBS-02]
related_adrs: [ADR-010]
scope:
  paths:
    - apps/api/internal/platform/logger/**
    - apps/api/internal/platform/tracing/**
    - apps/api/internal/platform/metrics/**
    - deploy/graylog/**
---

# EPIC-12: Observability

## Цель

Довести наблюдаемость (заложенную скелетом в EPIC-00) до полноты, описанной в ADR-010: GELF-экспорт логов, OpenTelemetry-трейсинг сквозь HTTP/Postgres/Kafka границы, Prometheus-метрики.

## Задачи

1. `internal/platform/logger/gelf_handler.go`: `slog.Handler`, отправляющий GELF UDP пакеты в Graylog, конфигурация host/port из env.
2. `internal/platform/tracing`: инициализация OTel SDK, `otelchi` middleware для HTTP, `otelsql`-обёртка для Postgres пула (из EPIC-00), кастомный span wrapper для Kafka producer/consumer с propagation `trace_id` через message headers (`traceparent`).
3. `internal/platform/metrics`: Prometheus registry, HTTP `/metrics` эндпоинт, стандартные Go runtime коллекторы + business-метрики счётчики/гистограммы, экспортируемые как публичный API для модулей (`metrics.OrdersCreatedTotal.Inc()` и т.п.) — модули регистрируют свои метрики через этот пакет, не создавая собственные registries (важно для единого `/metrics` эндпоинта).
4. Обновить все существующие usecase (из EPIC-01–08, после их реализации) добуавлением вызовов метрик, если не сделано инкрементально — этот Epic может идти параллельно и "догонять" через отдельные небольшие PR на каждый модуль, либо быть выполнен как часть каждого модуля инкрементально (рекомендуется последнее — распределить ответственность на модульные Epics, этот Epic фокусируется на платформенном фундаменте и финальном аудите полноты покрытия).
5. `deploy/graylog/`: preconfigured GELF input (через API-скрипт или Content Pack), базовые dashboard/search views (сохранённые поиски по `module`, `trace_id`).
6. Аудит: пройти по всем HTTP-хендлерам и Kafka-consumer'ам, подтвердить наличие лога входа/выхода и корректного `trace_id` propagation — составить чеклист в PR description.

## Acceptance Criteria

- [ ] Логи всех сервисов видны в Graylog UI с полями `module`, `trace_id`, `order_id` где применимо (`NFR-OBS-01`)
- [ ] Один `trace_id` позволяет увидеть полный путь заказа через order → payment → delivery → notification в трейсинг-инструменте (`NFR-OBS-02`)
- [ ] `/metrics` эндпоинт отдаёт бизнес-метрики, видимые Prometheus (или `curl` для проверки формата)
- [ ] `docker-compose up` включает предустановленный GELF input в Graylog "из коробки" (не требует ручной настройки после первого старта)

## Definition of Done

Стандартный DoD. Приложить в PR скриншот Graylog с отфильтрованным по `trace_id` полным путём одного тестового заказа — наглядная демонстрация ценности наблюдаемости.
