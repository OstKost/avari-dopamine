package domain

import "errors"

var (
	ErrForbidden     = errors.New("forbidden: access to order is not allowed")
	ErrOrderNotFound = errors.New("order not found")
)
