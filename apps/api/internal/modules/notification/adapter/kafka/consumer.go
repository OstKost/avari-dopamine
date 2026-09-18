package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/notification/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/notification/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/kafka"
)

type ConsumerHandler struct {
	notifUC *usecase.NotificationUseCase
}

func NewConsumerHandler(notifUC *usecase.NotificationUseCase) *ConsumerHandler {
	return &ConsumerHandler{notifUC: notifUC}
}

type OrderEventPayload struct {
	OrderID uuid.UUID `json:"order_id"`
	Status  string    `json:"status"`
	Reason  string    `json:"reason,omitempty"`
}

type DeliveryEventPayload struct {
	DeliveryID    uuid.UUID `json:"delivery_id"`
	OrderID       uuid.UUID `json:"order_id"`
	Status        string    `json:"status"`
	CourierName   string    `json:"courier_name"`
	CourierRating float64   `json:"courier_rating"`
	Reason        string    `json:"reason,omitempty"`
}

func (h *ConsumerHandler) HandleMessage(ctx context.Context, msg kafka.Message) error {
	eventType := msg.Headers["event_type"]
	if eventType == "" {
		return nil
	}

	switch eventType {
	case "delivery.status_changed.v1", "delivery.completed.v1":
		var p DeliveryEventPayload
		if err := json.Unmarshal(msg.Value, &p); err != nil {
			return fmt.Errorf("unmarshaling delivery event in notification: %w", err)
		}

		msgOut := domain.OrderStatusChangedMessage{
			OrderID: p.OrderID,
			Status:  p.Status,
			Courier: &domain.CourierInfo{
				Name:   p.CourierName,
				Rating: p.CourierRating,
			},
			Reason: p.Reason,
		}
		h.notifUC.BroadcastOrderEvent(p.OrderID, "status_changed", msgOut)

	case "order.created.v1", "order.status_changed.v1", "order.cancelled.v1", "payment.succeeded.v1", "payment.failed.v1":
		var p OrderEventPayload
		if err := json.Unmarshal(msg.Value, &p); err != nil {
			return fmt.Errorf("unmarshaling order/payment event in notification: %w", err)
		}

		status := p.Status
		if status == "" {
			if eventType == "payment.succeeded.v1" {
				status = "paid"
			} else if eventType == "payment.failed.v1" {
				status = "payment_failed"
			}
		}

		msgOut := domain.OrderStatusChangedMessage{
			OrderID: p.OrderID,
			Status:  status,
			Reason:  p.Reason,
		}
		h.notifUC.BroadcastOrderEvent(p.OrderID, "status_changed", msgOut)
	}

	return nil
}
