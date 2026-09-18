package domain

import "errors"

var (
	ErrCartEmpty                = errors.New("cart is empty, cannot create order")
	ErrInvalidStatusTransition  = errors.New("invalid order status transition")
	ErrDuplicatePendingOrder    = errors.New("user already has an active order pending payment (INV-02)")
	ErrOrderNotFound            = errors.New("order not found")
	ErrOrderNotModifiable       = errors.New("order cannot be modified after payment (INV-04)")
	ErrPickupPointRequired      = errors.New("pickup point must be selected before checkout")
)
