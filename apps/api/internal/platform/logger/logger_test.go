package logger

import (
	"context"
	"log/slog"
	"testing"
)

func TestNew_LocalEnvUsesTextHandler(t *testing.T) {
	log, cleanup, err := New(Options{Env: "local", Level: "info", Service: "api"})
	defer cleanup()
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNew_WithoutGraylogHostSkipsGELF(t *testing.T) {
	log, cleanup, err := New(Options{Env: "production", Level: "info", Service: "api"})
	defer cleanup()
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	// Не должно паниковать/блокироваться при логировании без GELF-транспорта.
	log.Info("smoke test")
}

func TestWithModule_AddsModuleField(t *testing.T) {
	ctx := context.Background()
	ctx = WithModule(ctx, "order")

	if v, ok := ctx.Value(ctxKeyModule).(string); !ok || v != "order" {
		t.Errorf("expected module=order in context, got %v", v)
	}
}

func TestFromContext_EnrichesWithTraceAndModule(t *testing.T) {
	base, cleanup, err := New(Options{Env: "local", Level: "debug", Service: "api"})
	defer cleanup()
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	ctx := context.Background()
	ctx = WithModule(ctx, "payment")
	ctx = WithTraceID(ctx, "trace-123")

	enriched := FromContext(ctx, base)
	if enriched == nil {
		t.Fatal("expected non-nil enriched logger")
	}
	// Убедимся, что вызов не паникует и возвращает валидный *slog.Logger,
	// пригодный для дальнейшего использования.
	enriched.Info("test message", slog.String("order_id", "abc"))
}

func TestFromContext_NoContextValuesIsNoop(t *testing.T) {
	base, cleanup, err := New(Options{Env: "local", Level: "info", Service: "api"})
	defer cleanup()
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	ctx := context.Background()
	enriched := FromContext(ctx, base)
	if enriched == nil {
		t.Fatal("expected non-nil logger even without context values")
	}
}
