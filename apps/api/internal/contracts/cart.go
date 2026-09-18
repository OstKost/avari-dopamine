package contracts

import (
	"context"

	"github.com/google/uuid"
)

// CartItemSnapshot — элемент корзины.
type CartItemSnapshot struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// CartSnapshot — снимок корзины для модуля заказов (EPIC-05).
type CartSnapshot struct {
	UserID        uuid.UUID          `json:"user_id"`
	Items         []CartItemSnapshot `json:"items"`
	PickupPointID *uuid.UUID         `json:"pickup_point_id,omitempty"`
}

// CartLookup — межмодульный интерфейс корзины.
type CartLookup interface {
	GetCart(ctx context.Context, userID uuid.UUID) (CartSnapshot, error)
	ClearCart(ctx context.Context, userID uuid.UUID) error
}
