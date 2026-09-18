package yookassa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
)

type Config struct {
	ShopID    string
	SecretKey string
	BaseURL   string // defaults to https://api.yookassa.ru/v3
}

type Provider struct {
	cfg        Config
	httpClient *http.Client
}

func NewProvider(cfg Config, client *http.Client) *Provider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.yookassa.ru/v3"
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Provider{
		cfg:        cfg,
		httpClient: client,
	}
}

func (p *Provider) Name() string {
	return "yookassa"
}

type yooAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type yooConfirmation struct {
	Type            string `json:"type"`
	ReturnURL       string `json:"return_url,omitempty"`
	ConfirmationURL string `json:"confirmation_url,omitempty"`
}

type yooCreatePaymentRequest struct {
	Amount       yooAmount         `json:"amount"`
	Capture      bool              `json:"capture"`
	Confirmation yooConfirmation   `json:"confirmation"`
	Description  string            `json:"description,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type yooPaymentResponse struct {
	ID           string           `json:"id"`
	Status       string           `json:"status"`
	Paid         bool             `json:"paid"`
	Amount       yooAmount        `json:"amount"`
	Confirmation *yooConfirmation `json:"confirmation,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

func (p *Provider) InitiatePayment(ctx context.Context, req contracts.InitiatePaymentRequest) (contracts.InitiatePaymentResult, error) {
	createReq := yooCreatePaymentRequest{
		Amount: yooAmount{
			Value:    "10.00", // INV-01
			Currency: "RUB",
		},
		Capture: true,
		Confirmation: yooConfirmation{
			Type:      "redirect",
			ReturnURL: req.ReturnURL,
		},
		Description: req.Description,
		Metadata: map[string]string{
			"order_id": req.OrderID.String(),
			"user_id":  req.UserID.String(),
		},
	}

	bodyBytes, err := json.Marshal(createReq)
	if err != nil {
		return contracts.InitiatePaymentResult{}, fmt.Errorf("marshaling yookassa request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.BaseURL+"/payments", bytes.NewReader(bodyBytes))
	if err != nil {
		return contracts.InitiatePaymentResult{}, fmt.Errorf("creating http request: %w", err)
	}

	httpReq.SetBasicAuth(p.cfg.ShopID, p.cfg.SecretKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", req.OrderID.String())

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return contracts.InitiatePaymentResult{}, fmt.Errorf("calling yookassa api: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return contracts.InitiatePaymentResult{}, fmt.Errorf("reading yookassa response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return contracts.InitiatePaymentResult{}, fmt.Errorf("yookassa error status %d: %s", resp.StatusCode, string(respBytes))
	}

	var yooResp yooPaymentResponse
	if err := json.Unmarshal(respBytes, &yooResp); err != nil {
		return contracts.InitiatePaymentResult{}, fmt.Errorf("unmarshaling yookassa response: %w", err)
	}

	var confirmationURL string
	if yooResp.Confirmation != nil {
		confirmationURL = yooResp.Confirmation.ConfirmationURL
	}

	return contracts.InitiatePaymentResult{
		ProviderPaymentID: yooResp.ID,
		ConfirmationURL:   confirmationURL,
		Status:            contracts.PaymentStatusPending,
	}, nil
}

type yooWebhookNotification struct {
	Type   string             `json:"type"`
	Event  string             `json:"event"`
	Object yooPaymentResponse `json:"object"`
}

func (p *Provider) VerifyWebhook(ctx context.Context, rawBody []byte, headers map[string]string) (contracts.WebhookResult, error) {
	var notif yooWebhookNotification
	if err := json.Unmarshal(rawBody, &notif); err != nil {
		return contracts.WebhookResult{}, fmt.Errorf("unmarshaling yookassa webhook: %w", err)
	}

	orderIDStr := notif.Object.Metadata["order_id"]
	orderUUID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return contracts.WebhookResult{}, fmt.Errorf("missing or invalid order_id in yookassa metadata: %w", err)
	}

	var status contracts.PaymentStatus
	var failureReason string

	switch notif.Event {
	case "payment.succeeded":
		status = contracts.PaymentStatusSucceeded
	case "payment.canceled":
		status = contracts.PaymentStatusFailed
		failureReason = "payment canceled by yookassa"
	default:
		status = contracts.PaymentStatusPending
	}

	return contracts.WebhookResult{
		ProviderPaymentID: notif.Object.ID,
		OrderID:           orderUUID,
		Status:            status,
		FailureReason:     failureReason,
	}, nil
}
