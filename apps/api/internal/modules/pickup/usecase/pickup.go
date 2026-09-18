package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/pickup/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/pickup/port"
	"github.com/ostkost/dopamine-market/api/internal/platform/random"
)

// Rostov-on-Don fallback coordinates (FR-PICKUP-01).
var DefaultFallbackLocation = domain.LatLng{
	Latitude:  47.2357,
	Longitude: 39.7015,
}

var landmarks = []string{
	"кофейни 'Бодрость'",
	"сквера с лавочками",
	"фонтана",
	"книжного магазина",
	"пекарни",
	"аптеки",
	"торгового центра",
	"парковки",
	"детской площадки",
	"цветочной лавки",
	"арт-пространства",
	"бизнес-центра 'Горизонт'",
	"остановки транспорта",
	"городского парка",
	"супермаркета",
}

type PickupUseCase struct {
	repo port.PickupRepository
	rnd  random.Source
}

func NewPickupUseCase(repo port.PickupRepository, rnd random.Source) *PickupUseCase {
	return &PickupUseCase{
		repo: repo,
		rnd:  rnd,
	}
}

func (uc *PickupUseCase) GeneratePickupPoints(ctx context.Context, userID uuid.UUID, origin domain.LatLng) ([]*domain.PickupPoint, error) {
	if origin.Latitude == 0 && origin.Longitude == 0 {
		origin = DefaultFallbackLocation
	}

	// 5-8 точек за вызов (FR-PICKUP-01)
	count := uc.rnd.IntRange(5, 8)
	points := make([]*domain.PickupPoint, 0, count)

	now := time.Now().UTC()

	for i := 1; i <= count; i++ {
		// Расстояние строго в [100.0, 500.0] метров (INV-03)
		distMeters := uc.rnd.Float64Range(100.0, 500.0)
		// Округляем до десятых
		distMeters = math.Round(distMeters*10) / 10

		bearing := uc.rnd.Float64Range(0.0, 360.0)
		loc := domain.DestinationPoint(origin, bearing, distMeters)

		landmark := uc.rnd.Pick(landmarks)
		name := fmt.Sprintf("Пункт выдачи №%d (у %s)", i, landmark)

		point := domain.NewPickupPoint(
			uuid.New(),
			userID,
			name,
			loc,
			distMeters,
			now,
		)
		points = append(points, point)
	}

	// Очищаем предыдущие точки пользователя и сохраняем свежесгенерированные
	_ = uc.repo.DeleteByUserID(ctx, userID)
	if err := uc.repo.SavePickupPoints(ctx, points); err != nil {
		return nil, fmt.Errorf("saving generated pickup points: %w", err)
	}

	return points, nil
}

func (uc *PickupUseCase) ListPickupPoints(ctx context.Context, userID uuid.UUID) ([]*domain.PickupPoint, error) {
	points, err := uc.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing pickup points: %w", err)
	}

	// Если у пользователя еще нет сгенерированных точек, генерируем с дефолтными координатами Ростова-на-Дону
	if len(points) == 0 {
		return uc.GeneratePickupPoints(ctx, userID, DefaultFallbackLocation)
	}

	return points, nil
}

func (uc *PickupUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.PickupPoint, error) {
	return uc.repo.GetByID(ctx, id)
}
