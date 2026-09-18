package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/catalog/domain"
)

type CatalogRepository interface {
	ListCategories(ctx context.Context) ([]*domain.Category, error)
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	ListProducts(ctx context.Context, categoryID *uuid.UUID, limit, offset int) ([]*domain.Product, int, error)
	SearchProducts(ctx context.Context, query string, limit, offset int) ([]*domain.Product, int, error)
	GetProductByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	GetProductsByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.Product, error)
	InsertCategory(ctx context.Context, cat *domain.Category) error
	InsertProduct(ctx context.Context, prod *domain.Product) error
}
