package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/catalog/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/catalog/port"
)

type CatalogUseCase struct {
	repo port.CatalogRepository
}

func NewCatalogUseCase(repo port.CatalogRepository) *CatalogUseCase {
	return &CatalogUseCase{repo: repo}
}

func (uc *CatalogUseCase) ListCategories(ctx context.Context) ([]*domain.Category, error) {
	return uc.repo.ListCategories(ctx)
}

func (uc *CatalogUseCase) ListProducts(ctx context.Context, categoryID *uuid.UUID, query string, limit, offset int) ([]*domain.Product, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery != "" {
		return uc.repo.SearchProducts(ctx, trimmedQuery, limit, offset)
	}

	return uc.repo.ListProducts(ctx, categoryID, limit, offset)
}

func (uc *CatalogUseCase) GetProductByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	return uc.repo.GetProductByID(ctx, id)
}

func (uc *CatalogUseCase) GetProductsByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.Product, error) {
	if len(ids) == 0 {
		return make(map[uuid.UUID]*domain.Product), nil
	}
	return uc.repo.GetProductsByIDs(ctx, ids)
}
