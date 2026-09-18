---
id: EPIC-00
title: Platform Bootstrap
module: [platform, infra]
priority: P0
status: todo
depends_on: []
covers_requirements: [NFR-MAINT-01, NFR-COST-01]
related_adrs: [ADR-001, ADR-002, ADR-004, ADR-011]
scope:
  paths:
    - apps/api/cmd/**
    - apps/api/internal/platform/**
    - apps/api/internal/contracts/**
    - apps/api/go.mod
    - apps/api/Makefile
    - apps/api/.golangci.yml
    - apps/api/migrations/**
    - deploy/**
    - .github/workflows/**
---

# EPIC-00: Platform Bootstrap

## Цель

Заложить фундамент монорепозитория: структуру Go-модуля, платформенные пакеты (config, db, kafka client, redis client, logger, http server, middleware), Docker Compose со всей инфраструктурой, скелет composition root и линтер, проверяющий архитектурные границы (ADR-004). Без этого Epic никакой другой backend-Epic не может начаться.

## Задачи

1. Инициализировать Go-модуль `apps/api` (`go mod init github.com/{org}/dopamine-market/api`), базовая структура директорий согласно ADR-001/ADR-004 (`cmd/server`, `cmd/worker`, `cmd/seed`, `cmd/migrator`, `internal/platform`, `internal/contracts`, `internal/modules` — пустой, модули создаются в своих Epics).
2. `internal/platform/config`: загрузка конфигурации из env (через `envconfig` или ручной парсинг), fail-fast валидация обязательных переменных при старте.
3. `internal/platform/db`: пул подключений к Postgres (`pgxpool`), health-check метод, обёртка транзакций (`WithTx(ctx, fn) error`) с поддержкой вложенных вызовов через `context`.
4. `internal/platform/kafka`: обёртка producer/consumer над `segmentio/kafka-go` или `confluent-kafka-go` — выбрать одну библиотеку и задокументировать выбор в комментарии (segmentio/kafka-go рекомендуется — pure Go, не требует cgo/librdkafka, проще для CI и Docker образов).
5. `internal/platform/redis`: клиент `go-redis/v9`, health-check.
6. `internal/platform/logger`: `slog`-обёртка с GELF-handler (ADR-010) + text-handler для локальной разработки, инжект `trace_id`/`module` полей из контекста.
7. `internal/platform/random`: детерминированный генератор с инъекцией seed (ADR-012), интерфейс `random.Source` с методами `Float64Range`, `IntRange`, `Pick[T](items []T) T`.
8. `internal/platform/httpserver`: обёртка над `chi`, стандартные middleware (recover-with-logging, request-id, CORS, rate-limit hook-point), graceful shutdown.
9. `internal/contracts/`: создать файлы-заглушки с итоговыми интерфейсами (payment.go, notification.go и т.д. — фактическое содержимое дополняется соответствующими Epics, но ЗАРЕЗЕРВИРОВАТЬ структуру пакета сейчас, чтобы депгард-правило можно было включить с первого дня).
10. `.golangci.yml`: настроить `depguard` согласно правилу ADR-004 (запрет импорта `internal/modules/**` откуда угодно, кроме `cmd/**` и самого модуля), плюс `errcheck`, `staticcheck`, `revive`, `gofmt`.
11. `Makefile` в `apps/api`: таргеты `run`, `test`, `lint`, `lint-arch` (отдельный таргет, оборачивающий depguard-проверку с понятным выводом), `migrate-up`, `migrate-down`, `seed`.
12. `deploy/docker-compose.yml`: сервисы `postgres`, `redis`, `kafka` (KRaft single-node), `graylog` + зависимости Graylog (opensearch, mongodb), сети, volumes для персистентности. Профиль `observability` опционально добавляет Jaeger.
13. `apps/api/migrations/000_init.sql` (goose): создание расширений (`pgcrypto`), базовой схемы `public`, если нужно.
14. Заготовка `openapi.yaml` в корне репозитория (`docs/api/openapi.yaml`) с базовой структурой (info, servers, components/securitySchemes) — модули дополняют своими путями.
15. `README.md` (уже создан) — проверить, что все команды в Quick Start актуальны после реализации Makefile/compose.

## Definition of Done

- `docker compose -f deploy/docker-compose.yml up -d` поднимает все сервисы без ошибок, все health-checks зелёные.
- `make run` в `apps/api` стартует HTTP-сервер, отвечающий `200 OK` на `/healthz` (проверяет Postgres, Redis, Kafka connectivity).
- `make lint-arch` — существующий пустой `internal/modules` не даёт false positive, но правило доказуемо работает: тестовый временный файл с нарушающим импортом должен провалить проверку (продемонстрировать в PR, затем убрать тестовый файл).
- `make test` проходит (даже если тестов пока минимум — smoke-тест на config-загрузку и db-подключение через testcontainers или docker-compose-based integration test).
- Логи сервера при старте видны в Graylog UI (`http://localhost:9000`) после ручной проверки.

## Открытые вопросы для реализующего агента

- Выбор Kafka Go-клиента (`segmentio/kafka-go` рекомендован, но допустимо обосновать альтернативу в PR description).
- Версия Graylog/OpenSearch — использовать последние стабильные образы на момент реализации, зафиксировать конкретные теги (не `latest`) в docker-compose для воспроизводимости.
