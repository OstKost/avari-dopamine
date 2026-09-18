---
id: EPIC-01
title: Identity & Auth
module: [identity]
priority: P0
status: todo
depends_on: [EPIC-00]
covers_requirements: [FR-AUTH-01, FR-AUTH-02, NFR-SEC-01, NFR-SEC-02]
related_adrs: [ADR-007, ADR-004, ADR-011]
scope:
  paths:
    - apps/api/internal/modules/identity/**
    - apps/api/internal/contracts/identity.go
    - apps/api/migrations/identity/**
    - docs/api/openapi.yaml   # только секция /auth/*
---

# EPIC-01: Identity & Auth

## Цель

Регистрация, вход, выход, refresh-ротация, rate limiting — согласно ADR-007. Результат — переиспользуемый способ для других модулей узнать `user_id` из запроса (middleware) и получить `contracts.IdentityLookup` при необходимости (например, order может понадобиться email пользователя для будущих email-уведомлений — не в MVP, но контракт зарезервировать).

## Задачи

1. `domain/user.go`: сущность `User` (id, email, password_hash, created_at), value object `Email` с валидацией формата и нормализацией (lowercase).
2. `domain/errors.go`: `ErrEmailAlreadyExists`, `ErrInvalidCredentials`, `ErrWeakPassword`.
3. `port/`: интерфейсы `UserRepository`, `RefreshTokenStore` (Redis-backed), `PasswordHasher`.
4. `usecase/register.go`, `usecase/login.go`, `usecase/refresh.go`, `usecase/logout.go` — согласно FR-AUTH-01/02 и ADR-007 (ротация + reuse detection).
5. `adapter/postgres`: схема `identity` (миграция `apps/api/migrations/identity/001_create_users.sql`), `sqlc`-запросы, репозиторий.
6. `adapter/redis`: реализация `RefreshTokenStore` (family-based ротация, reuse detection — см. ADR-007).
7. `adapter/argon2`: `PasswordHasher` через `golang.org/x/crypto/argon2` (параметры: memory=64MB, iterations=3, parallelism=соответствует CPU — задокументировать выбранные параметры в коде).
8. `adapter/httpapi`: роуты `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`. Установка/очистка `httpOnly` cookies.
9. `internal/platform/httpserver` middleware `RequireAuth` — парсит и валидирует access JWT из cookie, кладёт `user_id` в контекст. Разместить в `platform`, так как используется всеми модулями (не нарушает ADR-004 — это горизонтальная инфраструктура, зависящая от `contracts.IdentityLookup`, а не от `modules/identity` напрямую).
10. `internal/platform/ratelimit`: token-bucket в Redis, применить middleware к `/auth/login`, `/auth/register` (NFR-SEC-02: 5/мин на IP).
11. Обновить `docs/api/openapi.yaml` секцией `/auth/*` с полными схемами request/response.
12. Тесты: unit на domain (валидация email/пароля), usecase с mock-портами (успешный/неуспешный регистр, login, refresh reuse detection — обязательный тест-кейс), интеграционный тест полного flow register→login→refresh→logout против реального Postgres+Redis (testcontainers или docker-compose test profile).

## Acceptance Criteria (из PRD)

- [ ] Email уникален, case-insensitive (`FR-AUTH-01`)
- [ ] Пароль >= 8 символов валидируется на сервере (`FR-AUTH-01`)
- [ ] После регистрации выдаётся рабочая сессия без доп. шага верификации (`FR-AUTH-01`)
- [ ] Access token TTL 15 минут, refresh TTL 30 дней (`FR-AUTH-02`)
- [ ] Refresh token ротируется при каждом использовании (`FR-AUTH-02`)
- [ ] Повторное использование инвалидированного refresh-токена инвалидирует всю family (`ADR-007`)
- [ ] Logout инвалидирует refresh-сессию в Redis (`FR-AUTH-02`)
- [ ] 6-я попытка логина за минуту с одного IP получает `429` (`NFR-SEC-02`)

## Definition of Done

Всё из `AGENTS.md` раздел 5, плюс: интеграционный тест reuse-detection зелёный, `openapi.yaml` содержит `/auth/*` с примерами ответов, включая `4xx` ошибки.
