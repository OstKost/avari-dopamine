package tracing

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type contextKey string

const (
	TraceparentHeader            = "traceparent"
	TraceIDContextKey  contextKey = "trace_id"
)

var (
	propagator = propagation.TraceContext{}
	tracer     = otel.GetTracerProvider().Tracer("dopamine-market")
)

func Tracer() trace.Tracer {
	if tracer == nil {
		return noop.NewTracerProvider().Tracer("dopamine-market")
	}
	return tracer
}

// InjectTraceContext упаковывает текущий traceparent в заголовки Kafka сообщения.
func InjectTraceContext(ctx context.Context, headers map[string]string) {
	if headers == nil {
		return
	}
	carrier := propagation.MapCarrier(headers)
	propagator.Inject(ctx, carrier)
}

// ExtractTraceContext извлекает trace context из заголовков сообщения.
func ExtractTraceContext(ctx context.Context, headers map[string]string) context.Context {
	if headers == nil {
		return ctx
	}
	carrier := propagation.MapCarrier(headers)
	return propagator.Extract(ctx, carrier)
}

// TraceMiddleware извлекает или генерирует trace_id для HTTP запроса.
func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))

		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			span := trace.SpanFromContext(ctx)
			if span.SpanContext().IsValid() {
				traceID = span.SpanContext().TraceID().String()
			} else {
				traceID = uuid.New().String()
			}
		}

		ctx = context.WithValue(ctx, TraceIDContextKey, traceID)
		w.Header().Set("X-Trace-ID", traceID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
