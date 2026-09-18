package contracts

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PaymentStatus перечисление статусов платежа.
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
)

// PaymentSnapshot представляет снимок платежа для межмодульного взаимодействия.
type PaymentSnapshot struct {
	ID                uuid.UUID     `json:"id"`
	OrderID           uuid.UUID     `json:"order_id"`
	Provider          string        `json:"provider"`
	ProviderPaymentID string        `json:"provider_payment_id"`
	AmountRUB         string        `json:"amount_rub"` // Всегда "10.00" согласно INV-01
	Status            PaymentStatus `json:"status"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

// InitiatePaymentRequest запрос на инициацию платежа в провайдере.
type InitiatePaymentRequest struct {
	OrderID     uuid.UUID `json:"order_id"`
	UserID      uuid.UUID `json:"user_id"`
	AmountRUB   string    `json:"amount_rub"` // Всегда "10.00"
	Description string    `json:"description"`
	ReturnURL   string    `json:"return_url"`
}

// InitiatePaymentResult ответ провайдера на инициацию платежа.
type InitiatePaymentResult struct {
	ProviderPaymentID string        `json:"provider_payment_id"`
	ConfirmationURL   string        `json:"confirmation_url,omitempty"`
	Status            PaymentStatus `json:"status"`
}

// WebhookResult нормализованный результат обработки вебхука от платежного шлюза.
type WebhookResult struct {
	ProviderPaymentID string        `json:"provider_payment_id"`
	OrderID           uuid.UUID     `json:"order_id"`
	Status            PaymentStatus `json:"status"`
	FailureReason     string        `json:"failure_reason,omitempty"`
}

// PaymentProvider абстракция платежного провайдера (ADR-006).
type PaymentProvider interface {
	Name() string
	InitiatePayment(ctx context.Context, req InitiatePaymentRequest) (InitiatePaymentResult, error)
	VerifyWebhook(ctx context.Context, rawBody []byte, headers map[string]string) (WebhookResult, error)
}

// PaymentLookup межмодульный интерфейс для чтения данных о платежах.
type PaymentLookup interface {
	GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*PaymentSnapshot, error)
}
