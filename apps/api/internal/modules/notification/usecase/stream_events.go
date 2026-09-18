package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/notification/domain"
	"github.com/ostkost/dopamine-market/api/internal/platform/pubsub"
)

type NotificationUseCase struct {
	orderLookup    contracts.OrderLookup
	deliveryLookup contracts.DeliveryLookup
	hub            *pubsub.Hub
}

func NewNotificationUseCase(
	orderLookup contracts.OrderLookup,
	deliveryLookup contracts.DeliveryLookup,
	hub *pubsub.Hub,
) *NotificationUseCase {
	return &NotificationUseCase{
		orderLookup:    orderLookup,
		deliveryLookup: deliveryLookup,
		hub:            hub,
	}
}

// GetOrderSnapshot возвращает начальный снимок состояния заказа с проверкой прав доступа (FR-NOTIF-01).
func (uc *NotificationUseCase) GetOrderSnapshot(ctx context.Context, orderID, userID uuid.UUID) (*domain.OrderSnapshotMessage, error) {
	order, err := uc.orderLookup.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrOrderNotFound, err)
	}

	// Авторизация: запрещено слушать события чужого заказа
	if order.UserID != userID {
		return nil, domain.ErrForbidden
	}

	snapshot := &domain.OrderSnapshotMessage{
		OrderID:         order.ID,
		Status:          order.Status,
		TotalAmountRUB:  order.TotalAmountRUB.StringFixed(2),
		PickupPointName: order.PickupPoint.Name,
		DistanceMeters:  order.PickupPoint.DistanceMeters,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}

	// Обогащаем данными о доставке, если доставка уже создана
	if uc.deliveryLookup != nil {
		delivery, err := uc.deliveryLookup.GetDeliveryByOrderID(ctx, orderID)
		if err == nil && delivery != nil {
			if delivery.CourierName != "" {
				snapshot.Courier = &domain.CourierInfo{
					Name:   delivery.CourierName,
					Rating: delivery.CourierRating,
				}
			}
			if !delivery.EstimatedCompletionAt.IsZero() {
				snapshot.EstimatedCompletionAt = &delivery.EstimatedCompletionAt
			}
			// Если доставка продвинулась дальше, статус заказа актуализируется
			if delivery.Status != "" {
				snapshot.Status = delivery.Status
			}
		}
	}

	return snapshot, nil
}

// Subscribe подписывается на обновления конкретного заказа.
func (uc *NotificationUseCase) Subscribe(orderID uuid.UUID) (<-chan pubsub.Event, func()) {
	return uc.hub.Subscribe(orderID)
}

// BroadcastOrderEvent рассылает событие всем активным подписчикам заказа.
func (uc *NotificationUseCase) BroadcastOrderEvent(orderID uuid.UUID, eventType string, payload any) {
	uc.hub.Publish(orderID, pubsub.Event{
		Type:    eventType,
		Payload: payload,
	})
}
