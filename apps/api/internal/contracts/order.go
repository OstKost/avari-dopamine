package contracts

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// OrderItemSnapshot — снимок товара в заказе.
type OrderItemSnapshot struct {
	ProductID    uuid.UUID       `json:"product_id"`
	Name         string          `json:"name"`
	CategoryName string          `json:"category_name"`
	PriceRUB     decimal.Decimal `json:"price_rub"`
	ImageSeed    string          `json:"image_seed"`
	Quantity     int             `json:"quantity"`
	SubtotalRUB  decimal.Decimal `json:"subtotal_rub"`
}

// OrderSnapshot — снимок заказа для межмодульного взаимодействия (payment, delivery, notification).
type OrderSnapshot struct {
	ID             uuid.UUID           `json:"id"`
	UserID         uuid.UUID           `json:"user_id"`
	Status         string              `json:"status"`
	TotalAmountRUB decimal.Decimal     `json:"total_amount_rub"` // Всегда 10.00 RUB (INV-01)
	PickupPoint    PickupPointSnapshot `json:"pickup_point"`
	Items          []OrderItemSnapshot `json:"items"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// OrderLookup — интерфейс доступа к заказам для других модулей.
type OrderLookup interface {
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (OrderSnapshot, error)
	GetActiveOrderByUserID(ctx context.Context, userID uuid.UUID) (*OrderSnapshot, error)
}
