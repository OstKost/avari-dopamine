package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/catalog/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/catalog/usecase"
	"github.com/shopspring/decimal"
)

type mockCatalogRepo struct {
	categories []*domain.Category
	products   map[uuid.UUID]*domain.Product
}

func newMockCatalogRepo() *mockCatalogRepo {
	return &mockCatalogRepo{
		categories: nil,
		products:   make(map[uuid.UUID]*domain.Product),
	}
}

func (m *mockCatalogRepo) ListCategories(ctx context.Context) ([]*domain.Category, error) {
	return m.categories, nil
}

func (m *mockCatalogRepo) GetCategoryByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	for _, c := range m.categories {
		if c.ID() == id {
			return c, nil
		}
	}
	return nil, domain.ErrCategoryNotFound
}

func (m *mockCatalogRepo) ListProducts(ctx context.Context, categoryID *uuid.UUID, limit, offset int) ([]*domain.Product, int, error) {
	var list []*domain.Product
	for _, p := range m.products {
		if categoryID == nil || p.CategoryID() == *categoryID {
			list = append(list, p)
		}
	}
	return list, len(list), nil
}

func (m *mockCatalogRepo) SearchProducts(ctx context.Context, query string, limit, offset int) ([]*domain.Product, int, error) {
	var list []*domain.Product
	for _, p := range m.products {
		list = append(list, p)
	}
	return list, len(list), nil
}

func (m *mockCatalogRepo) GetProductByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	return p, nil
}

func (m *mockCatalogRepo) GetProductsByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.Product, error) {
	res := make(map[uuid.UUID]*domain.Product)
	for _, id := range ids {
		if p, ok := m.products[id]; ok {
			res[id] = p
		}
	}
	return res, nil
}

func (m *mockCatalogRepo) InsertCategory(ctx context.Context, cat *domain.Category) error {
	m.categories = append(m.categories, cat)
	return nil
}

func (m *mockCatalogRepo) InsertProduct(ctx context.Context, prod *domain.Product) error {
	m.products[prod.ID()] = prod
	return nil
}

func TestCatalogUseCase(t *testing.T) {
	repo := newMockCatalogRepo()
	uc := usecase.NewCatalogUseCase(repo)
	ctx := context.Background()

	catID := uuid.New()
	cat := domain.NewCategory(catID, "Антистресс", "antistress", "товары", time.Now())
	_ = repo.InsertCategory(ctx, cat)

	prodID := uuid.New()
	prod := domain.NewProduct(prodID, catID, cat.Name(), "Пузырчатая пленка", "Бесконечная", decimal.NewFromInt(10), "seed-1", time.Now())
	_ = repo.InsertProduct(ctx, prod)

	// List categories
	cats, err := uc.ListCategories(ctx)
	if err != nil || len(cats) != 1 {
		t.Fatalf("expected 1 category, got %v, err=%v", cats, err)
	}

	// Get product
	p, err := uc.GetProductByID(ctx, prodID)
	if err != nil || p.ID() != prodID {
		t.Fatalf("expected product %v, got %v, err=%v", prodID, p, err)
	}

	// Product Lookup GetByIDs
	res, err := uc.GetProductsByIDs(ctx, []uuid.UUID{prodID, uuid.New()})
	if err != nil {
		t.Fatalf("GetProductsByIDs failed: %v", err)
	}
	if len(res) != 1 || res[prodID] == nil {
		t.Fatalf("expected 1 matching product for partially existing IDs, got %v", res)
	}
}
