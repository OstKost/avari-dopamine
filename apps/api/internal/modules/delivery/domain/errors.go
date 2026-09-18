package domain

import "errors"

var (
	ErrDeliveryNotFound  = errors.New("delivery not found")
	ErrInvalidTransition = errors.New("invalid delivery status transition")
	ErrDeliveryExists    = errors.New("delivery for order already exists")
)
