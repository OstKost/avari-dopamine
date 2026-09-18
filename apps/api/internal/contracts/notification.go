package contracts

import (
	"context"

	"github.com/google/uuid"
)

// OrderEventBroadcast — контракт для трансляции событий заказов в модуль уведомлений (ADR-004, ADR-009).
type OrderEventBroadcast interface {
	BroadcastOrderEvent(ctx context.Context, orderID uuid.UUID, eventType string, payload any) error
}
