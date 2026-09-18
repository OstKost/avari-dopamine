package httpserver

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

const (
	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
)

// ContextWithUserID помещает user_id в контекст запроса.
func ContextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

// UserIDFromContext извлекает user_id из контекста запроса.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	val := ctx.Value(userIDContextKey)
	if val == nil {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// RequireAuth проверяет access JWT-токен из cookie или Authorization: Bearer заголовока.
func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := ""

			// 1. Сначала проверяем cookie (ADR-007: httpOnly cookies для веб-клиента)
			if cookie, err := r.Cookie(AccessTokenCookie); err == nil && cookie.Value != "" {
				tokenString = cookie.Value
			}

			// 2. Fallback на заголовок Authorization: Bearer (для API/Postman/мобильных клиентов)
			if tokenString == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if tokenString == "" {
				http.Error(w, `{"error":"unauthorized","message":"missing access token"}`, http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, `{"error":"unauthorized","message":"invalid or expired access token"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"unauthorized","message":"invalid token claims"}`, http.StatusUnauthorized)
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok {
				http.Error(w, `{"error":"unauthorized","message":"missing subject in token"}`, http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(sub)
			if err != nil {
				http.Error(w, `{"error":"unauthorized","message":"invalid user id in token"}`, http.StatusUnauthorized)
				return
			}

			ctx := ContextWithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
