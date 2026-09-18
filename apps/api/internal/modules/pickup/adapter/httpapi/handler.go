package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ostkost/dopamine-market/api/internal/modules/pickup/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/pickup/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/httpserver"
)

type Handler struct {
	pickupUC *usecase.PickupUseCase
}

func NewHandler(pickupUC *usecase.PickupUseCase) *Handler {
	return &Handler{pickupUC: pickupUC}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/generate", h.handleGenerate)
	r.Get("/points", h.handleListPoints)

	return r
}

type GenerateRequest struct {
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	City      string   `json:"city,omitempty"`
}

type PickupPointResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	DistanceMeters float64 `json:"distance_meters"`
	CreatedAt      string  `json:"created_at"`
}

func (h *Handler) handleGenerate(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req GenerateRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	origin := usecase.DefaultFallbackLocation
	if req.Latitude != nil && req.Longitude != nil {
		if latLng, err := domain.NewLatLng(*req.Latitude, *req.Longitude); err == nil {
			origin = latLng
		}
	}

	points, err := h.pickupUC.GeneratePickupPoints(r.Context(), userID, origin)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to generate pickup points")
		return
	}

	resp := make([]PickupPointResponse, 0, len(points))
	for _, p := range points {
		resp = append(resp, PickupPointResponse{
			ID:             p.ID().String(),
			Name:           p.Name(),
			Latitude:       p.Location().Latitude,
			Longitude:      p.Location().Longitude,
			DistanceMeters: p.DistanceMeters(),
			CreatedAt:      p.CreatedAt().Format(time.RFC3339),
		})
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"points": resp,
	})
}

func (h *Handler) handleListPoints(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	points, err := h.pickupUC.ListPickupPoints(r.Context(), userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to list pickup points")
		return
	}

	resp := make([]PickupPointResponse, 0, len(points))
	for _, p := range points {
		resp = append(resp, PickupPointResponse{
			ID:             p.ID().String(),
			Name:           p.Name(),
			Latitude:       p.Location().Latitude,
			Longitude:      p.Location().Longitude,
			DistanceMeters: p.DistanceMeters(),
			CreatedAt:      p.CreatedAt().Format(time.RFC3339),
		})
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"points": resp,
	})
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{
		"error":   http.StatusText(status),
		"message": msg,
	})
}
