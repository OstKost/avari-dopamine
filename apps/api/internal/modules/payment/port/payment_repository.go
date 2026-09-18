package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/domain"
)

type PaymentRepository interface {
	Save(ctx context.Context, payment *domain.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Payment, error)
	GetByProviderPaymentID(ctx context.Context, providerPaymentID string) (*domain.Payment, error)
	Update(ctx context.Context, payment *domain.Payment) error
}
