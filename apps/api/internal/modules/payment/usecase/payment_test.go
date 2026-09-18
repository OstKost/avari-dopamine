package usecase_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/adapter/mock"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/random"
)

type memoryPaymentRepo struct {
	payments map[uuid.UUID]*domain.Payment
}

func newMemoryRepo() *memoryPaymentRepo {
	return &memoryPaymentRepo{payments: make(map[uuid.UUID]*domain.Payment)}
}

func (r *memoryPaymentRepo) Save(ctx context.Context, p *domain.Payment) error {
	r.payments[p.OrderID()] = p
	return nil
}

func (r *memoryPaymentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	for _, p := range r.payments {
		if p.ID() == id {
			return p, nil
		}
	}
	return nil, domain.ErrPaymentNotFound
}

func (r *memoryPaymentRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Payment, error) {
	if p, ok := r.payments[orderID]; ok {
		return p, nil
	}
	return nil, domain.ErrPaymentNotFound
}

func (r *memoryPaymentRepo) GetByProviderPaymentID(ctx context.Context, pid string) (*domain.Payment, error) {
	for _, p := range r.payments {
		if p.ProviderPaymentID() == pid {
			return p, nil
		}
	}
	return nil, domain.ErrPaymentNotFound
}

func (r *memoryPaymentRepo) Update(ctx context.Context, p *domain.Payment) error {
	r.payments[p.OrderID()] = p
	return nil
}

func TestPaymentUseCase_Flow(t *testing.T) {
	repo := newMemoryRepo()
	rnd := random.New(42)
	provider := mock.NewProvider(1.0, rnd, "")
	uc := usecase.NewPaymentUseCase(repo, provider, nil)

	ctx := context.Background()
	orderID := uuid.New()
	userID := uuid.New()

	// 1. Initiate Payment
	payment, err := uc.InitiatePayment(ctx, orderID, userID, "http://localhost/return")
	if err != nil {
		t.Fatalf("unexpected error initiating payment: %v", err)
	}

	if payment.AmountRUB() != domain.FixedPaymentAmountRUB {
		t.Errorf("expected 10.00 RUB, got %s", payment.AmountRUB())
	}
	if payment.Status() != domain.StatusPending {
		t.Errorf("expected status pending, got %s", payment.Status())
	}

	// 2. Handle Succeeded Webhook
	payload := map[string]interface{}{
		"provider_payment_id": payment.ProviderPaymentID(),
		"order_id":           orderID.String(),
		"status":              string(contracts.PaymentStatusSucceeded),
	}
	rawBody, _ := json.Marshal(payload)

	if err := uc.HandleWebhook(ctx, rawBody, nil); err != nil {
		t.Fatalf("unexpected error handling webhook: %v", err)
	}

	updated, err := uc.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		t.Fatalf("unexpected error getting payment: %v", err)
	}
	if updated.Status() != domain.StatusSucceeded {
		t.Errorf("expected status succeeded, got %s", updated.Status())
	}
}
