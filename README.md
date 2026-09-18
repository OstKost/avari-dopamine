# Dopamine Market

> Портфолио-проект: имитация онлайн-маркетплейса/доставки еды для получения «лёгкого дофамина» без реальных трат — фиксированная плата 10₽ за заказ, полностью синтетический каталог, симуляция доставки через таймер-стейт-машину.

Проект демонстрирует владение полным циклом продуктовой и инженерной работы: PRD → ADR → agentic development через `AGENTS.md` → продакшн-grade монолит на Go + Next.js SSR фронтенд с событийной архитектурой на Kafka.

## Зачем это существует

Продуктовая гипотеза: части пользователей не нужен реальный шопинг — им нужен **ритуал** (поиск → корзина → оплата → ожидание → получение), дающий кратковременный дофаминовый отклик, но без риска потратить много денег и без реальной логистики. Монетизация — фиксированные 10₽ за оформленный заказ, независимо от состава корзины.

Как инженерный артефакт проект показывает:
- Product thinking → машиночитаемый PRD → декомпозиция на Epics/Issues, готовые для агентной разработки.
- Ведение архитектурных решений через ADR (Architecture Decision Records).
- Чистую доменную архитектуру на Go (модульный монолит, порты/адаптеры).
- Event-driven связку через Kafka с outbox pattern и идемпотентностью.
- Next.js 15 App Router SSR фронтенд с потоковой доставкой статусов (SSE/WS).
- Полный DevOps-контур: Docker Compose, миграции, структурные логи в Graylog (GELF), CI.

## Технологический стек

| Слой | Технология |
|---|---|
| Frontend | Next.js 15 (App Router, SSR/RSC), TypeScript, TanStack Query |
| Backend | Go 1.23+, chi router, sqlc, goose (миграции) |
| Данные | PostgreSQL 16 (source of truth), Redis 7 (корзина, кэш, rate-limit) |
| Шина событий | Apache Kafka (KRaft), Outbox pattern |
| Наблюдаемость | Graylog (GELF), structured logging (`slog`), Prometheus + OpenTelemetry traces |
| Инфраструктура | Docker Compose, монорепозиторий, GitHub Actions CI |

## Структура монорепозитория

```
dopamine-market/
├── AGENTS.md                  # Инструкции для AI-агентов разработки
├── docs/
│   ├── prd/                   # Product Requirements Document (машиночитаемый)
│   ├── adr/                   # Architecture Decision Records
│   └── epics/                 # Декомпозиция на Epics/Stories для агентов
├── apps/
│   ├── web/                   # Next.js SSR фронтенд
│   └── api/                   # Go backend (модульный монолит)
│       ├── cmd/               # Точки входа (server, worker, migrator)
│       ├── internal/
│       │   ├── modules/       # Доменные модули (identity, catalog, cart, order, payment, delivery, pickup, notification)
│       │   ├── platform/      # Общая инфраструктура (db, kafka, redis, logger, http, config)
│       │   │   └── migrations/sql/  # SQL-миграции (goose), все в одной папке из-за ограничения go:embed (см. AGENTS.md → “Отклонения от исходных Epics”)
│       │   └── contracts/     # Публичные Go-интерфейсы модулей (для развязки)
│       ├── Makefile           # run/test/lint/lint-arch/migrate-*/seed-catalog/build
│       └── .golangci.yml      # вкл. архитектурный depguard (ADR-004)
├── deploy/
│   └── docker-compose.yml     # postgres, redis, kafka (KRaft), graylog+opensearch+mongodb, [observability] jaeger
└── .github/workflows/         # CI (планируется, ещё не реализовано)
```

## Быстрый старт

См. [`docs/prd/PRD.md`](docs/prd/PRD.md) для продуктовых требований, [`AGENTS.md`](AGENTS.md) для правил разработки и [`docs/epics/`](docs/epics/) для очереди задач.

```bash
# 1. Секреты для docker-compose (пароли Postgres/Redis/Graylog)
cp .env.example .env && $EDITOR .env

# 2. Инфраструктура: postgres, redis, kafka (KRaft), graylog+opensearch+mongodb
docker compose --env-file .env -f deploy/docker-compose.yml up -d
docker compose -f deploy/docker-compose.yml ps   # все сервисы healthy?

# 3. Backend: миграции и запуск API (требует DATABASE_URL/REDIS_ADDR/... в окружении — см. .env.example)
cd apps/api
export $(grep -v '^#' ../../.env | xargs)   # или используйте direnv/dotenv-загрузчик по вкусу
make migrate-up
make run              # → GET http://localhost:8080/healthz должен вернуть 200

# 4. Frontend (появится начиная с EPIC-09, пока не реализован)
cd apps/web && pnpm install && pnpm dev
```

Полезные команды разработки (вызываются из `apps/api/`, см. `Makefile`):

| Цель | Назначение |
|---|---|
| `make run` / `make run-worker` | запуск API / worker локально |
| `make test` | юнит-тесты (`-race`) |
| `make test-integration` | интеграционные тесты (требуют запущенную инфраструктуру) |
| `make lint` | полный `golangci-lint run ./...` |
| `make lint-arch` | только архитектурная проверка границ модулей (ADR-004) |
| `make migrate-up` / `migrate-down` / `migrate-status` | управление миграциями (goose) |
| `make seed-catalog` | генерация синтетического каталога (EPIC-02, пока заглушка) |
| `make build` | собрать все бинарники в `bin/` |

## Документация

- [PRD](docs/prd/PRD.md) — машиночитаемые продуктовые требования
- [ADR Index](docs/adr/README.md) — все архитектурные решения
- [Epics](docs/epics/README.md) — разбиение проекта для агентной разработки
- [AGENTS.md](AGENTS.md) — конституция агентной разработки
