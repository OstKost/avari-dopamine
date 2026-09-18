package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/cart/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/cart/port"
	"github.com/shopspring/decimal"
)

const DefaultCartTTL = 7 * 24 * time.Hour

type EnrichedCartItem struct {
	ProductID    uuid.UUID       `json:"product_id"`
	CategoryID   uuid.UUID       `json:"category_id"`
	CategoryName string          `json:"category_name"`
	Name         string          `json:"name"`
	PriceRUB     decimal.Decimal `json:"price_rub"`
	ImageSeed    string          `json:"image_seed"`
	Quantity     int             `json:"quantity"`
	SubtotalRUB  decimal.Decimal `json:"subtotal_rub"`
}

type EnrichedCart struct {
	Items          []EnrichedCartItem             `json:"items"`
	PickupPoint    *contracts.PickupPointSnapshot `json:"pickup_point,omitempty"`
	TotalQuantity  int                            `json:"total_quantity"`
	TotalPriceRUB  decimal.Decimal                `json:"total_price_rub"`
	FixedOrderCost decimal.Decimal                `json:"fixed_order_cost"` // INV-01: 10.00 RUB
}

type CartUseCase struct {
	store         port.CartStore
	productLookup contracts.ProductLookup
	pickupLookup  contracts.PickupPointLookup
	ttl           time.Duration
}

func NewCartUseCase(
	store port.CartStore,
	productLookup contracts.ProductLookup,
	pickupLookup contracts.PickupPointLookup,
	ttl time.Duration,
) *CartUseCase {
	if ttl <= 0 {
		ttl = DefaultCartTTL
	}
	return &CartUseCase{
		store:         store,
		productLookup: productLookup,
		pickupLookup:  pickupLookup,
		ttl:           ttl,
	}
}

func (uc *CartUseCase) AddItem(ctx context.Context, userID, productID uuid.UUID, quantity int) (*EnrichedCart, error) {
	if quantity <= 0 {
		return nil, domain.ErrInvalidQuantity
	}

	exists, err := uc.productLookup.Exists(ctx, productID)
	if err != nil || !exists {
		return nil, domain.ErrProductNotFound
	}

	cart, err := uc.store.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting cart: %w", err)
	}

	updatedCart, err := cart.WithItem(productID, quantity)
	if err != nil {
		return nil, err
	}

	if err := uc.store.SaveCart(ctx, userID, updatedCart, uc.ttl); err != nil {
		return nil, fmt.Errorf("saving cart: %w", err)
	}

	return uc.enrichCart(ctx, updatedCart)
}

func (uc *CartUseCase) UpdateQuantity(ctx context.Context, userID, productID uuid.UUID, quantity int) (*EnrichedCart, error) {
	if quantity < 0 {
		return nil, domain.ErrInvalidQuantity
	}

	cart, err := uc.store.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting cart: %w", err)
	}

	if quantity > 0 {
		exists, err := uc.productLookup.Exists(ctx, productID)
		if err != nil || !exists {
			return nil, domain.ErrProductNotFound
		}
	}

	updatedCart, err := cart.WithQuantity(productID, quantity)
	if err != nil {
		return nil, err
	}

	if err := uc.store.SaveCart(ctx, userID, updatedCart, uc.ttl); err != nil {
		return nil, fmt.Errorf("saving cart: %w", err)
	}

	return uc.enrichCart(ctx, updatedCart)
}

func (uc *CartUseCase) RemoveItem(ctx context.Context, userID, productID uuid.UUID) (*EnrichedCart, error) {
	cart, err := uc.store.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting cart: %w", err)
	}

	updatedCart := cart.WithoutItem(productID)
	if err := uc.store.SaveCart(ctx, userID, updatedCart, uc.ttl); err != nil {
		return nil, fmt.Errorf("saving cart: %w", err)
	}

	return uc.enrichCart(ctx, updatedCart)
}

func (uc *CartUseCase) SetPickupPoint(ctx context.Context, userID, pointID uuid.UUID) (*EnrichedCart, error) {
	exists, err := uc.pickupLookup.Exists(ctx, pointID)
	if err != nil || !exists {
		return nil, domain.ErrPickupPointNotFound
	}

	cart, err := uc.store.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting cart: %w", err)
	}

	updatedCart := cart.WithPickupPoint(&pointID)
	if err := uc.store.SaveCart(ctx, userID, updatedCart, uc.ttl); err != nil {
		return nil, fmt.Errorf("saving cart: %w", err)
	}

	return uc.enrichCart(ctx, updatedCart)
}

func (uc *CartUseCase) GetCart(ctx context.Context, userID uuid.UUID) (*EnrichedCart, error) {
	cart, err := uc.store.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting cart: %w", err)
	}

	// Продлеваем TTL при чтении
	_ = uc.store.Touch(ctx, userID, uc.ttl)

	return uc.enrichCart(ctx, cart)
}

func (uc *CartUseCase) GetRawCart(ctx context.Context, userID uuid.UUID) (domain.Cart, error) {
	return uc.store.GetCart(ctx, userID)
}

func (uc *CartUseCase) ClearCart(ctx context.Context, userID uuid.UUID) error {
	return uc.store.ClearCart(ctx, userID)
}

func (uc *CartUseCase) enrichCart(ctx context.Context, cart domain.Cart) (*EnrichedCart, error) {
	items := cart.Items()
	var productIDs []uuid.UUID
	for id := range items {
		productIDs = append(productIDs, id)
	}

	products, err := uc.productLookup.GetByIDs(ctx, productIDs)
	if err != nil {
		return nil, fmt.Errorf("fetching product snapshots: %w", err)
	}

	enrichedItems := make([]EnrichedCartItem, 0, len(items))
	totalPrice := decimal.Zero
	totalQty := 0

	for id, qty := range items {
		prod, ok := products[id]
		if !ok {
			// Если товар был удален из каталога, пропускаем или показываем плейсхолдер
			continue
		}

		subtotal := prod.PriceRUB.Mul(decimal.NewFromInt(int64(qty)))
		totalPrice = totalPrice.Add(subtotal)
		totalQty += qty

		enrichedItems = append(enrichedItems, EnrichedCartItem{
			ProductID:    prod.ID,
			CategoryID:   prod.CategoryID,
			CategoryName: prod.CategoryName,
			Name:         prod.Name,
			PriceRUB:     prod.PriceRUB,
			ImageSeed:    prod.ImageSeed,
			Quantity:     qty,
			SubtotalRUB:  subtotal,
		})
	}

	var pickupSnapshot *contracts.PickupPointSnapshot
	if cart.PickupPointID() != nil {
		point, err := uc.pickupLookup.GetByID(ctx, *cart.PickupPointID())
		if err == nil {
			pickupSnapshot = &point
		}
	}

	return &EnrichedCart{
		Items:          enrichedItems,
		PickupPoint:    pickupSnapshot,
		TotalQuantity:  totalQty,
		TotalPriceRUB:  totalPrice,
		FixedOrderCost: decimal.NewFromInt(10), // INV-01
	}, nil
}
