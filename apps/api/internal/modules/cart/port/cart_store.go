package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/cart/domain"
)

type CartStore interface {
	GetCart(ctx context.Context, userID uuid.UUID) (domain.Cart, error)
	SaveCart(ctx context.Context, userID uuid.UUID, cart domain.Cart, ttl time.Duration) error
	ClearCart(ctx context.Context, userID uuid.UUID) error
	Touch(ctx context.Context, userID uuid.UUID, ttl time.Duration) error
}
