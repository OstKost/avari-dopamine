package pickup

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/pickup/adapter/httpapi"
	"github.com/ostkost/dopamine-market/api/internal/modules/pickup/adapter/postgres"
	"github.com/ostkost/dopamine-market/api/internal/modules/pickup/port"
	"github.com/ostkost/dopamine-market/api/internal/modules/pickup/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/random"
)

type Module struct {
	pickupUC *usecase.PickupUseCase
	handler  *httpapi.Handler
	repo     port.PickupRepository
}

func NewModule(dbPool *pgxpool.Pool, rnd random.Source) *Module {
	repo := postgres.NewPickupRepository(dbPool)
	pickupUC := usecase.NewPickupUseCase(repo, rnd)
	handler := httpapi.NewHandler(pickupUC)

	return &Module{
		pickupUC: pickupUC,
		handler:  handler,
		repo:     repo,
	}
}

func (m *Module) Routes() chi.Router {
	return m.handler.Routes()
}

func (m *Module) Repository() port.PickupRepository {
	return m.repo
}

// Реализация contracts.PickupPointLookup

func (m *Module) GetByID(ctx context.Context, id uuid.UUID) (contracts.PickupPointSnapshot, error) {
	p, err := m.pickupUC.GetByID(ctx, id)
	if err != nil {
		return contracts.PickupPointSnapshot{}, fmt.Errorf("getting pickup point: %w", err)
	}
	return contracts.PickupPointSnapshot{
		ID:             p.ID(),
		UserID:         p.UserID(),
		Name:           p.Name(),
		Latitude:       p.Location().Latitude,
		Longitude:      p.Location().Longitude,
		DistanceMeters: p.DistanceMeters(),
	}, nil
}

func (m *Module) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := m.pickupUC.GetByID(ctx, id)
	if err != nil {
		return false, nil
	}
	return true, nil
}
