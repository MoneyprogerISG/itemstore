package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"myproject/internal/model"
	"myproject/internal/service"
	"net/http"
)

type HandlerItem struct {
	service *service.ItemService
}

func NewHandlerItem(svc *service.ItemService) *HandlerItem {
	return &HandlerItem{service: svc}
}

func (h *HandlerItem) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /items", h.List)
	mux.HandleFunc("GET /items/{id}", h.Get)
	mux.HandleFunc("POST /items", h.Create)
	mux.HandleFunc("PUT /items/{id}", h.Update)
	mux.HandleFunc("DELETE /items/{id}", h.Delete)
}

func (h *HandlerItem) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.GetAll()
	if err != nil {
		slog.Error("list items", "error", err)
		handlerError(w, err)
		return
	}
	respondeJSON(w, http.StatusOK, items)
}

func (h *HandlerItem) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	item, err := h.service.GetByID(id)
	if err != nil {
		handlerError(w, err)
		return
	}
	respondeJSON(w, http.StatusOK, item)
}

func (h *HandlerItem) Create(w http.ResponseWriter, r *http.Request) {
	var item model.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "невалидный JSON", http.StatusBadRequest)
		return
	}

	created, err := h.service.Create(item)
	if err != nil {
		slog.Error("create item", "error", err)
		handlerError(w, err)
		return
	}
	respondeJSON(w, http.StatusCreated, created)
}

func (h *HandlerItem) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var item model.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "невалидный JSON", http.StatusBadRequest)
		return
	}

	update, err := h.service.Update(id, item)
	if err != nil {
		slog.Error("update item", "id", id, "error", err)
		handlerError(w, err)
		return
	}

	respondeJSON(w, http.StatusOK, update)
}

func (h *HandlerItem) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.service.Delete(id)
	if err != nil {
		handlerError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handlerError(w http.ResponseWriter, err error) {
	var ve *model.ValidationError
	switch {
	case errors.As(err, &ve):
		http.Error(w, ve.Error(), http.StatusBadRequest)
	case errors.Is(err, model.ErrNotFound):
		http.Error(w, "не найдено", http.StatusNotFound)
	case errors.Is(err, model.ErrAlreadyExists):
		http.Error(w, "уже существует", http.StatusConflict)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func respondeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
