package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/usecase"
	"github.com/ostkost/dopamine-market/api/internal/platform/httpserver"
)

type Handler struct {
	orderUC *usecase.OrderUseCase
}

func NewHandler(orderUC *usecase.OrderUseCase) *Handler {
	return &Handler{orderUC: orderUC}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.handleCreateOrder)
	r.Get("/", h.handleListOrders)
	r.Get("/{id}", h.handleGetOrder)
	r.Post("/{id}/cancel", h.handleCancelOrder)

	return r
}

type OrderItemResponse struct {
	ProductID    string `json:"product_id"`
	Name         string `json:"name"`
	CategoryName string `json:"category_name"`
	PriceRUB     string `json:"price_rub"`
	ImageSeed    string `json:"image_seed"`
	Quantity     int    `json:"quantity"`
	SubtotalRUB  string `json:"subtotal_rub"`
}

type PickupPointResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	DistanceMeters float64 `json:"distance_meters"`
}

type StatusHistoryResponse struct {
	FromStatus string `json:"from_status"`
	ToStatus   string `json:"to_status"`
	Reason     string `json:"reason"`
	CreatedAt  string `json:"created_at"`
}

type OrderResponse struct {
	ID             string                  `json:"id"`
	UserID         string                  `json:"user_id"`
	Status         string                  `json:"status"`
	TotalAmountRUB string                  `json:"total_amount_rub"` // 10.00 RUB
	PickupPoint    PickupPointResponse     `json:"pickup_point"`
	Items          []OrderItemResponse     `json:"items,omitempty"`
	History        []StatusHistoryResponse `json:"history,omitempty"`
	CreatedAt      string                  `json:"created_at"`
	UpdatedAt      string                  `json:"updated_at"`
}

type OrderListResponse struct {
	Orders []OrderResponse `json:"orders"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

func (h *Handler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	order, err := h.orderUC.CreateOrder(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCartEmpty):
			h.writeError(w, http.StatusBadRequest, "cart is empty")
		case errors.Is(err, domain.ErrPickupPointRequired):
			h.writeError(w, http.StatusBadRequest, "pickup point is required before checkout")
		case errors.Is(err, domain.ErrDuplicatePendingOrder):
			h.writeError(w, http.StatusConflict, "you already have an active order waiting for payment (INV-02)")
		default:
			h.writeError(w, http.StatusInternalServerError, "failed to create order")
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, h.mapOrderResponse(order, nil))
}

func (h *Handler) handleListOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := 20
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	orders, total, err := h.orderUC.ListByUserID(r.Context(), userID, limit, offset)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to list orders")
		return
	}

	resp := make([]OrderResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, h.mapOrderResponse(o, nil))
	}

	h.writeJSON(w, http.StatusOK, OrderListResponse{
		Orders: resp,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *Handler) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.orderUC.GetByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			h.writeError(w, http.StatusNotFound, "order not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "failed to get order")
		return
	}

	if order.UserID() != userID {
		h.writeError(w, http.StatusNotFound, "order not found")
		return
	}

	history, _ := h.orderUC.GetStatusHistory(r.Context(), orderID)

	var historyResp []StatusHistoryResponse
	for _, rec := range history {
		historyResp = append(historyResp, StatusHistoryResponse{
			FromStatus: rec.FromStatus.String(),
			ToStatus:   rec.ToStatus.String(),
			Reason:     rec.Reason,
			CreatedAt:  rec.CreatedAt.Format(time.RFC3339),
		})
	}

	h.writeJSON(w, http.StatusOK, h.mapOrderResponse(order, historyResp))
}

func (h *Handler) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpserver.UserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	order, err := h.orderUC.CancelOrder(r.Context(), userID, orderID, req.Reason)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidStatusTransition) {
			h.writeError(w, http.StatusBadRequest, "order cannot be cancelled in current status")
			return
		}
		if errors.Is(err, domain.ErrOrderNotFound) {
			h.writeError(w, http.StatusNotFound, "order not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "failed to cancel order")
		return
	}

	h.writeJSON(w, http.StatusOK, h.mapOrderResponse(order, nil))
}

func (h *Handler) mapOrderResponse(o *domain.Order, history []StatusHistoryResponse) OrderResponse {
	items := make([]OrderItemResponse, 0, len(o.Items()))
	for _, it := range o.Items() {
		items = append(items, OrderItemResponse{
			ProductID:    it.ProductID().String(),
			Name:         it.Name(),
			CategoryName: it.CategoryName(),
			PriceRUB:     it.PriceRUB().StringFixed(2),
			ImageSeed:    it.ImageSeed(),
			Quantity:     it.Quantity(),
			SubtotalRUB:  it.SubtotalRUB().StringFixed(2),
		})
	}

	return OrderResponse{
		ID:             o.ID().String(),
		UserID:         o.UserID().String(),
		Status:         o.Status().String(),
		TotalAmountRUB: o.TotalAmountRUB().StringFixed(2),
		PickupPoint: PickupPointResponse{
			ID:             o.PickupPoint().ID.String(),
			Name:           o.PickupPoint().Name,
			Latitude:       o.PickupPoint().Latitude,
			Longitude:      o.PickupPoint().Longitude,
			DistanceMeters: o.PickupPoint().DistanceMeters,
		},
		Items:     items,
		History:   history,
		CreatedAt: o.CreatedAt().Format(time.RFC3339),
		UpdatedAt: o.UpdatedAt().Format(time.RFC3339),
	}
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
