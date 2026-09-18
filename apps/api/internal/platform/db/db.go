// Package db предоставляет общий пул подключений к PostgreSQL и обёртку
// транзакций, используемую всеми модулями для реализации transactional
// outbox (ADR-003): бизнес-изменение и запись в outbox_events всегда
// выполняются в одной транзакции через WithTx.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ctxKey string

const ctxKeyTx ctxKey = "db_tx"

// Pool оборачивает *pgxpool.Pool, добавляя WithTx и health-check.
// Модули не создают собственные пулы — Pool инжектируется через
// composition root (cmd/server/main.go, cmd/worker/main.go).
type Pool struct {
	pool *pgxpool.Pool
}

// Config — параметры пула, транслируемые из internal/platform/config.
type Config struct {
	DSN             string
	MaxOpenConns    int32
	MaxIdleConns    int32
	ConnMaxLifetime time.Duration
}

// New открывает пул подключений и сразу проверяет связность (fail-fast —
// AGENTS.md/ADR-002 предпочитают явную ошибку при старте над скрытой
// деградацией в runtime).
func New(ctx context.Context, cfg Config) (*Pool, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parsing postgres DSN: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		pgxCfg.MaxConns = cfg.MaxOpenConns
	}
	if cfg.MaxIdleConns > 0 {
		pgxCfg.MinConns = cfg.MaxIdleConns
	}
	if cfg.ConnMaxLifetime > 0 {
		pgxCfg.MaxConnLifetime = cfg.ConnMaxLifetime
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("creating postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}

	return &Pool{pool: pool}, nil
}

// Close освобождает пул. Вызывается один раз при graceful shutdown процесса.
func (p *Pool) Close() {
	p.pool.Close()
}

// HealthCheck используется /healthz эндпоинтом (см. httpserver) для проверки
// живости зависимости перед тем, как объявить сервис готовым принимать трафик.
func (p *Pool) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return p.pool.Ping(ctx)
}

// Querier — минимальный интерфейс, который реализуют и *pgxpool.Pool,
// и pgx.Tx. Репозитории модулей типизируют свои sqlc Queries через этот
// интерфейс, чтобы один и тот же repository-код работал как внутри
// транзакции WithTx, так и вне неё (обычный read-запрос без транзакции).
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Conn возвращает Querier для выполнения запроса: если context уже несёт
// открытую транзакцию (был начат WithTx выше по стеку вызовов), возвращает
// её — так репозитории транспарентно участвуют в транзакции вызывающего
// usecase без явной передачи tx через все слои. Если транзакции нет —
// возвращает пул напрямую (обычный, не транзакционный запрос).
func (p *Pool) Conn(ctx context.Context) Querier {
	if tx, ok := ctx.Value(ctxKeyTx).(pgx.Tx); ok {
		return tx
	}
	return p.pool
}

// WithTx выполняет fn в рамках одной Postgres-транзакции. При успешном
// завершении fn (nil error) — commit; при ошибке или панике — rollback,
// панику пробрасывает дальше после rollback.
//
// Критично для ADR-003 (outbox pattern): любой usecase, публикующий
// событие, обязан писать бизнес-изменение и outbox_events запись внутри
// одного WithTx вызова.
func (p *Pool) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	txCtx := context.WithValue(ctx, ctxKeyTx, tx)

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rolling back after error %q: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}
