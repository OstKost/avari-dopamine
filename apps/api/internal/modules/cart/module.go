package cart

import (
	"context"
	"fmt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/cart/adapter/httpapi"
	"github.com/ostkost/dopamine-market/api/internal/modules/cart/adapter/redis"
	"github.com/ostkost/dopamine-market/api/internal/modules/cart/usecase"
	redisClient "github.com/redis/go-redis/v9"
)

type Module struct {
	cartUC  *usecase.CartUseCase
	handler *httpapi.Handler
}

func NewModule(
	rClient *redisClient.Client,
	productLookup contracts.ProductLookup,
	pickupLookup contracts.PickupPointLookup,
	ttl time.Duration,
) *Module {
	store := redis.NewCartStore(rClient)
	cartUC := usecase.NewCartUseCase(store, productLookup, pickupLookup, ttl)
	handler := httpapi.NewHandler(cartUC)

	return &Module{
		cartUC:  cartUC,
		handler: handler,
	}
}

func (m *Module) Routes() chi.Router {
	return m.handler.Routes()
}

// Реализация contracts.CartLookup

func (m *Module) GetCart(ctx context.Context, userID uuid.UUID) (contracts.CartSnapshot, error) {
	rawCart, err := m.cartUC.GetRawCart(ctx, userID)
	if err != nil {
		return contracts.CartSnapshot{}, fmt.Errorf("getting raw cart: %w", err)
	}

	items := rawCart.Items()
	snapshotItems := make([]contracts.CartItemSnapshot, 0, len(items))
	for id, qty := range items {
		snapshotItems = append(snapshotItems, contracts.CartItemSnapshot{
			ProductID: id,
			Quantity:  qty,
		})
	}

	return contracts.CartSnapshot{
		UserID:        userID,
		Items:         snapshotItems,
		PickupPointID: rawCart.PickupPointID(),
	}, nil
}

func (m *Module) ClearCart(ctx context.Context, userID uuid.UUID) error {
	return m.cartUC.ClearCart(ctx, userID)
}
