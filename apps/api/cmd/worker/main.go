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
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ostkost/dopamine-market/api/internal/modules/cart"
	"github.com/ostkost/dopamine-market/api/internal/modules/catalog"
	"github.com/ostkost/dopamine-market/api/internal/modules/order"
	"github.com/ostkost/dopamine-market/api/internal/modules/pickup"
	"github.com/ostkost/dopamine-market/api/internal/platform/config"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
	"github.com/ostkost/dopamine-market/api/internal/platform/kafka"
	"github.com/ostkost/dopamine-market/api/internal/platform/logger"
	"github.com/ostkost/dopamine-market/api/internal/platform/outbox"
	"github.com/ostkost/dopamine-market/api/internal/platform/random"
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

	// Инициализация модулей для воркера
	rnd := random.New(time.Now().UnixNano())
	catalogModule := catalog.NewModule(dbPool.Raw())
	pickupModule := pickup.NewModule(dbPool.Raw(), rnd)
	cartModule := cart.NewModule(redisClient.Raw(), catalogModule, pickupModule, 7*24*time.Hour)
	orderModule := order.NewModule(dbPool, cartModule, catalogModule, pickupModule)

	// Outbox Relay (ADR-003): единый релей для опроса outbox_events
	kafkaProducer := kafka.NewProducer(kafka.Config{
		Brokers:  cfg.Kafka.Brokers,
		ClientID: "dopamine-worker-outbox",
	}, "")
	defer func() {
		if closeErr := kafkaProducer.Close(); closeErr != nil {
			log.Error("closing kafka producer", slog.String("error", closeErr.Error()))
		}
	}()

	outboxRelay := outbox.NewRelay(dbPool, kafkaProducer, []string{"order"}, log, 500*time.Millisecond)

	var g errgroup.Group

	// 1. Запуск outbox relay
	g.Go(func() error {
		log.Info("starting outbox relay", slog.Any("schemas", []string{"order"}))
		if err := outboxRelay.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Error("outbox relay failed", slog.String("error", err.Error()))
			return err
		}
		return nil
	})

	// 2. Запуск Kafka consumer для order-service (ADR-003: consumer group per module)
	if len(cfg.Kafka.Brokers) > 0 {
		orderConsumer := kafka.NewConsumer(
			kafka.Config{Brokers: cfg.Kafka.Brokers, ClientID: "dopamine-order-consumer"},
			"dopamine.events",
			"order-service-group",
		)
		defer func() {
			if closeErr := orderConsumer.Close(); closeErr != nil {
				log.Error("closing order consumer", slog.String("error", closeErr.Error()))
			}
		}()

		g.Go(func() error {
			log.Info("starting order kafka consumer")
			handler := orderModule.ConsumerHandler().HandleMessage
			if err := orderConsumer.Run(ctx, handler); err != nil && !errors.Is(err, context.Canceled) {
				log.Error("order consumer failed", slog.String("error", err.Error()))
				return err
			}
			return nil
		})
	}

	log.Info("worker bootstrap complete, background jobs running")

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("worker background task error: %w", err)
	}

	log.Info("worker shutdown complete")
	return nil
}
