package payment

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/adapter/httpapi"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/adapter/kafka"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/adapter/mock"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/adapter/postgres"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/adapter/yookassa"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/config"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
	"github.com/ostkost/dopamine-market/api/internal/platform/random"
)

type Module struct {
	paymentUC       *usecase.PaymentUseCase
	handler         *httpapi.Handler
	consumerHandler *kafka.ConsumerHandler
}

func NewModule(
	dbPool *db.Pool,
	paymentCfg config.PaymentConfig,
	rnd random.Source,
) (*Module, error) {
	var provider contracts.PaymentProvider

	switch paymentCfg.Provider {
	case "yookassa":
		if paymentCfg.YooKassaShopID == "" || paymentCfg.YooKassaSecretKey == "" {
			return nil, fmt.Errorf("yookassa provider selected but shop_id or secret_key is missing")
		}
		provider = yookassa.NewProvider(yookassa.Config{
			ShopID:    paymentCfg.YooKassaShopID,
			SecretKey: paymentCfg.YooKassaSecretKey,
		}, &http.Client{})
	case "mock":
		fallthrough
	default:
		webhookURL := "http://localhost:8080/payments/webhook/mock"
		provider = mock.NewProvider(0.95, rnd, webhookURL)
	}

	repo := postgres.NewPaymentRepository(dbPool)
	paymentUC := usecase.NewPaymentUseCase(repo, provider, dbPool)
	handler := httpapi.NewHandler(paymentUC, provider.Name())
	consumerHandler := kafka.NewConsumerHandler(paymentUC)

	return &Module{
		paymentUC:       paymentUC,
		handler:         handler,
		consumerHandler: consumerHandler,
	}, nil
}

func (m *Module) Routes() chi.Router {
	return m.handler.Routes()
}

func (m *Module) ConsumerHandler() *kafka.ConsumerHandler {
	return m.consumerHandler
}

func (m *Module) UseCase() *usecase.PaymentUseCase {
	return m.paymentUC
}

// Реализация contracts.PaymentLookup

func (m *Module) GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*contracts.PaymentSnapshot, error) {
	p, err := m.paymentUC.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting payment for order %s: %w", orderID, err)
	}

	return &contracts.PaymentSnapshot{
		ID:                p.ID(),
		OrderID:           p.OrderID(),
		Provider:          p.Provider(),
		ProviderPaymentID: p.ProviderPaymentID(),
		AmountRUB:         p.AmountRUB(),
		Status:            contracts.PaymentStatus(p.Status()),
		CreatedAt:         p.CreatedAt(),
		UpdatedAt:         p.UpdatedAt(),
	}, nil
}
