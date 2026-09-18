package ratelimit

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter реализует ограничение частоты запросов на базе Redis (NFR-SEC-02).
type Limiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

// New создаёт новый экземпляр Limiter.
func New(client *redis.Client, limit int, window time.Duration) *Limiter {
	return &Limiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

// Middleware возвращает HTTP middleware, ограничивающий запросы по IP клиента.
func (l *Limiter) Middleware(keyPrefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			key := fmt.Sprintf("ratelimit:%s:%s", keyPrefix, ip)

			ctx := r.Context()
			count, err := l.client.Incr(ctx, key).Result()
			if err != nil {
				// При сбое Redis пропускаем запрос, чтобы не ломать доступность
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				_ = l.client.Expire(ctx, key, l.window).Err()
			}

			if count > int64(l.limit) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", l.window.Seconds()))
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"too_many_requests","message":"rate limit exceeded, please try again later"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
