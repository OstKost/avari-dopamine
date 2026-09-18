package domain

import "errors"

var (
	ErrInvalidQuantity     = errors.New("quantity must be greater than 0")
	ErrCartEmpty           = errors.New("cart is empty")
	ErrItemNotFound        = errors.New("item not found in cart")
	ErrProductNotFound     = errors.New("product not found in catalog")
	ErrPickupPointNotFound = errors.New("pickup point not found")
)
