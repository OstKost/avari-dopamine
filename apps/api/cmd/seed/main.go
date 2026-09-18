package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ostkost/dopamine-market/api/internal/modules/catalog"
	"github.com/ostkost/dopamine-market/api/internal/platform/config"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: seed <command>")
		fmt.Fprintln(os.Stderr, "commands:")
		fmt.Fprintln(os.Stderr, "  catalog   generate synthetic product catalog (EPIC-02)")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "loading config: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbPool, err := db.New(ctx, db.Config{
		DSN:             cfg.DB.DSN,
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "connecting to db: %v\n", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	switch os.Args[1] {
	case "catalog":
		catalogMod := catalog.NewModule(dbPool.Raw())
		if err := seedCatalog(ctx, catalogMod.Repository()); err != nil {
			fmt.Fprintf(os.Stderr, "seeding catalog failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("catalog seed completed successfully")
	default:
		fmt.Fprintf(os.Stderr, "unknown seed command: %q\n", os.Args[1])
		os.Exit(1)
	}
}
