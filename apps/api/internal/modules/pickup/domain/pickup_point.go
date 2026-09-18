package domain

import (
	"time"

	"github.com/google/uuid"
)

type PickupPoint struct {
	id             uuid.UUID
	userID         uuid.UUID
	name           string
	location       LatLng
	distanceMeters float64
	createdAt      time.Time
}

func NewPickupPoint(
	id uuid.UUID,
	userID uuid.UUID,
	name string,
	location LatLng,
	distanceMeters float64,
	createdAt time.Time,
) *PickupPoint {
	if id == uuid.Nil {
		id = uuid.New()
	}
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	return &PickupPoint{
		id:             id,
		userID:         userID,
		name:           name,
		location:       location,
		distanceMeters: distanceMeters,
		createdAt:      createdAt,
	}
}

func (p *PickupPoint) ID() uuid.UUID {
	return p.id
}

func (p *PickupPoint) UserID() uuid.UUID {
	return p.userID
}

func (p *PickupPoint) Name() string {
	return p.name
}

func (p *PickupPoint) Location() LatLng {
	return p.location
}

func (p *PickupPoint) DistanceMeters() float64 {
	return p.distanceMeters
}

func (p *PickupPoint) CreatedAt() time.Time {
	return p.createdAt
}
