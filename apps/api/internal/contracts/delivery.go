package contracts

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DeliverySnapshot — снимок состояния доставки для межмодульного взаимодействия (notification, order).
type DeliverySnapshot struct {
	ID                    uuid.UUID `json:"id"`
	OrderID               uuid.UUID `json:"order_id"`
	Status                string    `json:"status"`
	CourierName           string    `json:"courier_name"`
	CourierRating         float64   `json:"courier_rating"`
	StartedAt             time.Time `json:"started_at"`
	EstimatedCompletionAt time.Time `json:"estimated_completion_at"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// DeliveryLookup — интерфейс доступа к доставкам для других модулей (ADR-001, ADR-004).
type DeliveryLookup interface {
	GetDeliveryByOrderID(ctx context.Context, orderID uuid.UUID) (*DeliverySnapshot, error)
}
