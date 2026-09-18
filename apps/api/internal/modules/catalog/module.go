package catalog

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/catalog/adapter/httpapi"
	"github.com/ostkost/dopamine-market/api/internal/modules/catalog/adapter/postgres"
	"github.com/ostkost/dopamine-market/api/internal/modules/catalog/port"
	"github.com/ostkost/dopamine-market/api/internal/modules/catalog/usecase"
)

type Module struct {
	catalogUC *usecase.CatalogUseCase
	handler   *httpapi.Handler
	repo      port.CatalogRepository
}

func NewModule(dbPool *pgxpool.Pool) *Module {
	repo := postgres.NewCatalogRepository(dbPool)
	catalogUC := usecase.NewCatalogUseCase(repo)
	handler := httpapi.NewHandler(catalogUC)

	return &Module{
		catalogUC: catalogUC,
		handler:   handler,
		repo:      repo,
	}
}

func (m *Module) Routes() chi.Router {
	return m.handler.Routes()
}

func (m *Module) Repository() port.CatalogRepository {
	return m.repo
}

// Реализация contracts.ProductLookup

func (m *Module) GetByID(ctx context.Context, id uuid.UUID) (contracts.ProductSnapshot, error) {
	p, err := m.catalogUC.GetProductByID(ctx, id)
	if err != nil {
		return contracts.ProductSnapshot{}, fmt.Errorf("getting product: %w", err)
	}
	return contracts.ProductSnapshot{
		ID:           p.ID(),
		CategoryID:   p.CategoryID(),
		CategoryName: p.CategoryName(),
		Name:         p.Name(),
		Description:  p.Description(),
		PriceRUB:     p.PriceRUB(),
		ImageSeed:    p.ImageSeed(),
	}, nil
}

func (m *Module) GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]contracts.ProductSnapshot, error) {
	products, err := m.catalogUC.GetProductsByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("getting products by ids: %w", err)
	}

	result := make(map[uuid.UUID]contracts.ProductSnapshot, len(products))
	for id, p := range products {
		result[id] = contracts.ProductSnapshot{
			ID:           p.ID(),
			CategoryID:   p.CategoryID(),
			CategoryName: p.CategoryName(),
			Name:         p.Name(),
			Description:  p.Description(),
			PriceRUB:     p.PriceRUB(),
			ImageSeed:    p.ImageSeed(),
		}
	}
	return result, nil
}

func (m *Module) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := m.catalogUC.GetProductByID(ctx, id)
	if err != nil {
		return false, nil
	}
	return true, nil
}
