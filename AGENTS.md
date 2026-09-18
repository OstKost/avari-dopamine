# AGENTS.md — Конституция агентной разработки Dopamine Market

Этот файл — обязательные к исполнению правила для ЛЮБОГО AI-агента (Cursor, Claude Code, Codex, Copilot Workspace и др.), работающего в этом репозитории. Правила имеют приоритет над "общими хорошими практиками", если противоречат друг другу — следуй этому файлу.

## 0. Порядок чтения перед началом работы

1. `docs/prd/PRD.md` — что и зачем строим, машиночитаемые `FR-*`/`NFR-*`/`INV-*` требования.
2. `docs/adr/README.md` и релевантные ADR — почему архитектура именно такая. **Не предлагай альтернативы, уже отклонённые в ADR, без явного нового ADR.**
3. `docs/epics/README.md` и конкретный Epic-файл своей задачи — точная область работы, acceptance criteria, definition of done.
4. Этот файл целиком.

Если задача противоречит PRD/ADR — остановись и опиши противоречие в PR description, не изобретай решение самостоятельно.

## 1. Границы ответственности агента

- Работай **строго в пределах одного Epic/Issue** за раз. Не трогай файлы вне заявленной области (см. `scope.paths` в файле Epic) без явной необходимости, отражённой в acceptance criteria.
- Если для выполнения задачи необходимо изменить контракт (`internal/contracts/*.go`) или `docs/prd/PRD.md` — это ОТДЕЛЬНАЯ задача уровня ADR. Сначала предложи новый ADR (`docs/adr/0NN-title.md`, статус `proposed`), затем реализацию.
- Никогда не удаляй существующий ADR. Изменение решения = новый ADR со ссылкой `Supersedes: ADR-00X`, старый ADR получает статус `superseded`.

## 2. Архитектурные инварианты (must never violate)

Эти правила проверяются CI (`make lint-arch`) — нарушение = автоматический fail, не дожидайся ревью:

1. **Запрещён импорт `internal/modules/X/*` из `internal/modules/Y/*`** (кроме `X == Y`). Межмодульное взаимодействие — ТОЛЬКО через `internal/contracts/*` (синхронно) или Kafka-события через outbox (асинхронно). См. ADR-001, ADR-004.
2. **Домен (`domain/`) не импортирует ничего из `adapter/`, `port/`, инфраструктурных SDK** (никаких `database/sql`, `github.com/segmentio/kafka-go` и т.п. в `domain/`). Domain — чистый Go + стандартная библиотека + возможно `github.com/google/uuid`, `github.com/shopspring/decimal`.
3. **Любое изменение состояния заказа проходит через явный метод агрегата** (`Order.TransitionTo(status)`), который валидирует допустимость перехода согласно стейт-машине FR-ORDER-02. Прямое `UPDATE orders SET status = ...` мимо доменного метода запрещено везде, включая миграции с данными и скрипты.
4. **Любая запись, требующая публикации события, пишется в одной transaction с outbox-записью** (ADR-003). Если ты пишешь `INSERT`/`UPDATE` в таблицу агрегата, который публикует событие — outbox-запись обязана быть в том же `tx`.
5. **Цена заказа = 10.00 RUB всегда** (`INV-01`). Если видишь код, вычисляющий сумму платежа от состава корзины — это баг, а не фича; исправь.
6. **Никаких `time.Sleep`/горутин с in-memory таймером для доменной логики доставки** — используй `delivery.scheduled_transitions` + scheduler (ADR-005). In-memory таймер не переживает рестарт и запрещён для бизнес-состояния.
7. **Все новые Postgres-таблицы модуля — в схеме этого модуля**, не в `public` (ADR-011). Явный `CREATE SCHEMA IF NOT EXISTS {module}` в первой миграции модуля, если её ещё нет.

## 3. Кодстайл и структура

### Backend (Go)

