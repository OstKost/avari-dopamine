// cmd/server — HTTP API процесс. Единственная точка входа, где все модули
// инстанцируются и связываются друг с другом (composition root, ADR-004).
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

	"github.com/go-chi/chi/v5"
	"github.com/ostkost/dopamine-market/api/internal/modules/cart"
	"github.com/ostkost/dopamine-market/api/internal/modules/catalog"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity"
	"github.com/ostkost/dopamine-market/api/internal/modules/order"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment"
	"github.com/ostkost/dopamine-market/api/internal/modules/pickup"
	"github.com/ostkost/dopamine-market/api/internal/platform/config"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
	"github.com/ostkost/dopamine-market/api/internal/platform/httpserver"
	"github.com/ostkost/dopamine-market/api/internal/platform/logger"
	"github.com/ostkost/dopamine-market/api/internal/platform/metrics"
	"github.com/ostkost/dopamine-market/api/internal/platform/random"
	"github.com/ostkost/dopamine-market/api/internal/platform/ratelimit"
	"github.com/ostkost/dopamine-market/api/internal/platform/redis"
	"github.com/ostkost/dopamine-market/api/internal/platform/tracing"
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

	// 1. Модуль Identity & Auth (EPIC-01)
	identityModule := identity.NewModule(dbPool.Raw(), redisClient.Raw(), identity.Config{
		JWTSecret:       cfg.Auth.JWTSecret,
		AccessTokenTTL:  cfg.Auth.AccessTokenTTL,
		RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
		IsSecureCookie:  cfg.Env == "production",
	})

	// 2. Модуль Catalog & Synthetic Data (EPIC-02)
	catalogModule := catalog.NewModule(dbPool.Raw())

	// 3. Модуль Pickup Points (EPIC-03)
	rnd := random.New(time.Now().UnixNano())
	pickupModule := pickup.NewModule(dbPool.Raw(), rnd)

	// 4. Модуль Cart (EPIC-04)
	cartModule := cart.NewModule(redisClient.Raw(), catalogModule, pickupModule, 7*24*time.Hour)

	// 5. Модуль Order Lifecycle (EPIC-05)
	orderModule := order.NewModule(dbPool, cartModule, catalogModule, pickupModule)

	// 6. Модуль Payment Abstraction (EPIC-06)
	paymentModule, err := payment.NewModule(dbPool, cfg.Payment, rnd)
	if err != nil {
		return fmt.Errorf("initializing payment module: %w", err)
	}

	// TODO(EPIC-07 delivery), TODO(EPIC-08 notification): см. соответствующие Epic-файлы docs/epics/.

	srv := httpserver.New(httpserver.Options{
		Port:            cfg.HTTP.Port,
		ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
		Logger:          log,
		HealthCheckers: map[string]httpserver.HealthChecker{
			"postgres": dbPool,
			"redis":    redisClient,
		},
	})

	// OpenTelemetry трейсинг middleware и Prometheus метрики (EPIC-12)
	srv.Router().Use(tracing.TraceMiddleware)
	srv.Router().Handle("/metrics", metrics.Handler())

	// Rate limiter для Auth (NFR-SEC-02: 5 запросов/мин на IP)
	authLimiter := ratelimit.New(redisClient.Raw(), 5, time.Minute)

	// Монтирование роутов модулей
	srv.Router().Group(func(r chi.Router) {
		r.Use(authLimiter.Middleware("auth"))
		r.Mount("/auth", identityModule.Routes())
	})

	// Эндпоинт текущего пользователя /auth/me под RequireAuth
	srv.Router().Group(func(r chi.Router) {
		r.Use(httpserver.RequireAuth(cfg.Auth.JWTSecret))
		r.Get("/auth/me", identityModule.Handler().HandleMe)
	})

	// Каталог — публичный доступ
	srv.Router().Mount("/catalog", catalogModule.Routes())

	// ПВЗ — требует авторизации
	srv.Router().Group(func(r chi.Router) {
		r.Use(httpserver.RequireAuth(cfg.Auth.JWTSecret))
		r.Mount("/pickup", pickupModule.Routes())
	})

	// Корзина — требует авторизации
	srv.Router().Group(func(r chi.Router) {
		r.Use(httpserver.RequireAuth(cfg.Auth.JWTSecret))
		r.Mount("/cart", cartModule.Routes())
	})

	// Заказы — требует авторизации
	srv.Router().Group(func(r chi.Router) {
		r.Use(httpserver.RequireAuth(cfg.Auth.JWTSecret))
		r.Mount("/orders", orderModule.Routes())
	})

	// Платежи и вебхуки (EPIC-06)
	srv.Router().Mount("/payments", paymentModule.Routes())

	log.Info("http server listening", slog.Int("port", cfg.HTTP.Port))

	if err := srv.Run(ctx, cfg.HTTP.ShutdownTimeout); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("running http server: %w", err)
	}

	log.Info("shutdown complete")
	return nil
}
