package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/notification/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/notification/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/httpserver"
)

type Handler struct {
	uc *usecase.NotificationUseCase
}

func NewHandler(uc *usecase.NotificationUseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}/events", h.HandleStreamOrderEvents)
	return r
}

// HandleStreamOrderEvents обслуживает SSE-поток статуса заказа (FR-NOTIF-01, ADR-009).
func (h *Handler) HandleStreamOrderEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid_order_id"}`, http.StatusBadRequest)
		return
	}

	// 1. Получаем начальный снимок состояния заказа с валидацией прав
	snapshot, err := h.uc.GetOrderSnapshot(r.Context(), orderID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			http.Error(w, `{"error":"forbidden","message":"access to order is forbidden"}`, http.StatusForbidden)
			return
		}
		if errors.Is(err, domain.ErrOrderNotFound) {
			http.Error(w, `{"error":"not_found","message":"order not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"internal_error"}`, http.StatusInternalServerError)
		return
	}

	// 2. Устанавливаем SSE заголовки
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	// 3. Отправляем начальный снимок (snapshot)
	snapshotBytes, err := json.Marshal(snapshot)
	if err == nil {
		_, _ = fmt.Fprintf(w, "event: snapshot\ndata: %s\n\n", string(snapshotBytes))
		flusher.Flush()
	}

	// 4. Подписываемся на поток обновлений
	eventsCh, unsubscribe := h.uc.Subscribe(orderID)
	defer unsubscribe()

	keepAliveTicker := time.NewTicker(15 * time.Second)
	defer keepAliveTicker.Stop()

	for {
		select {
		case <-r.Context().Done():
			// Клиент закрыл соединение
			return

		case <-keepAliveTicker.C:
			// Keep-alive комментарий для предотвращения обрыва прокси (FR-NOTIF-01)
			_, err := fmt.Fprintf(w, ": keep-alive\n\n")
			if err != nil {
				return
			}
			flusher.Flush()

		case event, ok := <-eventsCh:
			if !ok {
				return
			}

			payloadBytes, err := json.Marshal(event.Payload)
			if err != nil {
				continue
			}

			_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(payloadBytes))
			if err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
