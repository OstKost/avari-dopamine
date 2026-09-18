package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/kafka"
)

type ConsumerHandler struct {
	orderUC *usecase.OrderUseCase
}

func NewConsumerHandler(orderUC *usecase.OrderUseCase) *ConsumerHandler {
	return &ConsumerHandler{orderUC: orderUC}
}

type PaymentEventPayload struct {
	OrderID uuid.UUID `json:"order_id"`
	Reason  string    `json:"reason,omitempty"`
}

type DeliveryEventPayload struct {
	OrderID   uuid.UUID `json:"order_id"`
	NewStatus string    `json:"new_status"`
	Reason    string    `json:"reason,omitempty"`
}

func (h *ConsumerHandler) HandleMessage(ctx context.Context, msg kafka.Message) error {
	eventType := msg.Headers["event_type"]
	eventID := msg.Headers["event_id"]
	if eventID == "" {
		eventID = string(msg.Key)
	}

	return h.orderUC.ProcessExternalEvent(ctx, eventID, "order-service", func(ctx context.Context) error {
		switch eventType {
		case "payment.succeeded.v1":
			var p PaymentEventPayload
			if err := json.Unmarshal(msg.Value, &p); err != nil {
				return fmt.Errorf("unmarshaling payment event: %w", err)
			}
			_, err := h.orderUC.TransitionOrderStatus(ctx, p.OrderID, domain.StatusPaid, "payment succeeded")
			return err

		case "payment.failed.v1":
			var p PaymentEventPayload
			if err := json.Unmarshal(msg.Value, &p); err != nil {
				return fmt.Errorf("unmarshaling payment event: %w", err)
			}
			_, err := h.orderUC.TransitionOrderStatus(ctx, p.OrderID, domain.StatusPaymentFailed, p.Reason)
			return err

		case "delivery.status_changed.v1":
			var p DeliveryEventPayload
			if err := json.Unmarshal(msg.Value, &p); err != nil {
				return fmt.Errorf("unmarshaling delivery status event: %w", err)
			}
			_, err := h.orderUC.TransitionOrderStatus(ctx, p.OrderID, domain.Status(p.NewStatus), p.Reason)
			return err

		case "delivery.completed.v1":
			var p DeliveryEventPayload
			if err := json.Unmarshal(msg.Value, &p); err != nil {
				return fmt.Errorf("unmarshaling delivery completed event: %w", err)
			}
			_, err := h.orderUC.TransitionOrderStatus(ctx, p.OrderID, domain.StatusDelivered, "order delivered to pickup point")
			return err

		default:
			// Неизвестное событие игнорируем без ошибки
			return nil
		}
	})
}