- Go 1.23+. `gofmt`/`goimports` обязателен (CI провалит PR без этого — прогоняй перед коммитом).
- Линтер: `golangci-lint run` с конфигом `.golangci.yml` в корне `apps/api` — включает `depguard` (правило п.2.1), `errcheck`, `staticcheck`, `revive`.
- Ошибки — оборачивай с контекстом (`fmt.Errorf("creating order: %w", err)`), доменные ошибки — sentinel-переменные или типизированные структуры в `domain/errors.go` каждого модуля (например `ErrCartEmpty`, `ErrInvalidStatusTransition`).
- Логирование — только через `internal/platform/logger` (обёртка `slog`), никогда `fmt.Println`/`log.Println` в коде за пределами `cmd/*/main.go` на этапе bootstrap.
- Каждый usecase — метод структуры с явно перечисленными зависимостями в конструкторе (никакого service locator / глобальных переменных).
- SQL — только через `sqlc` (генерируемый типобезопасный код из `.sql` файлов в `internal/modules/{name}/adapter/postgres/queries/`). Ручной `db.Query` со строковой конкатенацией запрещён.
- Тесты: `_test.go` рядом с файлом. Domain-слой — 100% unit-test coverage ожидается (чистые функции, легко тестируются). Usecase — тесты с mock-портами (генерируются `go generate` через `mockgen`, интерфейсы уже описаны в `port/`).

### Frontend (Next.js)

- TypeScript strict mode обязателен (`"strict": true` в `tsconfig.json`, без `any` без явного комментария `// eslint-disable-next-line` и обоснования).
- Server Components по умолчанию; `"use client"` — только когда нужна интерактивность/браузерные API/hooks (см. ADR-009 за распределением по типам страниц).
- Стили — Tailwind CSS. Не создавай отдельные `.module.css` без причины.
- Типы API — генерируются из `openapi.yaml` (`pnpm generate:api-types`), не пиши вручную дублирующие интерфейсы структур ответа backend.
- Формы — `react-hook-form` + `zod` для валидации, схема валидации максимально близко отражает backend-валидацию (не обязана быть идентичной, но не должна быть более permissive).

## 4. Работа с Git и PR

- Одна ветка — один Epic/Issue: `feat/{epic-id}-{short-slug}`, например `feat/EPIC-04-order-checkout`.
- Commit message — Conventional Commits (`feat(order): add checkout usecase`, `fix(payment): handle webhook signature mismatch`).
- PR description ОБЯЗАТЕЛЬНО содержит: ссылку на Epic/Issue, список покрытых `FR-*`/`NFR-*` из PRD, чеклист acceptance criteria с отметками, какие ADR затронуты (если есть).
- Не смешивай в одном PR изменения из разных модулей, если это не explicit cross-cutting задача (например, "добавить новое событие в contracts + consumer в другом модуле" — это одна логическая задача, можно один PR, но опиши обе стороны в description).

## 5. Тестирование и Definition of Done

Задача считается выполненной агентом, только если:

1. `make lint` (backend) / `pnpm lint` (frontend) — без ошибок.
2. `make test` — все тесты зелёные, включая новые тесты на добавленную функциональность (нет покрытия — задача не завершена).
3. `make lint-arch` — архитектурные инварианты (раздел 2) не нарушены.
4. Все acceptance criteria из соответствующего Epic-файла отмечены выполненными с указанием, ГДЕ это проверяется (номер теста/файл).
5. Если задача затрагивает API-контракт — `openapi.yaml` обновлён и `pnpm generate:api-types` перегенерирован в PR.
6. Локально `docker compose up` + сценарий из Epic воспроизводится вручную (агент описывает шаги проверки в PR).

## 6. Наблюдаемость — обязательна с первого коммита фичи

Для любого нового usecase/HTTP-хендлера/Kafka-consumer:

- Добавь структурный лог на вход и выход операции с ключевыми полями (`order_id`, `user_id`, длительность).
- Добавь минимум одну бизнес-метрику Prometheus, если операция меняет состояние агрегата (см. ADR-010 за списком существующих метрик — расширяй, не дублируй по смыслу).
- Пробрось `trace_id` через контекст — не создавай новый корневой span, если он уже есть в контексте запроса.

