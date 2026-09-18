package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/delivery/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/delivery/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/random"
)

type inMemoryDeliveryRepo struct {
	deliveries map[uuid.UUID]*domain.Delivery
	byOrder    map[uuid.UUID]*domain.Delivery
}

func newInMemoryRepo() *inMemoryDeliveryRepo {
	return &inMemoryDeliveryRepo{
		deliveries: make(map[uuid.UUID]*domain.Delivery),
		byOrder:    make(map[uuid.UUID]*domain.Delivery),
	}
}

func (r *inMemoryDeliveryRepo) Save(ctx context.Context, d *domain.Delivery) error {
	r.deliveries[d.ID] = d
	r.byOrder[d.OrderID] = d
	return nil
}

func (r *inMemoryDeliveryRepo) Update(ctx context.Context, d *domain.Delivery) error {
	r.deliveries[d.ID] = d
	r.byOrder[d.OrderID] = d
	return nil
}

func (r *inMemoryDeliveryRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
	d, ok := r.deliveries[id]
	if !ok {
		return nil, domain.ErrDeliveryNotFound
	}
	return d, nil
}

func (r *inMemoryDeliveryRepo) FindByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Delivery, error) {
	d, ok := r.byOrder[orderID]
	if !ok {
		return nil, domain.ErrDeliveryNotFound
	}
	return d, nil
}

func TestDelivery_GetDeliveryByOrderID(t *testing.T) {
	t.Parallel()

	repo := newInMemoryRepo()
	rnd := random.New(42)
	cfg := usecase.DefaultConfig()
	uc := usecase.NewDeliveryUseCase(repo, nil, rnd, cfg)

	ctx := context.Background()
	orderID := uuid.New()

	del := domain.NewDelivery(uuid.New(), orderID, "Курьер Иван", 4.85, time.Now(), 3*time.Minute)
	if err := repo.Save(ctx, del); err != nil {
		t.Fatalf("saving delivery: %v", err)
	}

	found, err := uc.GetDeliveryByOrderID(ctx, orderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil || found.ID != del.ID {
		t.Fatalf("expected delivery %v, got %v", del.ID, found)
	}
	if found.CourierName != "Курьер Иван" {
		t.Errorf("expected courier name 'Курьер Иван', got %s", found.CourierName)
	}
}

func TestDelivery_RandomIntervalProperties(t *testing.T) {
	t.Parallel()

	// Property-based test: проверяем, что все сгенерированные значения курьеров и рейтингов
	// лежат строго в заявленных диапазонах [4.2, 5.0] (FR-DELIVERY-01).
	rnd := random.New(42) // Fixed seed for reproducibility (ADR-012)

	for i := 0; i < 100; i++ {
		rating := rnd.Float64Range(4.2, 5.0)
		if rating < 4.2 || rating > 5.0 {
			t.Fatalf("iteration %d: rating %f out of range [4.2, 5.0]", i, rating)
		}

		assemblingDelay := rnd.IntRange(10, 30)
		if assemblingDelay < 10 || assemblingDelay > 30 {
			t.Fatalf("iteration %d: assembling delay %d out of range [10, 30]", i, assemblingDelay)
		}

		inTransitDelay := rnd.IntRange(60, 180)
		if assemblingDelay < 10 || inTransitDelay > 180 {
			t.Fatalf("iteration %d: in_transit delay %d out of range [60, 180]", i, inTransitDelay)
		}
	}
}

func TestDelivery_DelayedProbability(t *testing.T) {
	t.Parallel()

	// Проверяем статистическое распределение delayed ветки (10%)
	rnd := random.New(12345)
	trials := 1000
	delayedCount := 0

	for i := 0; i < trials; i++ {
		if rnd.Bool(0.10) {
			delayedCount++
		}
	}

	ratio := float64(delayedCount) / float64(trials)
	// Допускаем статистический разброс вокруг 10% (от 7% до 13%)
	if ratio < 0.07 || ratio > 0.13 {
		t.Fatalf("unexpected delayed ratio: %f (count %d out of %d)", ratio, delayedCount, trials)
	}
}
