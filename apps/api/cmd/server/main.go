// cmd/server — HTTP API процесс. Единственная точка входа, где все модули
// инстанцируются и связываются друг с другом (composition root, ADR-004).
//
// На этапе EPIC-00 модули (internal/modules/*) ещё не существуют — main
// поднимает только платформенный фундамент (config, logger, db, redis,
// kafka producer заготовка, HTTP-сервер с /healthz). Каждый следующий Epic
// (EPIC-01..EPIC-08) добавляет сюда инстанцирование своего модуля и
// монтирование его роутов через srv.Router().Mount(...) — см. TODO-маркеры
// ниже с указанием, куда что добавлять.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ostkost/dopamine-market/api/internal/platform/config"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
	"github.com/ostkost/dopamine-market/api/internal/platform/httpserver"
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
		Service:     "api",
		GraylogHost: cfg.Logger.GraylogHost,
		GraylogPort: cfg.Logger.GraylogPort,
	})
	if err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}
	defer loggerCleanup()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("starting dopamine-market api", slog.String("env", cfg.Env))

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
	log.Info("connected to postgres")

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
	log.Info("connected to redis")

	// TODO(EPIC-01 identity): инстанцировать identity-модуль здесь,
	//   инжектировать dbPool/redisClient/cfg.Auth, получить
	//   contracts.IdentityLookup для последующих модулей и chi.Router
	//   для монтирования на "/auth".
	// TODO(EPIC-02 catalog): аналогично для catalog -> "/catalog".
	// TODO(EPIC-03 pickup): аналогично для pickup -> "/pickup".
	// TODO(EPIC-04 cart): аналогично для cart -> "/cart", зависит от
	//   catalog.ProductLookup и pickup.PickupPointLookup контрактов.
	// TODO(EPIC-05 order): аналогично для order -> "/orders", зависит от
	//   cart.CartLookup и catalog.ProductLookup контрактов.
	// TODO(EPIC-06 payment): PAYMENT_PROVIDER switch (mock|yookassa) здесь,
	//   согласно ADR-006 — выбор адаптера конфигурацией, не кодом.
	// TODO(EPIC-07 delivery), TODO(EPIC-08 notification): см. соответствующие
	//   Epic-файлы docs/epics/ для деталей.

	srv := httpserver.New(httpserver.Options{
		Port:            cfg.HTTP.Port,
		ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
		Logger:          log,
		HealthCheckers: map[string]httpserver.HealthChecker{
			"postgres": dbPool,
			"redis":    redisClient,
		},
	})

	log.Info("http server listening", slog.Int("port", cfg.HTTP.Port))

	if err := srv.Run(ctx, cfg.HTTP.ShutdownTimeout); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("running http server: %w", err)
	}

	log.Info("shutdown complete")
	return nil
}
