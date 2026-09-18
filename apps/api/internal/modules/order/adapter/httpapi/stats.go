package httpapi

import (
	"net/http"

	"github.com/ostkost/dopamine-market/api/internal/platform/httpserver"
)

// handleGetUserStats возвращает статистику заказов и streak текущего пользователя (EPIC-14, FR-GAMIFY-01).
func (h *Handler) handleGetUserStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tz := r.URL.Query().Get("tz")
	stats, err := h.orderUC.GetUserStats(r.Context(), userID, tz)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to calculate user stats")
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}
