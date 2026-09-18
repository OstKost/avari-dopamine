package domain

import (
	"time"

	"github.com/google/uuid"
)

// CourierInfo — информация о назначенном курьере для клиента (FR-DELIVERY-01, FR-NOTIF-01).
type CourierInfo struct {
	Name   string  `json:"name"`
	Rating float64 `json:"rating"`
}

// OrderSnapshotMessage — снимок состояния заказа при установлении SSE соединения.
type OrderSnapshotMessage struct {
	OrderID               uuid.UUID    `json:"order_id"`
	Status                string       `json:"status"`
	TotalAmountRUB        string       `json:"total_amount_rub"`
	Courier               *CourierInfo `json:"courier,omitempty"`
	PickupPointName       string       `json:"pickup_point_name"`
	DistanceMeters        float64      `json:"distance_meters"`
	EstimatedCompletionAt *time.Time   `json:"estimated_completion_at,omitempty"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`
}

// OrderStatusChangedMessage — событие изменения статуса заказа в реальном времени.
type OrderStatusChangedMessage struct {
	OrderID               uuid.UUID    `json:"order_id"`
	Status                string       `json:"status"`
	Courier               *CourierInfo `json:"courier,omitempty"`
	Reason                string       `json:"reason,omitempty"`
	EstimatedCompletionAt *time.Time   `json:"estimated_completion_at,omitempty"`
	Timestamp             time.Time    `json:"timestamp"`
}
