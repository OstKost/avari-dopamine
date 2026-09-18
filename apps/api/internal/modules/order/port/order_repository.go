package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/domain"
)

type StatusHistoryRecord struct {
	ID         uuid.UUID
	OrderID    uuid.UUID
	FromStatus domain.Status
	ToStatus   domain.Status
	Reason     string
	CreatedAt  time.Time
}

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	UpdateStatus(ctx context.Context, orderID uuid.UUID, fromStatus, toStatus domain.Status, reason string) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Order, int, error)
	HasActivePendingOrder(ctx context.Context, userID uuid.UUID) (bool, error)
	GetStatusHistory(ctx context.Context, orderID uuid.UUID) ([]StatusHistoryRecord, error)
	IsEventProcessed(ctx context.Context, eventID, consumerName string) (bool, error)
	MarkEventProcessed(ctx context.Context, eventID, consumerName string) error
}
