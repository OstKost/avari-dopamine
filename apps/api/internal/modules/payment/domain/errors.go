package domain

import "errors"

var (
	ErrPaymentNotFound          = errors.New("payment not found")
	ErrPaymentAlreadyProcessed  = errors.New("payment already processed (terminal status)")
	ErrInvalidPaymentAmount     = errors.New("invalid payment amount: must be strictly 10.00 RUB (INV-01)")
	ErrInvalidStatusTransition  = errors.New("invalid payment status transition")
	ErrInvalidWebhookSignature  = errors.New("invalid webhook signature or untrusted source")
	ErrProviderFailure          = errors.New("payment provider failure")
)
