package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/pickup/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/pickup/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/random"
)

type mockPickupRepo struct {
	points map[uuid.UUID][]*domain.PickupPoint
}

func newMockPickupRepo() *mockPickupRepo {
	return &mockPickupRepo{
		points: make(map[uuid.UUID][]*domain.PickupPoint),
	}
}

func (m *mockPickupRepo) SavePickupPoints(ctx context.Context, points []*domain.PickupPoint) error {
	for _, p := range points {
		m.points[p.UserID()] = append(m.points[p.UserID()], p)
	}
	return nil
}

func (m *mockPickupRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.PickupPoint, error) {
	return m.points[userID], nil
}

func (m *mockPickupRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PickupPoint, error) {
	for _, list := range m.points {
		for _, p := range list {
			if p.ID() == id {
				return p, nil
			}
		}
	}
	return nil, domain.ErrPickupPointNotFound
}

func (m *mockPickupRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	delete(m.points, userID)
	return nil
}

func TestPickupUseCase_PropertyBasedGeneration(t *testing.T) {
	rnd := random.New(42) // детерминированный генератор
	repo := newMockPickupRepo()
	uc := usecase.NewPickupUseCase(repo, rnd)
	ctx := context.Background()

	origins := []domain.LatLng{
		{Latitude: 47.2357, Longitude: 39.7015}, // Rostov-on-Don
		{Latitude: 55.7558, Longitude: 37.6173}, // Moscow
		{Latitude: 59.9343, Longitude: 30.3351}, // Saint Petersburg
		{Latitude: -33.8688, Longitude: 151.2093}, // Sydney
	}

	// 1000 прогонов генерации (property-based тест по INV-03 и FR-PICKUP-01)
	for i := 0; i < 1000; i++ {
		origin := origins[i%len(origins)]
		userID := uuid.New()

		points, err := uc.GeneratePickupPoints(ctx, userID, origin)
		if err != nil {
			t.Fatalf("iteration %d: failed to generate points: %v", i, err)
		}

		// FR-PICKUP-01: 5-8 точек
		if len(points) < 5 || len(points) > 8 {
			t.Fatalf("iteration %d: expected 5-8 points, got %d", i, len(points))
		}

		for _, p := range points {
			// INV-03: расстояние строго в [100.0, 500.0] метров
			if p.DistanceMeters() < 100.0 || p.DistanceMeters() > 500.0 {
				t.Fatalf("iteration %d: point %s distance %.2f out of bounds [100, 500]",
					i, p.ID(), p.DistanceMeters())
			}

			// Геодезическая проверка реального расстояния
			actualDist := domain.DistanceBetween(origin, p.Location())
			if actualDist < 95.0 || actualDist > 505.0 {
				t.Fatalf("iteration %d: actual calculated distance %.2f violates bounds [100, 500]",
					i, actualDist)
			}
		}
	}
}
