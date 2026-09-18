package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/kafka"
)

type ConsumerHandler struct {
	paymentUC *usecase.PaymentUseCase
}

func NewConsumerHandler(paymentUC *usecase.PaymentUseCase) *ConsumerHandler {
	return &ConsumerHandler{paymentUC: paymentUC}
}

type OrderCreatedPayload struct {
	OrderID uuid.UUID `json:"order_id"`
	UserID  uuid.UUID `json:"user_id"`
}

func (h *ConsumerHandler) HandleMessage(ctx context.Context, msg kafka.Message) error {
	eventType := msg.Headers["event_type"]
	if eventType != "order.created.v1" {
		return nil
	}

	var payload OrderCreatedPayload
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		return fmt.Errorf("unmarshaling order.created event: %w", err)
	}

	_, err := h.paymentUC.InitiatePayment(ctx, payload.OrderID, payload.UserID, "http://localhost:3000/orders/"+payload.OrderID.String())
	if err != nil {
		return fmt.Errorf("handling order.created in payment module: %w", err)
	}

	return nil
}
