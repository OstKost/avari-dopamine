package domain

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderItem struct {
	id           uuid.UUID
	productID    uuid.UUID
	name         string
	categoryName string
	priceRUB     decimal.Decimal
	imageSeed    string
	quantity     int
	subtotalRUB  decimal.Decimal
}

func NewOrderItem(
	id uuid.UUID,
	productID uuid.UUID,
	name string,
	categoryName string,
	priceRUB decimal.Decimal,
	imageSeed string,
	quantity int,
) OrderItem {
	if id == uuid.Nil {
		id = uuid.New()
	}
	if priceRUB.IsZero() {
		priceRUB = decimal.NewFromInt(10)
	}
	subtotal := priceRUB.Mul(decimal.NewFromInt(int64(quantity)))

	return OrderItem{
		id:           id,
		productID:    productID,
		name:         name,
		categoryName: categoryName,
		priceRUB:     priceRUB,
		imageSeed:    imageSeed,
		quantity:     quantity,
		subtotalRUB:  subtotal,
	}
}

func (i OrderItem) ID() uuid.UUID {
	return i.id
}

func (i OrderItem) ProductID() uuid.UUID {
	return i.productID
}

func (i OrderItem) Name() string {
	return i.name
}

func (i OrderItem) CategoryName() string {
	return i.categoryName
}

func (i OrderItem) PriceRUB() decimal.Decimal {
	return i.priceRUB
}

func (i OrderItem) ImageSeed() string {
	return i.imageSeed
}

func (i OrderItem) Quantity() int {
	return i.quantity
}

func (i OrderItem) SubtotalRUB() decimal.Decimal {
	return i.subtotalRUB
}
