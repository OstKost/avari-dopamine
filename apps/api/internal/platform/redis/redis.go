// Package redis предоставляет общий клиент Redis, используемый модулями
// cart (ADR-008), identity (refresh-токены, ADR-007) и platform/ratelimit.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config — параметры подключения, транслируемые из internal/platform/config.
type Config struct {
	Addr     string
	Password string
	DB       int
}

// Client оборачивает *redis.Client. Модули получают *Client через
// composition root, а не создают собственные соединения — единый пул
// соединений на процесс (ADR-002: single-node, экономия ресурсов).
type Client struct {
	rdb *redis.Client
}

// New создаёт клиента и проверяет связность (fail-fast).
func New(ctx context.Context, cfg Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("pinging redis at %s: %w", cfg.Addr, err)
	}

	return &Client{rdb: rdb}, nil
}

// Raw возвращает нижележащий *redis.Client для случаев, когда модулю нужен
// полный API (например Lua-скрипты для атомарных сложных операций в cart).
// Предпочтительно оборачивать часто используемые паттерны отдельными
// методами Client, а не вызывать Raw() повсеместно — это упрощает будущую
// замену клиента библиотеки при необходимости.
func (c *Client) Raw() *redis.Client {
	return c.rdb
}

// Close закрывает соединение. Вызывается один раз при graceful shutdown.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// HealthCheck используется /healthz эндпоинтом.
func (c *Client) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return c.rdb.Ping(ctx).Err()
}
