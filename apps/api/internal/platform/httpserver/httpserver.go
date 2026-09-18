// Package httpserver оборачивает go-chi/chi для HTTP-серверов apps/api
// (cmd/server). Предоставляет стандартный набор middleware и graceful
// shutdown — composition root собирает роутер модулей через Mount и
// передаёт сюда для запуска.
package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// HealthChecker — контракт, который реализуют db.Pool, redis.Client и
// (опционально) kafka-проверки. /healthz агрегирует все зарегистрированные
// проверки — сервис считается здоровым только если ВСЕ зависимости отвечают.
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// Server — тонкая обёртка над chi.Router + net/http.Server с graceful shutdown.
type Server struct {
	router *chi.Mux
	srv    *http.Server
	log    *slog.Logger
}

// Options конфигурирует Server.
type Options struct {
	Port            int
	ShutdownTimeout time.Duration
	Logger          *slog.Logger
	// HealthCheckers — именованные зависимости, проверяемые /healthz.
	// Имя используется только в диагностическом ответе при сбое.
	HealthCheckers map[string]HealthChecker
	// Middlewares — дополнительные глобальные middleware, выполняемые перед маршрутами.
	Middlewares []func(http.Handler) http.Handler
}

// New создаёт Server со стандартными middleware:
//   - RequestID: генерирует/пробрасывает X-Request-Id для трассировки
//   - Recoverer: перехватывает панику в хендлерах, не убивая процесс
//     целиком (важно в модульном монолите — паника одного модуля не должна
//     валить остальные, ADR-001 "последствия")
//   - CORS: настраивается здесь единообразно для всех модулей
//
// Аутентификация (RequireAuth) НЕ регистрируется здесь глобально — она
// применяется выборочно к защищённым роутам композиционным корнем модуля
// identity (EPIC-01), так как не все эндпоинты требуют авторизации
// (например GET /catalog/products доступен гостям).
func New(opts Options) *Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(recovererWithLogging(opts.Logger))
	r.Use(corsMiddleware())
	r.Use(middleware.Timeout(30 * time.Second))
	for _, mw := range opts.Middlewares {
		if mw != nil {
			r.Use(mw)
		}
	}

	r.Get("/healthz", healthzHandler(opts.HealthCheckers))

	return &Server{
		router: r,
		log:    opts.Logger,
		srv: &http.Server{
			Addr:         fmt.Sprintf(":%d", opts.Port),
			Handler:      r,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 60 * time.Second, // выше read timeout — учитывает долгоживущие SSE-соединения (ADR-009, EPIC-08)
		},
	}
}

// Router предоставляет доступ к chi.Mux для монтирования модульных роутов
// из composition root: srv.Router().Mount("/auth", identityRouter).
func (s *Server) Router() chi.Router {
	return s.router
}

// Run блокирует до отмены ctx, затем выполняет graceful shutdown с таймаутом.
func (s *Server) Run(ctx context.Context, shutdownTimeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("http server error: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
		return nil
	}
}

func healthzHandler(checkers map[string]HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		failures := map[string]string{}
		for name, checker := range checkers {
			if err := checker.HealthCheck(r.Context()); err != nil {
				failures[name] = err.Error()
			}
		}

		if len(failures) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprintf(w, `{"status":"unhealthy","failures":%q}`, failures)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	}
}

func corsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Конкретный origin (не "*") — необходимо для httpOnly cookie-based
			// auth (ADR-007), браузеры не отправляют cookie на wildcard CORS
			// с credentials:include. Разрешённый origin читается из конфигурации
			// в composition root и передаётся сюда через переменную окружения
			// в дальнейших доработках; на этапе bootstrap — permissive для
			// локальной разработки, ужесточяется в EPIC-01 при подключении auth.
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func recovererWithLogging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					if log != nil {
						log.Error("panic recovered in http handler",
							slog.Any("panic", rec),
							slog.String("path", r.URL.Path),
							slog.String("method", r.Method),
						)
					}
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"internal_server_error"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
