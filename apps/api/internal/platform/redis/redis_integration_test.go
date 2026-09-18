//go:build integration

// Требует реального Redis (REDIS_ADDR). См. пояснение в
// internal/platform/db/db_integration_test.go — тот же паттерн и то же
// ограничение окружения агента (нет Docker в песочнице на момент EPIC-00).
//
//	REDIS_ADDR=localhost:6379 go test -tags=integration ./internal/platform/redis/...
package redis_test

import (
	"context"
	"os"
	"testing"

	"github.com/ostkost/dopamine-market/api/internal/platform/redis"
)

func TestNew_ConnectsAndHealthChecks(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR not set")
	}

	ctx := context.Background()
	client, err := redis.New(ctx, redis.Config{Addr: addr})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer client.Close()

	if err := client.HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck() error: %v", err)
	}
}
