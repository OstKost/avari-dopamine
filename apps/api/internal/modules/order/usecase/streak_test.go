package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/port"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/usecase"
	"github.com/shopspring/decimal"
)

type mockStatsOrderRepo struct {
	orders []*domain.Order
}

func (r *mockStatsOrderRepo) Create(ctx context.Context, o *domain.Order) error {
	r.orders = append(r.orders, o)
	return nil
}

func (r *mockStatsOrderRepo) UpdateStatus(ctx context.Context, orderID uuid.UUID, fromStatus, toStatus domain.Status, reason string) error {
	return nil
}

func (r *mockStatsOrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	for _, o := range r.orders {
		if o.ID() == id {
			return o, nil
		}
	}
	return nil, domain.ErrOrderNotFound
}

func (r *mockStatsOrderRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Order, int, error) {
	var userOrders []*domain.Order
	for _, o := range r.orders {
		if o.UserID() == userID {
			userOrders = append(userOrders, o)
		}
	}
	return userOrders, len(userOrders), nil
}

func (r *mockStatsOrderRepo) HasActivePendingOrder(ctx context.Context, userID uuid.UUID) (bool, error) {
	return false, nil
}

func (r *mockStatsOrderRepo) GetStatusHistory(ctx context.Context, orderID uuid.UUID) ([]port.StatusHistoryRecord, error) {
	return nil, nil
}

func (r *mockStatsOrderRepo) IsEventProcessed(ctx context.Context, eventID, consumerName string) (bool, error) {
	return false, nil
}

func (r *mockStatsOrderRepo) MarkEventProcessed(ctx context.Context, eventID, consumerName string) error {
	return nil
}

func TestOrder_StreakCalculation(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	pp := domain.PickupPointInfo{
		ID:             uuid.New(),
		Name:           "ПВЗ Тест",
		Latitude:       55.75,
		Longitude:      37.61,
		DistanceMeters: 100,
	}

	t.Run("zero orders", func(t *testing.T) {
		repo := &mockStatsOrderRepo{}
		uc := usecase.NewOrderUseCase(repo, nil, nil, nil, nil)

		stats, err := uc.GetUserStats(context.Background(), userID, "UTC")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stats.TotalOrders != 0 || stats.CurrentStreakDays != 0 {
			t.Errorf("expected 0 stats, got %+v", stats)
		}
	})

	t.Run("consecutive orders today and yesterday", func(t *testing.T) {
		now := time.Now().UTC()
		yesterday := now.AddDate(0, 0, -1)
		twoDaysAgo := now.AddDate(0, 0, -2)

		o1 := domain.ReconstituteOrder(
			uuid.New(), userID, domain.StatusDelivered, decimal.NewFromFloat(10.00), pp, nil, twoDaysAgo, twoDaysAgo,
		)
		o2 := domain.ReconstituteOrder(
			uuid.New(), userID, domain.StatusDelivered, decimal.NewFromFloat(10.00), pp, nil, yesterday, yesterday,
		)
		o3 := domain.ReconstituteOrder(
			uuid.New(), userID, domain.StatusDelivered, decimal.NewFromFloat(10.00), pp, nil, now, now,
		)

		repo := &mockStatsOrderRepo{orders: []*domain.Order{o1, o2, o3}}
		uc := usecase.NewOrderUseCase(repo, nil, nil, nil, nil)

		stats, err := uc.GetUserStats(context.Background(), userID, "UTC")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stats.TotalOrders != 3 {
			t.Errorf("expected 3 orders, got %d", stats.TotalOrders)
		}
		if stats.CurrentStreakDays != 3 {
			t.Errorf("expected streak 3 days, got %d", stats.CurrentStreakDays)
		}
	})

	t.Run("broken streak", func(t *testing.T) {
		now := time.Now().UTC()
		fiveDaysAgo := now.AddDate(0, 0, -5)

		o1 := domain.ReconstituteOrder(
			uuid.New(), userID, domain.StatusDelivered, decimal.NewFromFloat(10.00), pp, nil, fiveDaysAgo, fiveDaysAgo,
		)

		repo := &mockStatsOrderRepo{orders: []*domain.Order{o1}}
		uc := usecase.NewOrderUseCase(repo, nil, nil, nil, nil)

		stats, err := uc.GetUserStats(context.Background(), userID, "UTC")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stats.TotalOrders != 1 {
			t.Errorf("expected 1 order, got %d", stats.TotalOrders)
		}
		if stats.CurrentStreakDays != 0 {
			t.Errorf("expected broken streak 0, got %d", stats.CurrentStreakDays)
		}
	})
}
