package contracts

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductSnapshot — снимок товара для использования другими модулями (order, cart).
// Денормализация цен и названий согласно ADR-011.
type ProductSnapshot struct {
	ID          uuid.UUID       `json:"id"`
	CategoryID  uuid.UUID       `json:"category_id"`
	CategoryName string         `json:"category_name"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	PriceRUB    decimal.Decimal `json:"price_rub"`
	ImageSeed   string          `json:"image_seed"`
}

// ProductLookup — межмодульный интерфейс каталога товаров.
type ProductLookup interface {
	GetByID(ctx context.Context, id uuid.UUID) (ProductSnapshot, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]ProductSnapshot, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}
