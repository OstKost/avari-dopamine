package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusAssembling        Status = "assembling"
	StatusCourierAssigned   Status = "courier_assigned"
	StatusInTransit         Status = "in_transit"
	StatusDeliveryDelayed   Status = "delivery_delayed"
	StatusDelivered         Status = "delivered"
)

// AllowedTransitions — допустимые переходы стейт-машины доставки (FR-DELIVERY-01, ADR-005).
var AllowedTransitions = map[Status][]Status{
	StatusAssembling: {
		StatusCourierAssigned,
	},
	StatusCourierAssigned: {
		StatusInTransit,
	},
	StatusInTransit: {
		StatusDeliveryDelayed,
		StatusDelivered,
	},
	StatusDeliveryDelayed: {
		StatusDelivered,
	},
	StatusDelivered: {},
}

// Delivery — доменный агрегат симуляции доставки (ADR-005, ADR-012).
type Delivery struct {
	ID                    uuid.UUID
	OrderID               uuid.UUID
	Status                Status
	CourierName           string
	CourierRating         float64
	StartedAt             time.Time
	EstimatedCompletionAt time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// NewDelivery инициализирует новую доставку в статусе assembling.
func NewDelivery(
	id uuid.UUID,
	orderID uuid.UUID,
	courierName string,
	courierRating float64,
	startedAt time.Time,
	estimatedDuration time.Duration,
) *Delivery {
	now := time.Now().UTC()
	if startedAt.IsZero() {
		startedAt = now
	}

	return &Delivery{
		ID:                    id,
		OrderID:               orderID,
		Status:                StatusAssembling,
		CourierName:           courierName,
		CourierRating:         courierRating,
		StartedAt:             startedAt,
		EstimatedCompletionAt: startedAt.Add(estimatedDuration),
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

// TransitionTo переводит доставку в новый статус с валидацией стейт-машины (FR-DELIVERY-01).
func (d *Delivery) TransitionTo(newStatus Status) error {
	if d.Status == newStatus {
		return nil
	}

	allowed, exists := AllowedTransitions[d.Status]
	if !exists {
		return fmt.Errorf("%w: current status %q has no transitions", ErrInvalidTransition, d.Status)
	}

	isAllowed := false
	for _, s := range allowed {
		if s == newStatus {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return fmt.Errorf("%w: cannot transition from %q to %q", ErrInvalidTransition, d.Status, newStatus)
	}

	d.Status = newStatus
	d.UpdatedAt = time.Now().UTC()
	return nil
}

// IsTerminal возвращает true, если доставка завершена.
func (d *Delivery) IsTerminal() bool {
	return d.Status == StatusDelivered
}
