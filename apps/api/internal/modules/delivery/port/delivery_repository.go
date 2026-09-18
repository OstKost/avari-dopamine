package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/delivery/domain"
)

type DeliveryRepository interface {
	Save(ctx context.Context, delivery *domain.Delivery) error
	Update(ctx context.Context, delivery *domain.Delivery) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Delivery, error)
	FindByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Delivery, error)
}
