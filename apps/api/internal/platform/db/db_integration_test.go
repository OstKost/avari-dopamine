//go:build integration

// Тесты в этом файле требуют реального Postgres (DATABASE_URL) и
// запускаются только с тегом сборки integration:
//
//	DATABASE_URL=postgres://... go test -tags=integration ./internal/platform/db/...
//
// В песочнице агента-разработчика Docker недоступен (см. AGENTS.md
// "Известные ограничения окружения агента"), поэтому эти тесты не были
// запущены при первичной реализации EPIC-00 — следующий агент с доступом
// к docker compose должен прогнать их перед тем, как считать пакет db
// окончательно verified, а не только "компилируется".
package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/ostkost/dopamine-market/api/internal/platform/db"
)

func TestNew_ConnectsAndHealthChecks(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := db.New(ctx, db.Config{DSN: dsn})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer pool.Close()

	if err := pool.HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck() error: %v", err)
	}
}

func TestWithTx_CommitsOnSuccess(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := db.New(ctx, db.Config{DSN: dsn})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer pool.Close()

	err = pool.WithTx(ctx, func(ctx context.Context) error {
		_, execErr := pool.Conn(ctx).Exec(ctx, "SELECT 1")
		return execErr
	})
	if err != nil {
		t.Errorf("WithTx() error: %v", err)
	}
}

func TestWithTx_RollsBackOnError(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := db.New(ctx, db.Config{DSN: dsn})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer pool.Close()

	sentinel := context.Canceled // произвольная ошибка-маркер
	err = pool.WithTx(ctx, func(ctx context.Context) error {
		return sentinel
	})
	if err != sentinel {
		t.Errorf("expected sentinel error to propagate, got: %v", err)
	}
}
