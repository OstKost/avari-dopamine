package notification

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/notification/adapter/httpapi"
	kafkaadapter "github.com/ostkost/dopamine-market/api/internal/modules/notification/adapter/kafka"
	"github.com/ostkost/dopamine-market/api/internal/modules/notification/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/pubsub"
)

var _ contracts.OrderEventBroadcast = (*Module)(nil)

type Module struct {
	uc              *usecase.NotificationUseCase
	handler         *httpapi.Handler
	consumerHandler *kafkaadapter.ConsumerHandler
}

func NewModule(
	orderLookup contracts.OrderLookup,
	deliveryLookup contracts.DeliveryLookup,
	hub *pubsub.Hub,
) *Module {
	uc := usecase.NewNotificationUseCase(orderLookup, deliveryLookup, hub)
	handler := httpapi.NewHandler(uc)
	consumer := kafkaadapter.NewConsumerHandler(uc)

	return &Module{
		uc:              uc,
		handler:         handler,
		consumerHandler: consumer,
	}
}

func (m *Module) Routes() chi.Router {
	return m.handler.Routes()
}

func (m *Module) ConsumerHandler() *kafkaadapter.ConsumerHandler {
	return m.consumerHandler
}

func (m *Module) NotificationUseCase() *usecase.NotificationUseCase {
	return m.uc
}

// BroadcastOrderEvent реализует контракт contracts.OrderEventBroadcast (ADR-004).
func (m *Module) BroadcastOrderEvent(ctx context.Context, orderID uuid.UUID, eventType string, payload any) error {
	m.uc.BroadcastOrderEvent(orderID, eventType, payload)
	return nil
}
