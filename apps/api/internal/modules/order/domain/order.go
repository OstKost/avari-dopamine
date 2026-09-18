package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// FixedOrderAmountRUB — инвариант INV-01: сумма любого заказа = 10.00 RUB всегда.
var FixedOrderAmountRUB = decimal.NewFromInt(10)

type PickupPointInfo struct {
	ID             uuid.UUID
	Name           string
	Latitude       float64
	Longitude      float64
	DistanceMeters float64
}

// Order — агрегат заказа (центральный агрегат системы, ADR-001 / ADR-011).
type Order struct {
	id             uuid.UUID
	userID         uuid.UUID
	status         Status
	totalAmountRUB decimal.Decimal
	pickupPoint    PickupPointInfo
	items          []OrderItem
	createdAt      time.Time
	updatedAt      time.Time
}

func NewOrder(
	id uuid.UUID,
	userID uuid.UUID,
	pickupPoint PickupPointInfo,
	items []OrderItem,
	createdAt time.Time,
) (*Order, error) {
	if id == uuid.Nil {
		id = uuid.New()
	}
	if len(items) == 0 {
		return nil, ErrCartEmpty
	}
	if pickupPoint.ID == uuid.Nil {
		return nil, ErrPickupPointRequired
	}
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return &Order{
		id:             id,
		userID:         userID,
		status:         StatusCreated,
		totalAmountRUB: FixedOrderAmountRUB, // INV-01
		pickupPoint:    pickupPoint,
		items:          items,
		createdAt:      createdAt,
		updatedAt:      createdAt,
	}, nil
}

// ReconstituteOrder воссоздаёт существующий агрегат из БД.
func ReconstituteOrder(
	id uuid.UUID,
	userID uuid.UUID,
	status Status,
	totalAmountRUB decimal.Decimal,
	pickupPoint PickupPointInfo,
	items []OrderItem,
	createdAt time.Time,
	updatedAt time.Time,
) *Order {
	return &Order{
		id:             id,
		userID:         userID,
		status:         status,
		totalAmountRUB: totalAmountRUB,
		pickupPoint:    pickupPoint,
		items:          items,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

func (o *Order) ID() uuid.UUID {
	return o.id
}

func (o *Order) UserID() uuid.UUID {
	return o.userID
}

func (o *Order) Status() Status {
	return o.status
}

func (o *Order) TotalAmountRUB() decimal.Decimal {
	return o.totalAmountRUB
}

func (o *Order) PickupPoint() PickupPointInfo {
	return o.pickupPoint
}

func (o *Order) Items() []OrderItem {
	cloned := make([]OrderItem, len(o.items))
	copy(cloned, o.items)
	return cloned
}

func (o *Order) CreatedAt() time.Time {
	return o.createdAt
}

func (o *Order) UpdatedAt() time.Time {
	return o.updatedAt
}

// TransitionTo переводит заказ в новый статус согласно стейт-машине (FR-ORDER-02).
func (o *Order) TransitionTo(target Status) error {
	if !o.status.CanTransitionTo(target) {
		return ErrInvalidStatusTransition
	}

	o.status = target
	o.updatedAt = time.Now().UTC()
	return nil
}
