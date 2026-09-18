package delivery

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	kafkaadapter "github.com/ostkost/dopamine-market/api/internal/modules/delivery/adapter/kafka"
	"github.com/ostkost/dopamine-market/api/internal/modules/delivery/adapter/postgres"
	"github.com/ostkost/dopamine-market/api/internal/modules/delivery/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/delivery/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
	"github.com/ostkost/dopamine-market/api/internal/platform/random"
	"github.com/ostkost/dopamine-market/api/internal/platform/scheduler"
)

var _ contracts.DeliveryLookup = (*Module)(nil)

type Module struct {
	uc              *usecase.DeliveryUseCase
	consumerHandler *kafkaadapter.ConsumerHandler
}

func NewModule(pool *db.Pool, rnd random.Source, cfg usecase.Config) *Module {
	repo := postgres.NewDeliveryRepository(pool)
	uc := usecase.NewDeliveryUseCase(repo, pool, rnd, cfg)
	consumer := kafkaadapter.NewConsumerHandler(uc)

	return &Module{
		uc:              uc,
		consumerHandler: consumer,
	}
}

func (m *Module) ConsumerHandler() *kafkaadapter.ConsumerHandler {
	return m.consumerHandler
}

func (m *Module) DeliveryUseCase() *usecase.DeliveryUseCase {
	return m.uc
}

// SchedulerHandler возвращает функцию обработки запланированных переходов для scheduler.Run (ADR-005).
func (m *Module) SchedulerHandler() scheduler.Handler {
	return func(ctx context.Context, item scheduler.Transition) error {
		_, err := m.uc.AdvanceDeliveryState(ctx, item.DeliveryID, domain.Status(item.TargetStatus))
		return err
	}
}

// GetDeliveryByOrderID реализует контракт contracts.DeliveryLookup для других модулей (notification, order).
func (m *Module) GetDeliveryByOrderID(ctx context.Context, orderID uuid.UUID) (*contracts.DeliverySnapshot, error) {
	del, err := m.uc.GetDeliveryByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrDeliveryNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting delivery by order ID: %w", err)
	}

	return &contracts.DeliverySnapshot{
		ID:                    del.ID,
		OrderID:               del.OrderID,
		Status:                string(del.Status),
		CourierName:           del.CourierName,
		CourierRating:         del.CourierRating,
		StartedAt:             del.StartedAt,
		EstimatedCompletionAt: del.EstimatedCompletionAt,
		CreatedAt:             del.CreatedAt,
		UpdatedAt:             del.UpdatedAt,
	}, nil
}
