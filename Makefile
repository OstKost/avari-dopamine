# =====================================================================
# Avari Dopamine (Dopamine Market) — Root Makefile
# =====================================================================
# Единая точка входа для сборки, тестирования, линтинга,
# локальной разработки (dev) и запуска в продакшене (prod).
# =====================================================================

SHELL := /bin/bash
GOTOOLCHAIN ?= local
export GOTOOLCHAIN

# Загружаем переменные из .env, если файл существует
ifneq (,$(wildcard .env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif

.PHONY: help dev dev-api dev-worker dev-web dev-infra dev-infra-down \
	build build-api build-web prod prod-build prod-start \
	test test-api test-web test-integration \
	lint lint-api lint-web lint-arch \
	migrate-up migrate-down migrate-status seed-catalog \
	stop clean

# ---------------------------------------------------------------------
# Default Target: Help
# ---------------------------------------------------------------------
help: ## Показать справку по доступным командам
	@echo ""
	@echo "Avari Dopamine (Dopamine Market) — Команды управления проектом:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""

# ---------------------------------------------------------------------
# Development Targets
# ---------------------------------------------------------------------
dev: ## Запустить API сервер, Worker и Frontend одновременно в dev-режиме
	@echo "Запуск Avari Dopamine в dev-режиме (API + Worker + Web)..."
	@trap 'kill 0' SIGINT SIGTERM EXIT; \
	(cd apps/api && go run ./cmd/server) & \
	(cd apps/api && go run ./cmd/worker) & \
	(cd apps/web && pnpm dev) & \
	wait

dev-api: ## Запустить только Go HTTP API сервер (порт 8080)
	cd apps/api && go run ./cmd/server

dev-worker: ## Запустить только Go background worker (outbox relay / delivery)
	cd apps/api && go run ./cmd/worker

dev-web: ## Запустить только Next.js Frontend dev сервер (порт 3000)
	cd apps/web && pnpm dev

dev-infra: ## Поднять Docker инфраструктуру (Postgres, Redis, Kafka, Graylog)
	docker compose --env-file .env -f deploy/docker-compose.yml up -d

dev-infra-down: ## Остановить Docker инфраструктуру
	docker compose -f deploy/docker-compose.yml down

# ---------------------------------------------------------------------
# Build Targets
# ---------------------------------------------------------------------
build: build-api build-web ## Собрать все компоненты проекта (Backend бинарники + Frontend)

build-api: ## Собрать все Go бинарники (server, worker, seed, migrator) в apps/api/bin/
	@echo "==> Сборка Go бинарников..."
	cd apps/api && make build

build-web: ## Собрать оптимизированный production билд Next.js
	@echo "==> Сборка Next.js Frontend..."
	cd apps/web && pnpm build

# ---------------------------------------------------------------------
# Production Targets
# ---------------------------------------------------------------------
prod-build: build ## Подготовить production сборку бэкенда и фронтенда

prod-start: ## Запустить production сборку (API server, Worker и Next.js production)
	@echo "Запуск Avari Dopamine в production режиме..."
	@trap 'kill 0' SIGINT SIGTERM EXIT; \
	./apps/api/bin/server & \
	./apps/api/bin/worker & \
	(cd apps/web && pnpm start) & \
	wait

prod: prod-build prod-start ## Собрать и запустить проект в production режиме

# ---------------------------------------------------------------------
# Testing Targets
# ---------------------------------------------------------------------
test: test-api test-web ## Запустить все тесты (Go unit tests с race detector + Frontend typecheck)

test-api: ## Запустить unit тесты бэкенда с детектором race condition
	@echo "==> Запуск тестов Backend (Go)..."
	cd apps/api && make test

test-web: ## Запустить проверку типов TypeScript во фронтенде
	@echo "==> Проверка типов Frontend (TypeScript)..."
	cd apps/web && pnpm typecheck

test-integration: ## Запустить интеграционные тесты бэкенда (требует поднятую инфраструктуру)
	@echo "==> Запуск интеграционных тестов Backend..."
	cd apps/api && make test-integration

# ---------------------------------------------------------------------
# Linting & Quality Gates
# ---------------------------------------------------------------------
lint: lint-arch lint-web ## Запустить проверки линтеров (Arch depguard, Frontend ESLint, Go vet)
	@echo "==> Проверка кода Backend (go vet)..."
	cd apps/api && make vet
	@echo "==> Все проверки линтеров успешно пройдены!"

lint-api: ## Запустить линтинг Go бэкенда (vet и golangci-lint)
	@echo "==> Проверка кода Backend..."
	cd apps/api && make vet && make lint-arch

lint-arch: ## Проверить архитектурные инварианты границ модулей (ADR-004 / AGENTS.md)
	@echo "==> Проверка архитектурных границ (lint-arch)..."
	cd apps/api && make lint-arch

lint-web: ## Запустить ESLint для Next.js фронтенда
	@echo "==> Линтинг Frontend (ESLint)..."
	cd apps/web && pnpm lint

# ---------------------------------------------------------------------
# Database & Seed Targets
# ---------------------------------------------------------------------
migrate-up: ## Применить миграции базы данных Postgres
	cd apps/api && make migrate-up

migrate-down: ## Откатить последнюю миграцию базы данных
	cd apps/api && make migrate-down

migrate-status: ## Проверить статус миграций базы данных
	cd apps/api && make migrate-status

seed-catalog: ## Сгенерировать и наполнить каталог товаров (INV-01)
	cd apps/api && make seed-catalog

# ---------------------------------------------------------------------
# Cleanup & Utility
# ---------------------------------------------------------------------
stop: dev-infra-down ## Остановить инфраструктуру и освободить ресурсы

clean: ## Очистить сгенерированные бинарники и кэш сборки
	rm -rf apps/api/bin apps/web/.next
