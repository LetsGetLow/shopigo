package cataloghttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	appcatalog "shopigo/internal/app/catalog"
	domainshared "shopigo/internal/domain/shared"

	"github.com/google/uuid"
)

const actorHeader = "X-Actor-ID"

type service interface {
	CreateCategory(ctx context.Context, req appcatalog.CreateCategoryRequest, user domainshared.ActorID) (appcatalog.CreateCategoryResponse, error)
	UpdateCategory(ctx context.Context, req appcatalog.UpdateCategoryRequest, user domainshared.ActorID) (appcatalog.UpdateCategoryResponse, error)
	DeleteCategory(ctx context.Context, req appcatalog.DeleteCategoryRequest, user domainshared.ActorID) (appcatalog.DeleteCategoryResponse, error)
	GetCategory(ctx context.Context, id string) (appcatalog.CategoryResponse, error)
	ListCategories(ctx context.Context) ([]appcatalog.CategoryResponse, error)
}

type CategoryHandler struct {
	service service
}

func NewCategoryHandler(service service) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /categories", h.handleCreateCategory)
	mux.HandleFunc("GET /categories", h.handleListCategories)
	mux.HandleFunc("GET /categories/{id}", h.handleGetCategory)
	mux.HandleFunc("PUT /categories/{id}", h.handleUpdateCategory)
	mux.HandleFunc("DELETE /categories/{id}", h.handleDeleteCategory)
}

func (h *CategoryHandler) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	userID, ok := actorIDFromRequest(r)
	if !ok {
		http.Error(w, "missing or invalid actor id", http.StatusBadRequest)
		return
	}

	var req appcatalog.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.service.CreateCategory(r.Context(), req, userID)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *CategoryHandler) handleGetCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	resp, err := h.service.GetCategory(r.Context(), r.PathValue("id"))
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *CategoryHandler) handleListCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	resp, err := h.service.ListCategories(r.Context())
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *CategoryHandler) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	userID, ok := actorIDFromRequest(r)
	if !ok {
		http.Error(w, "missing or invalid actor id", http.StatusBadRequest)
		return
	}

	var req appcatalog.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.ID = r.PathValue("id")

	resp, err := h.service.UpdateCategory(r.Context(), req, userID)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *CategoryHandler) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	userID, ok := actorIDFromRequest(r)
	if !ok {
		http.Error(w, "missing or invalid actor id", http.StatusBadRequest)
		return
	}

	resp, err := h.service.DeleteCategory(r.Context(), appcatalog.DeleteCategoryRequest{ID: r.PathValue("id")}, userID)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func actorIDFromRequest(r *http.Request) (domainshared.ActorID, bool) {
	raw := r.Header.Get(actorHeader)
	if raw == "" {
		return domainshared.ActorID(uuid.Nil), false
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return domainshared.ActorID(uuid.Nil), false
	}

	return domainshared.ActorID(id), true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeHandlerError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, appcatalog.ErrCategoryNotFound):
		http.Error(w, "category not found", http.StatusNotFound)
	case errors.Is(err, appcatalog.ErrCategorySaveFailed), errors.Is(err, appcatalog.ErrCategoryDeleteFailed):
		http.Error(w, "internal server error", http.StatusInternalServerError)
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
