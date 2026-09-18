---
id: EPIC-13
title: CI/CD & Quality Gates
module: [infra]
priority: P1
status: done
depends_on: [EPIC-00]
covers_requirements: [NFR-MAINT-01, NFR-PERF-01]
related_adrs: [ADR-001, ADR-002, ADR-004]
scope:
  paths:
    - .github/workflows/**
    - apps/api/Makefile
    - apps/web/package.json
---

# EPIC-13: CI/CD & Quality Gates

## Цель

GitHub Actions пайплайны, гарантирующие, что каждый PR (в том числе от AI-агентов) проходит lint/test/arch-check до мержа, плюс базовый нагрузочный smoke-тест против NFR-PERF-01.

## Задачи

1. `.github/workflows/backend-ci.yml`: на каждый PR, затрагивающий `apps/api/**` — `go build`, `golangci-lint run` (включая depguard из EPIC-00), `go test ./... -race -cover`, поднятие docker-compose test-профиля для интеграционных тестов (Postgres+Redis+Kafka), покрытие ≥ порога (зафиксировать целевой % при реализации, например 70% для usecase/domain слоёв).
2. `.github/workflows/frontend-ci.yml`: на каждый PR, затрагивающий `apps/web/**` — `pnpm install --frozen-lockfile`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build` (проверка, что production build не ломается).
3. `.github/workflows/arch-lint.yml`: отдельный быстрый job, запускающий `make lint-arch` независимо от остальных тестов — быстрый fail для нарушения границ модулей (важно для agentic development — агент должен получить фидбек за секунды, не за минуты полного test suite).
4. Опциональный `.github/workflows/load-test.yml` (может быть `workflow_dispatch`, не на каждый PR): простой k6/vegeta скрипт, гоняющий 20 RPS на `/catalog/products` и `/orders`, проверяющий p95 < 200-300мс (`NFR-PERF-01`) — не блокирующий CI по умолчанию (нагрузочные тесты нестабильны на shared GitHub runners), но доступный для ручного запуска и локального использования.
5. Branch protection настройки (задокументировать в `docs/epics/EPIC-13-ci-cd.md` рекомендации, реальное включение — на стороне владельца репозитория, не агента): require passing checks, require review (если применимо).
6. `dependabot.yml` для автоматических security-обновлений зависимостей Go/npm.

## Acceptance Criteria

- [x] PR с нарушением архитектурных границ (тестовый пример) падает в `arch-lint` job за <1 минуту
- [x] PR с проваленным тестом не может быть смержен (статус check блокирует)
- [x] Frontend build проходит на чистом `pnpm install` без кэша
- [x] Нагрузочный тест воспроизводим локально одной командой

## Definition of Done

Стандартный DoD. Продемонстрировать (скриншотом или логом в PR) хотя бы один "красный" прогон каждого gate (искусственно введённая ошибка), подтверждающий, что gate действительно ловит проблему, а не просто существует формально.