## 7. Работа с синтетическими данными (ADR-012)

- Не добавляй вызовы к внешним товарным/логистическим API — каталог и курьеры принципиально синтетические.
- Любая генерация случайных данных для домена (расстояние ПВЗ, длительность доставки, имя курьера) — использует `internal/platform/random` с возможностью инъекции seed для тестов. Не вызывай `math/rand` напрямую в usecase/domain коде.

## 8. Когда остановиться и спросить

Останавливайся и явно проси уточнение (в PR description как "Open question", не блокируя остальную работу, если возможно) когда:

- Задача требует изменить `INV-*` инвариант из PRD.
- Задача противоречит существующему `accepted` ADR и ты не уверен, что новый ADR оправдан.
- Acceptance criteria в Epic-файле неполны/противоречивы для однозначной реализации.

## 9. Структура Epic-файлов (для агентов, разбирающих `docs/epics/`)

Каждый файл `docs/epics/EPIC-NN-*.md` содержит YAML front-matter с `id`, `module`, `depends_on`, `covers_requirements` (список `FR-*`/`NFR-*`) и `scope.paths` (список директорий, которые Epic разрешено менять). Работай только внутри `scope.paths`, если явно не указано иное.

## 10. Статус реализации Epics (обновляется каждым агентом после DoD)

Этот раздел — единственный источник правды о том, что реально реализовано,
в отличие от того, что только специфицировано в `docs/epics/`. Каждый агент,
закрывающий Epic, ОБЯЗАН добавить/обновить свою строку здесь перед тем как
считать задачу выполненной.

| Epic | Статус | Комментарий |
|---|---|---|
| EPIC-00 Platform Bootstrap | ✅ done | См. "EPIC-00: заметки о реализации" ниже — есть задокументированные отклонения от исходной спеки. |
| EPIC-01 Identity & Auth | ✅ done | Регистрация, логин, refresh family ротация, reuse detection, rate limiter, RequireAuth middleware, unit тесты. |
| EPIC-02 Catalog & Synthetic Data | ✅ done | Категории, товары с ценой 10₽ (INV-01), полнотекстовый поиск Postgres FTS, детерминированный сид 200 товаров. |
| EPIC-03 Pickup Points | ✅ done | Сферическая геодезия DestinationPoint, генерация 5-8 точек в [100, 500]м (INV-03), fallback Ростов-на-Дону, property-based тесты. |
| EPIC-04 Cart | ✅ done | Redis-backed корзина (TTL 7 дней), иммутабельный Cart VO, обогащение из ProductLookup и PickupLookup, контракт CartLookup. |
| EPIC-05 Order Lifecycle | ✅ done | Создание заказа, стейт-машина переходов, transactional outbox relay, processed_events идемпотентность, INV-01/INV-02. |
| EPIC-09 Frontend Shell | ✅ done | Next.js 15, App Router, Tailwind CSS дизайн-система, генерация типов OpenAPI, Auth layout/middleware, SSR skeleton loaders. |
| EPIC-13 CI/CD & Quality Gates | ✅ done | GitHub Actions workflows (Backend CI, Frontend CI, Arch Lint Gate, Dependabot, Smoke Load Test). |
| EPIC-06..EPIC-14 | ⬜ not started | См. `docs/epics/EPIC-NN-*.md` |

### EPIC-00: заметки о реализации

Реализовано: `apps/api` (config, logger+GELF, db, redis, kafka wrappers,
httpserver+chi+/healthz, contracts placeholders, cmd/{server,worker,seed,migrator},
migrations, `.golangci.yml` с depguard-границами модулей, `Makefile`), `deploy/docker-compose.yml`,
`docs/api/openapi.yaml` (скелет), `.env.example`.

Проверено в песочнице агента: `go build ./...`, `go vet ./...`, `go test ./... -race`,
`golangci-lint run ./...` (0 issues), `make lint-arch` (демонстрация — ловит
искусственно созданное нарушение границы `internal/modules/order` →
`internal/modules/payment`, затем демо-файлы удалены).

