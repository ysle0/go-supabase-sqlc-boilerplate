package example_feature

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for items
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new item handler
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers all item routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/items", func(r chi.Router) {
		r.Get("/", h.ListItems)
		r.Post("/", h.CreateItem)
		r.Get("/{id}", h.GetItem)
		r.Put("/{id}", h.UpdateItem)
		r.Delete("/{id}", h.DeleteItem)
	})
}

// CreateItem handles POST /items
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.CreateItem(r.Context(), req)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, item)
}

// GetItem handles GET /items/{id}
func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid item ID")
		return
	}

	item, err := h.service.GetItem(r.Context(), id)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "item not found")
		return
	}

	h.respondJSON(w, http.StatusOK, item)
}

// ListItems handles GET /items
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr := r.URL.Query().Get("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			pageSize = s
		}
	}

	response, err := h.service.ListItems(r.Context(), page, pageSize)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to list items")
		return
	}

	h.respondJSON(w, http.StatusOK, response)
}

// UpdateItem handles PUT /items/{id}
func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid item ID")
		return
	}

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.UpdateItem(r.Context(), id, req)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, item)
}

// DeleteItem handles DELETE /items/{id}
func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid item ID")
		return
	}

	if err := h.service.DeleteItem(r.Context(), id); err != nil {
		h.respondError(w, http.StatusNotFound, "item not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// respondJSON responds with JSON
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// respondError responds with an error
func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}
