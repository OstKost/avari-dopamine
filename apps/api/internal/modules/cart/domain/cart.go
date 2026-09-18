package domain

import (
	"github.com/google/uuid"
)

// Cart — иммутабельный value object корзины (ADR-008).
// Все мутирующие методы возвращают новую копию.
type Cart struct {
	items         map[uuid.UUID]int
	pickupPointID *uuid.UUID
}

// NewCart создаёт пустую или наполненную корзину.
func NewCart(items map[uuid.UUID]int, pickupPointID *uuid.UUID) Cart {
	clonedItems := make(map[uuid.UUID]int, len(items))
	for k, v := range items {
		if v > 0 {
			clonedItems[k] = v
		}
	}

	var pointCopy *uuid.UUID
	if pickupPointID != nil {
		id := *pickupPointID
		pointCopy = &id
	}

	return Cart{
		items:         clonedItems,
		pickupPointID: pointCopy,
	}
}

func (c Cart) Items() map[uuid.UUID]int {
	cloned := make(map[uuid.UUID]int, len(c.items))
	for k, v := range c.items {
		cloned[k] = v
	}
	return cloned
}

func (c Cart) PickupPointID() *uuid.UUID {
	if c.pickupPointID == nil {
		return nil
	}
	id := *c.pickupPointID
	return &id
}

func (c Cart) IsEmpty() bool {
	return len(c.items) == 0
}

func (c Cart) TotalQuantity() int {
	total := 0
	for _, q := range c.items {
		total += q
	}
	return total
}

func (c Cart) WithItem(productID uuid.UUID, addQuantity int) (Cart, error) {
	if addQuantity <= 0 {
		return c, ErrInvalidQuantity
	}

	newItems := c.Items()
	newItems[productID] += addQuantity

	return Cart{
		items:         newItems,
		pickupPointID: c.PickupPointID(),
	}, nil
}

func (c Cart) WithQuantity(productID uuid.UUID, quantity int) (Cart, error) {
	if quantity < 0 {
		return c, ErrInvalidQuantity
	}

	newItems := c.Items()
	if quantity == 0 {
		delete(newItems, productID)
	} else {
		newItems[productID] = quantity
	}

	return Cart{
		items:         newItems,
		pickupPointID: c.PickupPointID(),
	}, nil
}

func (c Cart) WithoutItem(productID uuid.UUID) Cart {
	newItems := c.Items()
	delete(newItems, productID)

	return Cart{
		items:         newItems,
		pickupPointID: c.PickupPointID(),
	}
}

func (c Cart) WithPickupPoint(pointID *uuid.UUID) Cart {
	var newPoint *uuid.UUID
	if pointID != nil {
		id := *pointID
		newPoint = &id
	}

	return Cart{
		items:         c.Items(),
		pickupPointID: newPoint,
	}
}
