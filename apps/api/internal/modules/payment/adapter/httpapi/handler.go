package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/usecase"
)

type Handler struct {
	paymentUC    *usecase.PaymentUseCase
	providerName string
}

func NewHandler(paymentUC *usecase.PaymentUseCase, providerName string) *Handler {
	return &Handler{
		paymentUC:    paymentUC,
		providerName: providerName,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/webhook/mock", h.handleMockWebhook)
	r.Post("/webhook/yookassa", h.handleYooKassaWebhook)
	r.Get("/{order_id}", h.handleGetPayment)

	return r
}

func (h *Handler) handleMockWebhook(w http.ResponseWriter, r *http.Request) {
	if h.providerName != "mock" {
		http.Error(w, "mock webhook disabled in current configuration", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	if err := h.paymentUC.HandleWebhook(r.Context(), body, headers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) handleYooKassaWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	if err := h.paymentUC.HandleWebhook(r.Context(), body, headers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) handleGetPayment(w http.ResponseWriter, r *http.Request) {
	orderIDStr := chi.URLParam(r, "order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		http.Error(w, "invalid order_id", http.StatusBadRequest)
		return
	}

	payment, err := h.paymentUC.GetPaymentByOrderID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			http.Error(w, "payment not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":                  payment.ID(),
		"order_id":           payment.OrderID(),
		"provider":           payment.Provider(),
		"provider_payment_id": payment.ProviderPaymentID(),
		"amount_rub":         payment.AmountRUB(),
		"status":              payment.Status(),
		"confirmation_url":   payment.ConfirmationURL(),
		"created_at":          payment.CreatedAt(),
		"updated_at":          payment.UpdatedAt(),
	})
}
