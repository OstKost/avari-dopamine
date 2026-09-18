package order

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/adapter/httpapi"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/adapter/kafka"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/adapter/postgres"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
)

type Module struct {
	orderUC         *usecase.OrderUseCase
	handler         *httpapi.Handler
	consumerHandler *kafka.ConsumerHandler
}

func NewModule(
	dbPool *db.Pool,
	cartLookup contracts.CartLookup,
	productLookup contracts.ProductLookup,
	pickupLookup contracts.PickupPointLookup,
) *Module {
	repo := postgres.NewOrderRepository(dbPool)
	orderUC := usecase.NewOrderUseCase(repo, cartLookup, productLookup, pickupLookup, dbPool)
	handler := httpapi.NewHandler(orderUC)
	consumerHandler := kafka.NewConsumerHandler(orderUC)

	return &Module{
		orderUC:         orderUC,
		handler:         handler,
		consumerHandler: consumerHandler,
	}
}

func (m *Module) Routes() chi.Router {
	return m.handler.Routes()
}

func (m *Module) ConsumerHandler() *kafka.ConsumerHandler {
	return m.consumerHandler
}

func (m *Module) UseCase() *usecase.OrderUseCase {
	return m.orderUC
}

// Реализация contracts.OrderLookup

func (m *Module) GetOrderByID(ctx context.Context, orderID uuid.UUID) (contracts.OrderSnapshot, error) {
	o, err := m.orderUC.GetByID(ctx, orderID)
	if err != nil {
		return contracts.OrderSnapshot{}, fmt.Errorf("getting order %s: %w", orderID, err)
	}

	items := o.Items()
	snapshotItems := make([]contracts.OrderItemSnapshot, 0, len(items))
	for _, it := range items {
		snapshotItems = append(snapshotItems, contracts.OrderItemSnapshot{
			ProductID:    it.ProductID(),
			Name:         it.Name(),
			CategoryName: it.CategoryName(),
			PriceRUB:     it.PriceRUB(),
			ImageSeed:    it.ImageSeed(),
			Quantity:     it.Quantity(),
			SubtotalRUB:  it.SubtotalRUB(),
		})
	}

	return contracts.OrderSnapshot{
		ID:             o.ID(),
		UserID:         o.UserID(),
		Status:         o.Status().String(),
		TotalAmountRUB: o.TotalAmountRUB(),
		PickupPoint: contracts.PickupPointSnapshot{
			ID:             o.PickupPoint().ID,
			UserID:         o.UserID(),
			Name:           o.PickupPoint().Name,
			Latitude:       o.PickupPoint().Latitude,
			Longitude:      o.PickupPoint().Longitude,
			DistanceMeters: o.PickupPoint().DistanceMeters,
		},
		Items:     snapshotItems,
		CreatedAt: o.CreatedAt(),
		UpdatedAt: o.UpdatedAt(),
	}, nil
}

func (m *Module) GetActiveOrderByUserID(ctx context.Context, userID uuid.UUID) (*contracts.OrderSnapshot, error) {
	orders, _, err := m.orderUC.ListByUserID(ctx, userID, 1, 0)
	if err != nil || len(orders) == 0 {
		return nil, nil
	}

	latest := orders[0]
	if latest.Status().IsTerminal() {
		return nil, nil
	}

	snapshot, err := m.GetOrderByID(ctx, latest.ID())
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}
