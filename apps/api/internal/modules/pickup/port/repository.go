package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/pickup/domain"
)

type PickupRepository interface {
	SavePickupPoints(ctx context.Context, points []*domain.PickupPoint) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.PickupPoint, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PickupPoint, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
