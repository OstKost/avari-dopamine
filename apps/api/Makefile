# Makefile для apps/api. Единая точка входа для команд разработки —
# и человека, и агента. НЕ дублируйте эти команды напрямую в CI-скриптах,
# ссылайтесь на таргеты отсюда, чтобы поведение локально и в CI совпадало.
#
# Требует Go >= 1.23.4 в PATH. golangci-lint устанавливается отдельно
# (см. `make lint-install`), если ещё не установлен.

GOLANGCI_LINT_VERSION := v1.62.2

.PHONY: run run-worker test test-integration lint lint-arch lint-install \
	migrate-up migrate-down migrate-status seed-catalog build fmt vet

## Запускает HTTP API процесс локально (требует DATABASE_URL, REDIS_ADDR,
## KAFKA_BROKERS, AUTH_JWT_SECRET в окружении — см. README.md "Переменные окружения").
run:
	go run ./cmd/server

## Запускает фоновый worker-процесс (outbox relay, delivery scheduler — по мере реализации Epics).
run-worker:
	go run ./cmd/worker

## Юнит-тесты всех пакетов (без реальных Postgres/Redis/Kafka — используются моки/skip).
test:
	go test ./... -race -count=1

## Интеграционные тесты, требующие поднятой инфраструктуры (docker compose up).
## Переменные окружения: DATABASE_URL, REDIS_ADDR.
test-integration:
	go test -tags=integration ./... -race -count=1

## Полный набор линтеров (errcheck, staticcheck, revive, govet, gofmt, depguard, ...).
lint:
	golangci-lint run ./...

## Только архитектурная проверка границ модулей (ADR-004) — быстрый сигнал
## отдельно от полного lint, удобно гонять после каждого изменения импортов.
## Примечание: golangci-lint CLI --enable добавляет к набору линтеров из
## .golangci.yml, а не заменяет его, поэтому здесь просто фильтруем полный
## вывод по имени линтера (depguard).
lint-arch:
	@echo "Running lint, filtering for depguard (architecture boundary) findings..."; \
	OUTPUT=$$(golangci-lint run ./... 2>&1 | grep -i depguard || true); \
	if [ -n "$$OUTPUT" ]; then \
		echo "$$OUTPUT"; \
		echo ""; \
		echo "ARCHITECTURE VIOLATION: cross-module import detected (ADR-004). Use internal/contracts/ instead."; \
		exit 1; \
	else \
		echo "OK: no internal/modules/** boundary violations found."; \
	fi

## Устанавливает golangci-lint нужной версии в $(go env GOPATH)/bin, если ещё не установлен.
lint-install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

## Применяет все "up" миграции к DATABASE_URL из окружения.
migrate-up:
	go run ./cmd/migrator up

## Откатывает последнюю миграцию.
migrate-down:
	go run ./cmd/migrator down

## Показывает статус применённых/неприменённых миграций.
migrate-status:
	go run ./cmd/migrator status

## Генерирует синтетический каталог товаров (реализуется в EPIC-02).
seed-catalog:
	go run ./cmd/seed catalog

## Собирает все бинарники в ./bin/ (проверка, что весь монорепозиторий компилируется).
build:
	mkdir -p bin
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker
	go build -o bin/seed ./cmd/seed
	go build -o bin/migrator ./cmd/migrator

fmt:
	gofmt -w .

vet:
	go vet ./...
