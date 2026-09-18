package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ostkost/dopamine-market/api/internal/modules/catalog/domain"
	"github.com/shopspring/decimal"
)

type CatalogRepository struct {
	pool *pgxpool.Pool
}

func NewCatalogRepository(pool *pgxpool.Pool) *CatalogRepository {
	return &CatalogRepository{pool: pool}
}

func (r *CatalogRepository) ListCategories(ctx context.Context) ([]*domain.Category, error) {
	query := `
		SELECT id, name, slug, description, created_at
		FROM catalog.categories
		ORDER BY name ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing categories: %w", err)
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		var (
			id          uuid.UUID
			name        string
			slug        string
			description string
			createdAt   time.Time
		)
		if err := rows.Scan(&id, &name, &slug, &description, &createdAt); err != nil {
			return nil, fmt.Errorf("scanning category: %w", err)
		}
		categories = append(categories, domain.NewCategory(id, name, slug, description, createdAt))
	}
	return categories, rows.Err()
}

func (r *CatalogRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	query := `
		SELECT id, name, slug, description, created_at
		FROM catalog.categories
		WHERE id = $1
	`
	var (
		name        string
		slug        string
		description string
		createdAt   time.Time
	)
	err := r.pool.QueryRow(ctx, query, id).Scan(&id, &name, &slug, &description, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("getting category: %w", err)
	}
	return domain.NewCategory(id, name, slug, description, createdAt), nil
}

func (r *CatalogRepository) ListProducts(ctx context.Context, categoryID *uuid.UUID, limit, offset int) ([]*domain.Product, int, error) {
	countQuery := `SELECT COUNT(*) FROM catalog.products`
	query := `
		SELECT p.id, p.category_id, c.name, p.name, p.description, p.price_rub, p.image_seed, p.created_at
		FROM catalog.products p
		JOIN catalog.categories c ON p.category_id = c.id
	`
	var args []interface{}
	var countArgs []interface{}

	if categoryID != nil {
		countQuery += ` WHERE category_id = $1`
		query += ` WHERE p.category_id = $1`
		args = append(args, *categoryID)
		countArgs = append(countArgs, *categoryID)
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting products: %w", err)
	}

	query += fmt.Sprintf(` ORDER BY p.created_at DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying products: %w", err)
	}
	defer rows.Close()

	products, err := r.scanProducts(rows)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *CatalogRepository) SearchProducts(ctx context.Context, queryStr string, limit, offset int) ([]*domain.Product, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM catalog.products
		WHERE tsv @@ plainto_tsquery('russian', $1)
	`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, queryStr).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting search results: %w", err)
	}

	query := `
		SELECT p.id, p.category_id, c.name, p.name, p.description, p.price_rub, p.image_seed, p.created_at
		FROM catalog.products p
		JOIN catalog.categories c ON p.category_id = c.id
		WHERE p.tsv @@ plainto_tsquery('russian', $1)
		ORDER BY ts_rank(p.tsv, plainto_tsquery('russian', $1)) DESC, p.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, queryStr, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("searching products: %w", err)
	}
	defer rows.Close()

	products, err := r.scanProducts(rows)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *CatalogRepository) GetProductByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	query := `
		SELECT p.id, p.category_id, c.name, p.name, p.description, p.price_rub, p.image_seed, p.created_at
		FROM catalog.products p
		JOIN catalog.categories c ON p.category_id = c.id
		WHERE p.id = $1
	`
	var (
		categoryID   uuid.UUID
		categoryName string
		name         string
		description  string
		priceRUB     decimal.Decimal
		imageSeed    string
		createdAt    time.Time
	)
	err := r.pool.QueryRow(ctx, query, id).Scan(&id, &categoryID, &categoryName, &name, &description, &priceRUB, &imageSeed, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("getting product: %w", err)
	}
	return domain.NewProduct(id, categoryID, categoryName, name, description, priceRUB, imageSeed, createdAt), nil
}

func (r *CatalogRepository) GetProductsByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.Product, error) {
	if len(ids) == 0 {
		return make(map[uuid.UUID]*domain.Product), nil
	}

	query := `
		SELECT p.id, p.category_id, c.name, p.name, p.description, p.price_rub, p.image_seed, p.created_at
		FROM catalog.products p
		JOIN catalog.categories c ON p.category_id = c.id
		WHERE p.id = ANY($1)
	`
	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("querying products by ids: %w", err)
	}
	defer rows.Close()

	products, err := r.scanProducts(rows)
	if err != nil {
		return nil, err
	}

	res := make(map[uuid.UUID]*domain.Product, len(products))
	for _, p := range products {
		res[p.ID()] = p
	}
	return res, nil
}

func (r *CatalogRepository) InsertCategory(ctx context.Context, cat *domain.Category) error {
	query := `
		INSERT INTO catalog.categories (id, name, slug, description, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, cat.ID(), cat.Name(), cat.Slug(), cat.Description(), cat.CreatedAt())
	if err != nil {
		return fmt.Errorf("inserting category: %w", err)
	}
	return nil
}

func (r *CatalogRepository) InsertProduct(ctx context.Context, prod *domain.Product) error {
	query := `
		INSERT INTO catalog.products (id, category_id, name, description, price_rub, image_seed, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, prod.ID(), prod.CategoryID(), prod.Name(), prod.Description(), prod.PriceRUB(), prod.ImageSeed(), prod.CreatedAt())
	if err != nil {
		return fmt.Errorf("inserting product: %w", err)
	}
	return nil
}

func (r *CatalogRepository) scanProducts(rows pgx.Rows) ([]*domain.Product, error) {
	var products []*domain.Product
	for rows.Next() {
		var (
			id           uuid.UUID
			categoryID   uuid.UUID
			categoryName string
			name         string
			description  string
			priceRUB     decimal.Decimal
			imageSeed    string
			createdAt    time.Time
		)
		if err := rows.Scan(&id, &categoryID, &categoryName, &name, &description, &priceRUB, &imageSeed, &createdAt); err != nil {
			return nil, fmt.Errorf("scanning product: %w", err)
		}
		products = append(products, domain.NewProduct(id, categoryID, categoryName, name, description, priceRUB, imageSeed, createdAt))
	}
	return products, rows.Err()
}
