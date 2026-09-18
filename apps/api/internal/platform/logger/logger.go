// Package logger — единая точка структурного логирования для всего backend
// (ADR-010). Все сервисы обязаны логировать через этот пакет; прямой
// fmt.Println/log.Println запрещён вне cmd/*/main.go bootstrap-кода
// (AGENTS.md раздел 3).
//
// Формат — JSON slog-записи с обязательными полями timestamp/level/msg/
// service/module, дополняемые контекстными полями (trace_id, order_id,
// user_id) через context.Context, а не через явную передачу в каждый вызов.
package logger

import (
	"context"
	"log/slog"
	"os"
)

type ctxKey string

const (
	ctxKeyTraceID ctxKey = "trace_id"
	ctxKeySpanID  ctxKey = "span_id"
	ctxKeyModule  ctxKey = "module"
)

// Options конфигурирует New. Env=="local" переключает на человекочитаемый
// text-формат в stdout; любое другое значение — JSON (пригодный и для
// stdout в контейнере, и для последующего сбора GELF-хендлером).
type Options struct {
	Env         string // local | staging | production
	Level       string // debug | info | warn | error
	Service     string // "api" | "worker" | "seed" | "migrator"
	GraylogHost string // если пусто — GELF-транспорт не подключается
	GraylogPort int
}

// New собирает *slog.Logger согласно ADR-010. Возвращает также cleanup-функцию
// для graceful shutdown GELF-соединения (закрыть в defer в cmd/*/main.go).
func New(opts Options) (*slog.Logger, func(), error) {
	level := parseLevel(opts.Level)

	var handlers []slog.Handler

	if opts.Env == "local" {
		handlers = append(handlers, slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	} else {
		handlers = append(handlers, slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	}

	cleanup := func() {}

	if opts.GraylogHost != "" {
		gelfHandler, closeFn, err := newGELFHandler(opts.GraylogHost, opts.GraylogPort, level)
		if err != nil {
			return nil, cleanup, err
		}
		handlers = append(handlers, gelfHandler)
		cleanup = closeFn
	}

	var handler slog.Handler
	if len(handlers) == 1 {
		handler = handlers[0]
	} else {
		handler = multiHandler{handlers: handlers}
	}

	handler = &contextHandler{next: handler}

	base := slog.New(handler).With(
		slog.String("service", opts.Service),
	)

	return base, cleanup, nil
}

func parseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithModule возвращает context, помечающий все последующие логи в нём
// как принадлежащие указанному доменному модулю (identity, catalog, order...).
// Middleware/consumer bootstrap вызывает это один раз на входе в обработчик.
func WithModule(ctx context.Context, module string) context.Context {
	return context.WithValue(ctx, ctxKeyModule, module)
}

// WithTraceID прикрепляет trace_id к контексту — используется HTTP middleware
// и Kafka consumer wrapper для propagate trace через границы (ADR-010).
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ctxKeyTraceID, traceID)
}

// WithSpanID аналогично WithTraceID, но для span_id текущей операции.
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, ctxKeySpanID, spanID)
}

// FromContext возвращает logger, обогащённый module/trace_id/span_id из ctx,
// если они были туда положены. Использование:
//
//	log := logger.FromContext(ctx, baseLogger)
//	log.Info("order created", slog.String("order_id", id.String()))
func FromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	l := base
	if module, ok := ctx.Value(ctxKeyModule).(string); ok && module != "" {
		l = l.With(slog.String("module", module))
	}
	if traceID, ok := ctx.Value(ctxKeyTraceID).(string); ok && traceID != "" {
		l = l.With(slog.String("trace_id", traceID))
	}
	if spanID, ok := ctx.Value(ctxKeySpanID).(string); ok && spanID != "" {
		l = l.With(slog.String("span_id", spanID))
	}
	return l
}

// contextHandler — не используется напрямую для инъекции полей (это делает
// FromContext явно на стороне вызывающего кода), но зарезервирован как точка
// расширения, если в будущем понадобится автоматическая инъекция trace_id
// из ctx на каждый Handle-вызов без явного FromContext. Пока — простой passthrough,
// чтобы не создавать скрытую магию, усложняющую чтение кода агентами.
type contextHandler struct {
	next slog.Handler
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.next.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{next: h.next.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{next: h.next.WithGroup(name)}
}

// multiHandler дублирует запись в несколько handler'ов (stdout + GELF).
type multiHandler struct {
	handlers []slog.Handler
}

func (m multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r.Clone()); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (m multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithAttrs(attrs)
	}
	return multiHandler{handlers: next}
}

func (m multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithGroup(name)
	}
	return multiHandler{handlers: next}
}
