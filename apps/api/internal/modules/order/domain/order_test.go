package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/domain"
	"github.com/shopspring/decimal"
)

func TestOrder_Invariants(t *testing.T) {
	userID := uuid.New()
	pickup := domain.PickupPointInfo{
		ID:             uuid.New(),
		Name:           "ПВЗ №1",
		Latitude:       47.2,
		Longitude:      39.7,
		DistanceMeters: 150,
	}

	item := domain.NewOrderItem(
		uuid.New(),
		uuid.New(),
		"Товар 1",
		"Категория 1",
		decimal.NewFromInt(10),
		"seed-1",
		2,
	)

	// 1. Empty items fails with ErrCartEmpty
	_, err := domain.NewOrder(uuid.New(), userID, pickup, nil, time.Now())
	if !errors.Is(err, domain.ErrCartEmpty) {
		t.Fatalf("expected ErrCartEmpty, got %v", err)
	}

	// 2. Empty pickup point fails with ErrPickupPointRequired
	_, err = domain.NewOrder(uuid.New(), userID, domain.PickupPointInfo{}, []domain.OrderItem{item}, time.Now())
	if !errors.Is(err, domain.ErrPickupPointRequired) {
		t.Fatalf("expected ErrPickupPointRequired, got %v", err)
	}

	// 3. Valid creation: price is always 10.00 RUB (INV-01)
	order, err := domain.NewOrder(uuid.New(), userID, pickup, []domain.OrderItem{item}, time.Now())
	if err != nil {
		t.Fatalf("NewOrder failed: %v", err)
	}
	if !order.TotalAmountRUB().Equal(decimal.NewFromInt(10)) {
		t.Errorf("expected 10.00 RUB, got %v", order.TotalAmountRUB())
	}
	if order.Status() != domain.StatusCreated {
		t.Errorf("expected status created, got %v", order.Status())
	}

	// 4. State transition
	if err := order.TransitionTo(domain.StatusPaymentPending); err != nil {
		t.Fatalf("TransitionTo payment_pending failed: %v", err)
	}
	if order.Status() != domain.StatusPaymentPending {
		t.Errorf("status not updated")
	}

	// 5. Invalid transition
	if err := order.TransitionTo(domain.StatusDelivered); !errors.Is(err, domain.ErrInvalidStatusTransition) {
		t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
	}
}
