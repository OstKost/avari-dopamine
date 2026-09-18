package mock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/platform/random"
)

type Provider struct {
	successRate float64
	rnd         random.Source
	webhookURL  string
}

func NewProvider(successRate float64, rnd random.Source, webhookURL string) *Provider {
	if successRate <= 0 || successRate > 1.0 {
		successRate = 0.95
	}
	if rnd == nil {
		rnd = random.New(time.Now().UnixNano())
	}
	return &Provider{
		successRate: successRate,
		rnd:         rnd,
		webhookURL:  webhookURL,
	}
}

func (p *Provider) Name() string {
	return "mock"
}

func (p *Provider) InitiatePayment(ctx context.Context, req contracts.InitiatePaymentRequest) (contracts.InitiatePaymentResult, error) {
	providerPaymentID := fmt.Sprintf("mock_pay_%s", uuid.New().String())

	// Асинхронно симулируем отправку вебхука через 1 секунду, если указан webhookURL
	if p.webhookURL != "" {
		isSuccess := p.rnd.Bool(p.successRate)
		status := contracts.PaymentStatusSucceeded
		var failureReason string
		if !isSuccess {
			status = contracts.PaymentStatusFailed
			failureReason = "insufficient funds (mock simulation)"
		}

		go func() {
			time.Sleep(1 * time.Second)
			payload := map[string]interface{}{
				"provider_payment_id": providerPaymentID,
				"order_id":           req.OrderID.String(),
				"status":              string(status),
				"failure_reason":      failureReason,
			}
			body, _ := json.Marshal(payload)
			req, err := http.NewRequest(http.MethodPost, p.webhookURL, bytes.NewReader(body))
			if err == nil {
				req.Header.Set("Content-Type", "application/json")
				client := &http.Client{Timeout: 5 * time.Second}
				_, _ = client.Do(req)
			}
		}()
	}

	return contracts.InitiatePaymentResult{
		ProviderPaymentID: providerPaymentID,
		ConfirmationURL:   fmt.Sprintf("http://localhost:8080/payments/mock/confirm?payment_id=%s", providerPaymentID),
		Status:            contracts.PaymentStatusPending,
	}, nil
}

type MockWebhookPayload struct {
	ProviderPaymentID string `json:"provider_payment_id"`
	OrderID           string `json:"order_id"`
	Status            string `json:"status"`
	FailureReason     string `json:"failure_reason,omitempty"`
}

func (p *Provider) VerifyWebhook(ctx context.Context, rawBody []byte, headers map[string]string) (contracts.WebhookResult, error) {
	var payload MockWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return contracts.WebhookResult{}, fmt.Errorf("unmarshaling mock webhook: %w", err)
	}

	orderUUID, err := uuid.Parse(payload.OrderID)
	if err != nil {
		return contracts.WebhookResult{}, fmt.Errorf("parsing order_id: %w", err)
	}

	return contracts.WebhookResult{
		ProviderPaymentID: payload.ProviderPaymentID,
		OrderID:           orderUUID,
		Status:            contracts.PaymentStatus(payload.Status),
		FailureReason:     payload.FailureReason,
	}, nil
}
