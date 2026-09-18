// cmd/migrator — CLI-обёртка над goose для применения Postgres-миграций.
//
// Физические .sql файлы лежат в
// internal/platform/migrations/sql/{module}_NNN_description.sql
// (см. internal/platform/migrations/migrations.go за объяснением, почему
// они не могут физически лежать в apps/api/migrations/ верхнего уровня —
// ограничение go:embed на "../" в паттернах). ADR-011 логическое
// разделение "миграции на модуль" сохраняется через префикс имени файла
// (module_NNN_*), не через директорию.
//
// Использование:
//
//	go run ./cmd/migrator up
//	go run ./cmd/migrator down
//	go run ./cmd/migrator status
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver для goose
	"github.com/pressly/goose/v3"

	"github.com/ostkost/dopamine-market/api/internal/platform/migrations"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: migrator <up|down|status|redo>")
	}
	command := os.Args[1]

	// Миграции работают вне обычной config.Load() валидации, так как
	// migrator может запускаться до того, как остальные секреты (JWT,
	// платёжные) настроены в окружении — нужен только DATABASE_URL.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer func() {
		_ = db.Close() // процесс всё равно завершается после CLI-команды; ошибка закрытия здесь не меняет результат выполнения миграции
	}()

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting goose dialect: %w", err)
	}

	ctx := context.Background()

	switch command {
	case "up":
		return goose.UpContext(ctx, db, "sql")
	case "down":
		return goose.DownContext(ctx, db, "sql")
	case "status":
		return goose.StatusContext(ctx, db, "sql")
	case "redo":
		return goose.RedoContext(ctx, db, "sql")
	default:
		return fmt.Errorf("unknown command: %q (expected up|down|status|redo)", command)
	}
}
