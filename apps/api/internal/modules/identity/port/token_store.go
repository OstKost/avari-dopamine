package port

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RefreshTokenStore управляет хранением сессий и ротацией refresh-токенов (ADR-007).
type RefreshTokenStore interface {
	SaveToken(ctx context.Context, familyID, tokenID, userID uuid.UUID, ttl time.Duration) error
	ValidateAndRotate(ctx context.Context, familyID, tokenID, newID, userID uuid.UUID, ttl time.Duration) error
	RevokeFamily(ctx context.Context, familyID uuid.UUID) error
}
