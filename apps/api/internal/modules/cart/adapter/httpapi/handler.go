package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/cart/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/cart/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/httpserver"
)

type Handler struct {
	cartUC *usecase.CartUseCase
}

func NewHandler(cartUC *usecase.CartUseCase) *Handler {
	return &Handler{cartUC: cartUC}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.handleGetCart)
	r.Post("/items", h.handleAddItem)
	r.Patch("/items/{product_id}", h.handleUpdateQuantity)
	r.Delete("/items/{product_id}", h.handleRemoveItem)
	r.Put("/pickup-point", h.handleSetPickupPoint)
	r.Delete("/", h.handleClearCart)

	return r
}

type AddItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type UpdateQuantityRequest struct {
	Quantity int `json:"quantity"`
}

type SetPickupPointRequest struct {
	PickupPointID string `json:"pickup_point_id"`
}

func (h *Handler) handleGetCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cart, err := h.cartUC.GetCart(r.Context(), userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to get cart")
		return
	}

	h.writeJSON(w, http.StatusOK, cart)
}

func (h *Handler) handleAddItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid product_id")
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	cart, err := h.cartUC.AddItem(r.Context(), userID, productID, req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			h.writeError(w, http.StatusNotFound, "product not found in catalog")
		case errors.Is(err, domain.ErrInvalidQuantity):
			h.writeError(w, http.StatusBadRequest, "quantity must be greater than 0")
		default:
			h.writeError(w, http.StatusInternalServerError, "failed to add item to cart")
		}
		return
	}

	h.writeJSON(w, http.StatusOK, cart)
}

func (h *Handler) handleUpdateQuantity(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	prodIDStr := chi.URLParam(r, "product_id")
	productID, err := uuid.Parse(prodIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid product_id")
		return
	}

	var req UpdateQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cart, err := h.cartUC.UpdateQuantity(r.Context(), userID, productID, req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			h.writeError(w, http.StatusNotFound, "product not found in catalog")
		case errors.Is(err, domain.ErrInvalidQuantity):
			h.writeError(w, http.StatusBadRequest, "quantity must be non-negative")
		default:
			h.writeError(w, http.StatusInternalServerError, "failed to update quantity")
		}
		return
	}

	h.writeJSON(w, http.StatusOK, cart)
}

func (h *Handler) handleRemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	prodIDStr := chi.URLParam(r, "product_id")
	productID, err := uuid.Parse(prodIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid product_id")
		return
	}

	cart, err := h.cartUC.RemoveItem(r.Context(), userID, productID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to remove item")
		return
	}

	h.writeJSON(w, http.StatusOK, cart)
}

func (h *Handler) handleSetPickupPoint(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req SetPickupPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pointID, err := uuid.Parse(req.PickupPointID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid pickup_point_id")
		return
	}

	cart, err := h.cartUC.SetPickupPoint(r.Context(), userID, pointID)
	if err != nil {
		if errors.Is(err, domain.ErrPickupPointNotFound) {
			h.writeError(w, http.StatusNotFound, "pickup point not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "failed to set pickup point")
		return
	}

	h.writeJSON(w, http.StatusOK, cart)
}

func (h *Handler) handleClearCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.cartUC.ClearCart(r.Context(), userID); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to clear cart")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
