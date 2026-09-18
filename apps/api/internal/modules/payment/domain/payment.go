package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

const FixedPaymentAmountRUB = "10.00"

type Payment struct {
	id                uuid.UUID
	orderID           uuid.UUID
	provider          string
	providerPaymentID string
	amountRUB         string
	status            Status
	confirmationURL   string
	failureReason     string
	createdAt         time.Time
	updatedAt         time.Time
}

func NewPayment(
	id uuid.UUID,
	orderID uuid.UUID,
	provider string,
	providerPaymentID string,
	confirmationURL string,
) (*Payment, error) {
	if id == uuid.Nil {
		id = uuid.New()
	}
	if orderID == uuid.Nil {
		return nil, fmt.Errorf("order ID is required")
	}
	if provider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	if providerPaymentID == "" {
		return nil, fmt.Errorf("provider payment ID is required")
	}

	now := time.Now().UTC()
	return &Payment{
		id:                id,
		orderID:           orderID,
		provider:          provider,
		providerPaymentID: providerPaymentID,
		amountRUB:         FixedPaymentAmountRUB, // INV-01
		status:            StatusPending,
		confirmationURL:   confirmationURL,
		createdAt:         now,
		updatedAt:         now,
	}, nil
}

func ReconstitutePayment(
	id uuid.UUID,
	orderID uuid.UUID,
	provider string,
	providerPaymentID string,
	amountRUB string,
	status Status,
	confirmationURL string,
	failureReason string,
	createdAt time.Time,
	updatedAt time.Time,
) *Payment {
	return &Payment{
		id:                id,
		orderID:           orderID,
		provider:          provider,
		providerPaymentID: providerPaymentID,
		amountRUB:         amountRUB,
		status:            status,
		confirmationURL:   confirmationURL,
		failureReason:     failureReason,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
	}
}

func (p *Payment) MarkSucceeded() error {
	if p.status == StatusSucceeded {
		return nil // Идемпотентно
	}
	if p.status == StatusFailed {
		return ErrInvalidStatusTransition
	}
	p.status = StatusSucceeded
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Payment) MarkFailed(reason string) error {
	if p.status == StatusFailed {
		return nil // Идемпотентно
	}
	if p.status == StatusSucceeded {
		return ErrInvalidStatusTransition
	}
	p.status = StatusFailed
	p.failureReason = reason
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Payment) ID() uuid.UUID                { return p.id }
func (p *Payment) OrderID() uuid.UUID           { return p.orderID }
func (p *Payment) Provider() string             { return p.provider }
func (p *Payment) ProviderPaymentID() string     { return p.providerPaymentID }
func (p *Payment) AmountRUB() string            { return p.amountRUB }
func (p *Payment) Status() Status               { return p.status }
func (p *Payment) ConfirmationURL() string       { return p.confirmationURL }
func (p *Payment) FailureReason() string        { return p.failureReason }
func (p *Payment) CreatedAt() time.Time         { return p.createdAt }
func (p *Payment) UpdatedAt() time.Time         { return p.updatedAt }
