package domain

import "errors"

var (
	ErrPickupPointNotFound = errors.New("pickup point not found")
	ErrInvalidCoordinates  = errors.New("invalid coordinates: latitude must be in [-90, 90], longitude in [-180, 180]")
)
