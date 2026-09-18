// cmd/seed — CLI для детерминированной генерации синтетических данных
// (ADR-012). Запускается один раз при первом поднятии окружения, НЕ на
// каждом старте сервера (в отличие от миграций, которые могут применяться
// многократно идемпотентно).
//
// Использование (после реализации подкоманд в EPIC-02/EPIC-03):
//
//	go run ./cmd/seed catalog   # генерирует 150-300 товаров, 8-12 категорий
//	go run ./cmd/seed pickup    # НЕ применимо — ПВЗ генерируются per-request
//	                              в runtime (EPIC-03), а не сидом при старте.
//
// На этапе EPIC-00 — только скелет с разбором подкоманды и понятной
// ошибкой на нереализованные команды.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: seed <command>")
		fmt.Fprintln(os.Stderr, "commands:")
		fmt.Fprintln(os.Stderr, "  catalog   generate synthetic product catalog (EPIC-02)")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "catalog":
		// TODO(EPIC-02): реализовать генерацию каталога через
		//   internal/modules/catalog/adapter/seed package, детерминированный
		//   seed=42 (ADR-012), идемпотентная запись (ON CONFLICT DO NOTHING).
		fmt.Fprintln(os.Stderr, "seed catalog: not implemented yet — see EPIC-02 (docs/epics/EPIC-02-catalog.md)")
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "unknown seed command: %q\n", os.Args[1])
		os.Exit(1)
	}
}
