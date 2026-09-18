package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/cart/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/cart/usecase"
	"github.com/shopspring/decimal"
)

type mockCartStore struct {
	carts map[uuid.UUID]domain.Cart
}

func newMockCartStore() *mockCartStore {
	return &mockCartStore{carts: make(map[uuid.UUID]domain.Cart)}
}

func (m *mockCartStore) GetCart(ctx context.Context, userID uuid.UUID) (domain.Cart, error) {
	if c, ok := m.carts[userID]; ok {
		return c, nil
	}
	return domain.NewCart(nil, nil), nil
}

func (m *mockCartStore) SaveCart(ctx context.Context, userID uuid.UUID, cart domain.Cart, ttl time.Duration) error {
	m.carts[userID] = cart
	return nil
}

func (m *mockCartStore) ClearCart(ctx context.Context, userID uuid.UUID) error {
	delete(m.carts, userID)
	return nil
}

func (m *mockCartStore) Touch(ctx context.Context, userID uuid.UUID, ttl time.Duration) error {
	return nil
}

type mockProductLookup struct {
	products map[uuid.UUID]contracts.ProductSnapshot
}

func (m *mockProductLookup) GetByID(ctx context.Context, id uuid.UUID) (contracts.ProductSnapshot, error) {
	if p, ok := m.products[id]; ok {
		return p, nil
	}
	return contracts.ProductSnapshot{}, errors.New("not found")
}

func (m *mockProductLookup) GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]contracts.ProductSnapshot, error) {
	res := make(map[uuid.UUID]contracts.ProductSnapshot)
	for _, id := range ids {
		if p, ok := m.products[id]; ok {
			res[id] = p
		}
	}
	return res, nil
}

func (m *mockProductLookup) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := m.products[id]
	return ok, nil
}

type mockPickupLookup struct {
	points map[uuid.UUID]contracts.PickupPointSnapshot
}

func (m *mockPickupLookup) GetByID(ctx context.Context, id uuid.UUID) (contracts.PickupPointSnapshot, error) {
	if p, ok := m.points[id]; ok {
		return p, nil
	}
	return contracts.PickupPointSnapshot{}, errors.New("not found")
}

func (m *mockPickupLookup) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := m.points[id]
	return ok, nil
}

func TestCartUseCase_FullFlow(t *testing.T) {
	store := newMockCartStore()

	prodID1 := uuid.New()
	prodID2 := uuid.New()
	pickupID := uuid.New()
	userID := uuid.New()

	productLookup := &mockProductLookup{
		products: map[uuid.UUID]contracts.ProductSnapshot{
			prodID1: {
				ID:           prodID1,
				Name:         "Товар 1",
				PriceRUB:     decimal.NewFromInt(10),
				CategoryName: "Категория 1",
				ImageSeed:    "seed-1",
			},
			prodID2: {
				ID:           prodID2,
				Name:         "Товар 2",
				PriceRUB:     decimal.NewFromInt(10),
				CategoryName: "Категория 1",
				ImageSeed:    "seed-2",
			},
		},
	}

	pickupLookup := &mockPickupLookup{
		points: map[uuid.UUID]contracts.PickupPointSnapshot{
			pickupID: {
				ID:             pickupID,
				Name:           "Пункт 1",
				Latitude:       47.2,
				Longitude:      39.7,
				DistanceMeters: 250,
			},
		},
	}

	uc := usecase.NewCartUseCase(store, productLookup, pickupLookup, 7*24*time.Hour)
	ctx := context.Background()

	// 1. Add non-existing product must return ErrProductNotFound
	_, err := uc.AddItem(ctx, userID, uuid.New(), 1)
	if !errors.Is(err, domain.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}

	// 2. Add valid items
	cart, err := uc.AddItem(ctx, userID, prodID1, 2)
	if err != nil {
		t.Fatalf("AddItem failed: %v", err)
	}
	if cart.TotalQuantity != 2 || len(cart.Items) != 1 {
		t.Errorf("expected 2 items, got %d", cart.TotalQuantity)
	}

	// 3. Set pickup point
	cart, err = uc.SetPickupPoint(ctx, userID, pickupID)
	if err != nil {
		t.Fatalf("SetPickupPoint failed: %v", err)
	}
	if cart.PickupPoint == nil || cart.PickupPoint.ID != pickupID {
		t.Errorf("expected pickup point %v", pickupID)
	}

	// 4. Update quantity
	cart, err = uc.UpdateQuantity(ctx, userID, prodID1, 5)
	if err != nil {
		t.Fatalf("UpdateQuantity failed: %v", err)
	}
	if cart.TotalQuantity != 5 {
		t.Errorf("expected total quantity 5, got %d", cart.TotalQuantity)
	}

	// 5. Remove item
	cart, err = uc.RemoveItem(ctx, userID, prodID1)
	if err != nil {
		t.Fatalf("RemoveItem failed: %v", err)
	}
	if cart.TotalQuantity != 0 || len(cart.Items) != 0 {
		t.Errorf("expected empty cart items, got %d", cart.TotalQuantity)
	}
	// Pickup point is preserved
	if cart.PickupPoint == nil {
		t.Errorf("expected pickup point to be preserved")
	}
}
