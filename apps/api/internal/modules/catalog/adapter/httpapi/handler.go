package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/catalog/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/catalog/usecase"
)

type Handler struct {
	catalogUC *usecase.CatalogUseCase
}

func NewHandler(catalogUC *usecase.CatalogUseCase) *Handler {
	return &Handler{catalogUC: catalogUC}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/categories", h.handleListCategories)
	r.Get("/products", h.handleListProducts)
	r.Get("/products/{id}", h.handleGetProduct)

	return r
}

type CategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type ProductResponse struct {
	ID           string `json:"id"`
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	PriceRUB     string `json:"price_rub"`
	ImageSeed    string `json:"image_seed"`
	CreatedAt    string `json:"created_at"`
}

type ProductsListResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

func (h *Handler) handleListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.catalogUC.ListCategories(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}

	resp := make([]CategoryResponse, 0, len(categories))
	for _, c := range categories {
		resp = append(resp, CategoryResponse{
			ID:          c.ID().String(),
			Name:        c.Name(),
			Slug:        c.Slug(),
			Description: c.Description(),
		})
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"categories": resp,
	})
}

func (h *Handler) handleListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var categoryID *uuid.UUID
	if catStr := r.URL.Query().Get("category_id"); catStr != "" {
		if id, err := uuid.Parse(catStr); err == nil {
			categoryID = &id
		}
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

	products, total, err := h.catalogUC.ListProducts(r.Context(), categoryID, q, limit, offset)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to list products")
		return
	}

	respProducts := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		respProducts = append(respProducts, ProductResponse{
			ID:           p.ID().String(),
			CategoryID:   p.CategoryID().String(),
			CategoryName: p.CategoryName(),
			Name:         p.Name(),
			Description:  p.Description(),
			PriceRUB:     p.PriceRUB().StringFixed(2),
			ImageSeed:    p.ImageSeed(),
			CreatedAt:    p.CreatedAt().Format(time.RFC3339),
		})
	}

	h.writeJSON(w, http.StatusOK, ProductsListResponse{
		Products: respProducts,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	})
}

func (h *Handler) handleGetProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	prod, err := h.catalogUC.GetProductByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			h.writeError(w, http.StatusNotFound, "product not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "failed to get product")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"product": ProductResponse{
			ID:           prod.ID().String(),
			CategoryID:   prod.CategoryID().String(),
			CategoryName: prod.CategoryName(),
			Name:         prod.Name(),
			Description:  prod.Description(),
			PriceRUB:     prod.PriceRUB().StringFixed(2),
			ImageSeed:    prod.ImageSeed(),
			CreatedAt:    prod.CreatedAt().Format(time.RFC3339),
		},
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
