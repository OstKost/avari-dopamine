package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/port"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/usecase"
	"github.com/shopspring/decimal"
)

type mockOrderRepo struct {
	orders       map[uuid.UUID]*domain.Order
	history      map[uuid.UUID][]port.StatusHistoryRecord
	events       map[string]bool
	hasPendingFn func(userID uuid.UUID) bool
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{
		orders:  make(map[uuid.UUID]*domain.Order),
		history: make(map[uuid.UUID][]port.StatusHistoryRecord),
		events:  make(map[string]bool),
	}
}

func (m *mockOrderRepo) Create(ctx context.Context, order *domain.Order) error {
	m.orders[order.ID()] = order
	return nil
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, orderID uuid.UUID, fromStatus, toStatus domain.Status, reason string) error {
	o, ok := m.orders[orderID]
	if !ok {
		return domain.ErrOrderNotFound
	}
	_ = o.TransitionTo(toStatus)
	m.history[orderID] = append(m.history[orderID], port.StatusHistoryRecord{
		ID:         uuid.New(),
		OrderID:    orderID,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		Reason:     reason,
		CreatedAt:  time.Now(),
	})
	return nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	o, ok := m.orders[id]
	if !ok {
		return nil, domain.ErrOrderNotFound
	}
	return o, nil
}

func (m *mockOrderRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Order, int, error) {
	var list []*domain.Order
	for _, o := range m.orders {
		if o.UserID() == userID {
			list = append(list, o)
		}
	}
	return list, len(list), nil
}

func (m *mockOrderRepo) HasActivePendingOrder(ctx context.Context, userID uuid.UUID) (bool, error) {
	if m.hasPendingFn != nil {
		return m.hasPendingFn(userID), nil
	}
	for _, o := range m.orders {
		if o.UserID() == userID && o.Status() == domain.StatusPaymentPending {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockOrderRepo) GetStatusHistory(ctx context.Context, orderID uuid.UUID) ([]port.StatusHistoryRecord, error) {
	return m.history[orderID], nil
}

func (m *mockOrderRepo) IsEventProcessed(ctx context.Context, eventID, consumerName string) (bool, error) {
	return m.events[eventID+consumerName], nil
}

func (m *mockOrderRepo) MarkEventProcessed(ctx context.Context, eventID, consumerName string) error {
	m.events[eventID+consumerName] = true
	return nil
}

type mockCartLookup struct {
	cart    contracts.CartSnapshot
	cleared bool
}

func (m *mockCartLookup) GetCart(ctx context.Context, userID uuid.UUID) (contracts.CartSnapshot, error) {
	return m.cart, nil
}

func (m *mockCartLookup) ClearCart(ctx context.Context, userID uuid.UUID) error {
	m.cleared = true
	m.cart.Items = nil
	return nil
}

type mockProductLookup struct {
	products map[uuid.UUID]contracts.ProductSnapshot
}

func (m *mockProductLookup) GetByID(ctx context.Context, id uuid.UUID) (contracts.ProductSnapshot, error) {
	return m.products[id], nil
}

func (m *mockProductLookup) GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]contracts.ProductSnapshot, error) {
	return m.products, nil
}

func (m *mockProductLookup) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := m.products[id]
	return ok, nil
}

type mockPickupLookup struct {
	point contracts.PickupPointSnapshot
}

func (m *mockPickupLookup) GetByID(ctx context.Context, id uuid.UUID) (contracts.PickupPointSnapshot, error) {
	return m.point, nil
}

func (m *mockPickupLookup) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return true, nil
}

func TestOrderUseCase_CreateOrder_EmptyCartFails(t *testing.T) {
	repo := newMockOrderRepo()
	cartLookup := &mockCartLookup{}
	productLookup := &mockProductLookup{}
	pickupLookup := &mockPickupLookup{}

	uc := usecase.NewOrderUseCase(repo, cartLookup, productLookup, pickupLookup, nil)
	ctx := context.Background()

	_, err := uc.CreateOrder(ctx, uuid.New())
	if !errors.Is(err, domain.ErrCartEmpty) {
		t.Fatalf("expected ErrCartEmpty, got %v", err)
	}
}

func TestOrderUseCase_CreateOrder_DuplicatePendingFails(t *testing.T) {
	repo := newMockOrderRepo()
	repo.hasPendingFn = func(userID uuid.UUID) bool { return true }

	pointID := uuid.New()
	cartLookup := &mockCartLookup{
		cart: contracts.CartSnapshot{
			UserID:        uuid.New(),
			PickupPointID: &pointID,
			Items: []contracts.CartItemSnapshot{
				{ProductID: uuid.New(), Quantity: 1},
			},
		},
	}
	productLookup := &mockProductLookup{}
	pickupLookup := &mockPickupLookup{}

	uc := usecase.NewOrderUseCase(repo, cartLookup, productLookup, pickupLookup, nil)
	ctx := context.Background()

	_, err := uc.CreateOrder(ctx, uuid.New())
	if !errors.Is(err, domain.ErrDuplicatePendingOrder) {
		t.Fatalf("expected ErrDuplicatePendingOrder (INV-02), got %v", err)
	}
}

func TestOrderUseCase_TransitionOrder(t *testing.T) {
	repo := newMockOrderRepo()
	uc := usecase.NewOrderUseCase(repo, nil, nil, nil, nil)
	ctx := context.Background()

	orderID := uuid.New()
	userID := uuid.New()
	pickup := domain.PickupPointInfo{ID: uuid.New(), Name: "ПВЗ 1"}
	item := domain.NewOrderItem(uuid.New(), uuid.New(), "Товар", "Категория", decimal.NewFromInt(10), "seed", 1)

	order, err := domain.NewOrder(orderID, userID, pickup, []domain.OrderItem{item}, time.Now())
	if err != nil {
		t.Fatalf("NewOrder failed: %v", err)
	}
	_ = repo.Create(ctx, order)

	// Transition: created -> payment_pending
	_, err = uc.TransitionOrderStatus(ctx, orderID, domain.StatusPaymentPending, "checkout")
	if err != nil {
		t.Fatalf("transition failed: %v", err)
	}

	// Transition: payment_pending -> paid
	_, err = uc.TransitionOrderStatus(ctx, orderID, domain.StatusPaid, "payment confirmed")
	if err != nil {
		t.Fatalf("transition failed: %v", err)
	}

	// Invalid transition: paid -> delivered (skipping stages)
	_, err = uc.TransitionOrderStatus(ctx, orderID, domain.StatusDelivered, "invalid")
	if !errors.Is(err, domain.ErrInvalidStatusTransition) {
		t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
	}
}
