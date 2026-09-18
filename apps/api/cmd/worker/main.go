// cmd/worker — фоновый процесс: outbox-relay (ADR-003), delivery/payment
// scheduler (ADR-005, ADR-006), consumer'ы Kafka-событий каждого модуля.
//
// Отдельный от cmd/server бинарник, потому что HTTP API и фоновая
// обработка событий имеют разные профили нагрузки и жизненного цикла —
// но оба используют один и тот же образ Docker (см. deploy/docker-compose.yml,
// разные команды запуска одного бинарника через go build ./cmd/...).
//
// На этапе EPIC-00 поднимает только платформенный фундамент и заглушку
// главного цикла. Каждый следующий Epic добавляет сюда свой outbox-relay
// подписку/scheduler запуск через errgroup — см. TODO-маркеры.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ostkost/dopamine-market/api/internal/platform/config"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
	"github.com/ostkost/dopamine-market/api/internal/platform/logger"
	"github.com/ostkost/dopamine-market/api/internal/platform/redis"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	log, loggerCleanup, err := logger.New(logger.Options{
		Env:         cfg.Env,
		Level:       cfg.Logger.Level,
		Service:     "worker",
		GraylogHost: cfg.Logger.GraylogHost,
		GraylogPort: cfg.Logger.GraylogPort,
	})
	if err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}
	defer loggerCleanup()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("starting dopamine-market worker", slog.String("env", cfg.Env))

	dbPool, err := db.New(ctx, db.Config{
		DSN:             cfg.DB.DSN,
		MaxOpenConns:    int32(cfg.DB.MaxOpenConns),
		MaxIdleConns:    int32(cfg.DB.MaxIdleConns),
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
	})
	if err != nil {
		return fmt.Errorf("connecting to postgres: %w", err)
	}
	defer dbPool.Close()

	redisClient, err := redis.New(ctx, redis.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		return fmt.Errorf("connecting to redis: %w", err)
	}
	defer func() {
		if closeErr := redisClient.Close(); closeErr != nil {
			log.Error("closing redis client", slog.String("error", closeErr.Error()))
		}
	}()

	// TODO(EPIC-05 order + platform/outbox): запустить outbox-relay для
	//   схем order/payment/delivery — единый relay-процесс, параметризуемый
	//   списком схем (ADR-003 "единый процесс обслуживает outbox всех
	//   модулей, чтобы не поднимать N воркеров на 100 MAU").
	// TODO(EPIC-07 delivery + platform/scheduler): запустить scheduler poll
	//   loop для scheduled_transitions (ADR-005).
	// TODO(EPIC-05/06/07/08): запустить Kafka consumer'ы каждого модуля
	//   параллельно через errgroup.Group, с graceful shutdown по ctx.

	log.Info("worker bootstrap complete, no background jobs registered yet (EPIC-00 baseline)")

	<-ctx.Done()
	log.Info("worker shutdown complete")
	return nil
}
