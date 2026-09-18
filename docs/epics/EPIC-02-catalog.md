---
id: EPIC-02
title: Catalog & Synthetic Data
module: [catalog]
priority: P0
status: todo
depends_on: [EPIC-00]
covers_requirements: [FR-CATALOG-01, FR-CATALOG-02]
related_adrs: [ADR-012, ADR-011]
scope:
  paths:
    - apps/api/internal/modules/catalog/**
    - apps/api/internal/contracts/catalog.go
    - apps/api/migrations/catalog/**
    - apps/api/cmd/seed/**
    - docs/api/openapi.yaml   # секция /catalog/*
---

# EPIC-02: Catalog & Synthetic Data

## Цель

Каталог из 150-300 синтетических товаров с категориями и full-text поиском, плюс детерминированный seed-скрипт (ADR-012). Модуль полностью read-heavy на MVP (нет UI управления товарами — админ-функциональность вне MVP, задокументирована как backlog в PRD).

## Задачи

1. `domain/product.go`: сущность `Product` (id, name, category_id, price_rub — decimal, description, image_seed, created_at), `domain/category.go`.
2. `port/`: `ProductRepository` (List с фильтром/пагинацией, Search, GetByID, GetByIDs — последнее нужно EPIC-04/05 для валидации корзины/заказа).
3. `usecase/`: `ListProducts`, `SearchProducts`, `ListCategories`.
4. `adapter/postgres`: схема `catalog`, таблицы `products`, `categories`, `tsvector`-колонка + GIN-индекс для полнотекстового поиска (`FR-CATALOG-02`, требование <200мс).
5. `adapter/httpapi`: `GET /catalog/products` (пагинация, фильтр по category_id, query param `q` для поиска), `GET /catalog/categories`, `GET /catalog/products/{id}`.
6. `cmd/seed/catalog.go`: детерминированная генерация (ADR-012) — фиксированный сид, 10 категорий с юмористическим уклоном, 150-300 товаров, идемпотентный запуск (повторный запуск не дублирует, `ON CONFLICT DO NOTHING` по детерминированному UUID, выведенному из индекса генерации).
7. Экспортировать через `internal/contracts/catalog.go` интерфейс `ProductLookup` с методом `GetByIDs(ctx, ids []uuid.UUID) (map[uuid.UUID]ProductSnapshot, error)` — используется модулем `order` для снапшота цен при оформлении заказа (ADR-011: денормализация вместо cross-schema FK).
8. Обновить `docs/api/openapi.yaml`.
9. Тесты: unit на usecase, интеграционный тест full-text поиска на реальных сид-данных (проверка производительности — простой benchmark, не строгий SLA-тест в CI, но зафиксировать baseline).

## Acceptance Criteria

- [ ] Каталог содержит 150-300 товаров, 8-12 категорий после `make seed` (`FR-CATALOG-01`)
- [ ] Повторный запуск seed не создаёт дублей (идемпотентность)
- [ ] Поиск по названию/описанию возвращает релевантные результаты за <200мс на полном каталоге (`FR-CATALOG-02`)
- [ ] Пустой поисковый запрос возвращает дефолтную постраничную выдачу (`FR-CATALOG-02`)
- [ ] `ProductLookup.GetByIDs` корректно обрабатывает частично несуществующие ID (возвращает только найденные, не ошибку)

## Definition of Done

Стандартный DoD из `AGENTS.md`, плюс: `make seed` документирован в README, снимок сгенерированных категорий приложен к PR description для ревью тона/юмора данных.