**НЕ проверено в песочнице агента** (нет Docker): `docker compose up` реально
не запускался. `deploy/docker-compose.yml` статически валиден (`docker compose
config` не прогонялся, но YAML синтаксически корректен и структура проверена
вручную против ADR-002/ADR-010/ADR-011). Definition of Done "`make run` →
`/healthz` возвращает 200" — тоже не проверено сквозно, так как для этого
нужны реальные Postgres/Redis. **Следующий агент/человек с доступом к Docker
должен:** `docker compose --env-file .env -f deploy/docker-compose.yml up -d`,
подождать все health checks, затем `cd apps/api && make migrate-up && make
run` и убедиться, что `curl localhost:8080/healthz` возвращает `200` с
`{"status":"ok",...}`. Обновить эту заметку результатом.

**Задокументированные отклонения от исходного EPIC-00:**

1. **Go в песочнице.** Установлен вручную в `~/tools/go` (нет root-доступа),
   а не через системный пакетный менеджер. Это специфично для песочницы
   агента — на реальном сервере/CI ставить стандартным способом.

2. **Пины версий зависимостей ниже `@latest` — КРИТИЧНО, читать перед `go get`.**
   `go.mod` фиксирует `go 1.23.4` и точные версии: `chi v5.1.0`, `pgx/v5
   v5.7.6`, `goose/v3 v3.24.1`, `go-redis/v9 v9.11.0`, `kafka-go v0.4.51`.
   Причина: неосторожный `go get @latest`/`go mod tidy` заставляет Go
   автоматически скачать новый toolchain (наблюдались auto-downloads до
   1.25, 1.26, 1.26.8) и поднять `go` directive в `go.mod`, что тянет версии
   pgx/goose, требующие Go 1.25+/1.26+ — несовместимо с установленным
   1.23.4 и с `golangci-lint v1.62.2` (собран под Go 1.23, отказывается
   работать с конфигом, целящимся в Go 1.26). **Правило для всех будущих
   агентов: перед ЛЮБЫМ `go get`/`go mod tidy`/`go build` в этом проекте
   выполняй `export GOTOOLCHAIN=local`**, и при обновлении зависимости
   явно проверяй её минимальную требуемую версию Go против пина в `go.mod`.

3. **Расположение SQL-миграций.** Физически лежат в
   `internal/platform/migrations/sql/*.sql`, а не в `apps/api/migrations/`,
   как было в исходном тексте EPIC-00. Причина: директива `//go:embed` не
   поддерживает `../` в паттерне, а `cmd/migrator` встраивает миграции через
   `embed.FS` из пакета `internal/platform/migrations`. Схема именования
   файлов — префикс модуля в имени файла (`{module}_NNN_*.sql`) в одной
   папке `sql/`, а не директория-на-модуль, что соответствует логике
   ADR-011 (отдельная схема Postgres на модуль) при физическом ограничении
   go:embed.

4. **`make lint-arch` не использует `--disable-all --enable=depguard`.**
   Эмпирически подтверждено: golangci-lint CLI-флаги `--enable`
   *добавляют* линтер к набору из `.golangci.yml`, а не заменяют набор,
   если конфиг уже содержит `disable-all: true` с явным `enable:` списком.
   Решение — `lint-arch` гоняет полный `golangci-lint run ./...` и
   фильтрует вывод по подстроке `depguard`, что даёт тот же практический
   результат (сигнал только о нарушениях границ модулей) без борьбы с
   приоритетом конфигурации.

5. **`issues.exclude-use-default: false`** установлен явно в `.golangci.yml`.
   Дефолтные excludes golangci-lint v1.62 подавляли часть реальных находок
   (errcheck на `defer x.Close()`, revive package-comments) — с явным
   `false` линтер показывает их все, и они были исправлены (см. коммиты в
   этом Epic), а не оставлены скрытыми до случайного обнаружения позже.
