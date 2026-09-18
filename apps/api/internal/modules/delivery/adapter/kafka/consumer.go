package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/delivery/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/kafka"
)

type ConsumerHandler struct {
	deliveryUC *usecase.DeliveryUseCase
}

func NewConsumerHandler(deliveryUC *usecase.DeliveryUseCase) *ConsumerHandler {
	return &ConsumerHandler{deliveryUC: deliveryUC}
}

type PaymentEventPayload struct {
	OrderID uuid.UUID `json:"order_id"`
}

func (h *ConsumerHandler) HandleMessage(ctx context.Context, msg kafka.Message) error {
	eventType := msg.Headers["event_type"]
	eventID := msg.Headers["event_id"]
	if eventID == "" {
		eventID = string(msg.Key)
	}

	return h.deliveryUC.ProcessExternalEvent(ctx, eventID, "delivery-service", func(ctx context.Context) error {
		switch eventType {
		case "payment.succeeded.v1":
			var p PaymentEventPayload
			if err := json.Unmarshal(msg.Value, &p); err != nil {
				return fmt.Errorf("unmarshaling payment event in delivery: %w", err)
			}
			_, err := h.deliveryUC.StartDelivery(ctx, p.OrderID)
			return err

		default:
			// Неизвестное событие игнорируем без ошибки
			return nil
		}
	})
}
