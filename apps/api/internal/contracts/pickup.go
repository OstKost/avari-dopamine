package contracts

import (
	"context"

	"github.com/google/uuid"
)

// PickupPointSnapshot — снимок пункта выдачи для использования другими модулями (order, cart).
type PickupPointSnapshot struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Name           string    `json:"name"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	DistanceMeters float64   `json:"distance_meters"`
}

// PickupPointLookup — межмодульный интерфейс ПВЗ.
type PickupPointLookup interface {
	GetByID(ctx context.Context, id uuid.UUID) (PickupPointSnapshot, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}
