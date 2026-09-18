package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/notification/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/notification/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/pubsub"
)

type mockOrderLookup struct {
	orders map[uuid.UUID]contracts.OrderSnapshot
}

func (m *mockOrderLookup) GetOrderByID(ctx context.Context, orderID uuid.UUID) (contracts.OrderSnapshot, error) {
	o, ok := m.orders[orderID]
	if !ok {
		return contracts.OrderSnapshot{}, errors.New("order not found")
	}
	return o, nil
}

func (m *mockOrderLookup) GetActiveOrderByUserID(ctx context.Context, userID uuid.UUID) (*contracts.OrderSnapshot, error) {
	for _, o := range m.orders {
		if o.UserID == userID {
			return &o, nil
		}
	}
	return nil, nil
}

type mockDeliveryLookup struct {
	deliveries map[uuid.UUID]*contracts.DeliverySnapshot
}

func (m *mockDeliveryLookup) GetDeliveryByOrderID(ctx context.Context, orderID uuid.UUID) (*contracts.DeliverySnapshot, error) {
	d, ok := m.deliveries[orderID]
	if !ok {
		return nil, nil
	}
	return d, nil
}

func TestNotification_GetOrderSnapshot(t *testing.T) {
	t.Parallel()

	orderID := uuid.New()
	userID := uuid.New()
	otherUserID := uuid.New()

	orderLookup := &mockOrderLookup{
		orders: map[uuid.UUID]contracts.OrderSnapshot{
			orderID: {
				ID:             orderID,
				UserID:         userID,
				Status:         "in_transit",
				TotalAmountRUB: decimal.NewFromFloat(10.00),
				PickupPoint: contracts.PickupPointSnapshot{
					ID:             uuid.New(),
					Name:           "ПВЗ Центр",
					DistanceMeters: 250,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	deliveryLookup := &mockDeliveryLookup{
		deliveries: map[uuid.UUID]*contracts.DeliverySnapshot{
			orderID: {
				ID:            uuid.New(),
				OrderID:       orderID,
				Status:        "in_transit",
				CourierName:   "Михаил Кузнецов",
				CourierRating: 4.95,
			},
		},
	}

	hub := pubsub.NewHub()
	uc := usecase.NewNotificationUseCase(orderLookup, deliveryLookup, hub)

	ctx := context.Background()

	t.Run("owner access success", func(t *testing.T) {
		snapshot, err := uc.GetOrderSnapshot(ctx, orderID, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if snapshot.OrderID != orderID {
			t.Errorf("expected orderID %v, got %v", orderID, snapshot.OrderID)
		}
		if snapshot.Courier == nil || snapshot.Courier.Name != "Михаил Кузнецов" {
			t.Errorf("expected courier Михаил Кузнецов, got %+v", snapshot.Courier)
		}
		if snapshot.PickupPointName != "ПВЗ Центр" {
			t.Errorf("expected pickup point 'ПВЗ Центр', got %s", snapshot.PickupPointName)
		}
	})

	t.Run("forbidden for other user", func(t *testing.T) {
		_, err := uc.GetOrderSnapshot(ctx, orderID, otherUserID)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestNotification_PubSubStreaming(t *testing.T) {
	t.Parallel()

	orderID := uuid.New()
	hub := pubsub.NewHub()
	uc := usecase.NewNotificationUseCase(nil, nil, hub)

	eventsCh, unsubscribe := uc.Subscribe(orderID)
	defer unsubscribe()

	msg := domain.OrderStatusChangedMessage{
		OrderID: orderID,
		Status:  "delivered",
	}

	uc.BroadcastOrderEvent(orderID, "status_changed", msg)

	select {
	case event := <-eventsCh:
		if event.Type != "status_changed" {
			t.Errorf("expected event type 'status_changed', got %s", event.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for pubsub event")
	}
}
