package identity

import (
	"context"
	"fmt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity/adapter/argon2"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity/adapter/httpapi"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity/adapter/postgres"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity/adapter/redis"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity/usecase"
	redisClient "github.com/redis/go-redis/v9"
)

type Config struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	IsSecureCookie  bool
}

type Module struct {
	authUC  *usecase.AuthUseCase
	handler *httpapi.Handler
}

func NewModule(dbPool *pgxpool.Pool, rClient *redisClient.Client, cfg Config) *Module {
	userRepo := postgres.NewUserRepository(dbPool)
	tokenStore := redis.NewTokenStore(rClient)
	hasher := argon2.NewHasher()

	authUC := usecase.NewAuthUseCase(
		userRepo,
		tokenStore,
		hasher,
		cfg.JWTSecret,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	handler := httpapi.NewHandler(authUC, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, cfg.IsSecureCookie)

	return &Module{
		authUC:  authUC,
		handler: handler,
	}
}

func (m *Module) Routes() chi.Router {
	return m.handler.Routes()
}

func (m *Module) Handler() *httpapi.Handler {
	return m.handler
}

// Реализация contracts.IdentityLookup

func (m *Module) GetUserByID(ctx context.Context, id uuid.UUID) (contracts.UserInfo, error) {
	user, err := m.authUC.GetUserByID(ctx, id)
	if err != nil {
		return contracts.UserInfo{}, fmt.Errorf("getting user info: %w", err)
	}
	return contracts.UserInfo{
		ID:        user.ID(),
		Email:     user.Email().String(),
		CreatedAt: user.CreatedAt(),
	}, nil
}

func (m *Module) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := m.authUC.GetUserByID(ctx, id)
	if err != nil {
		return false, nil
	}
	return true, nil
}
