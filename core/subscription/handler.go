package subscription

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

type createSubscriptionRequest struct {
	TargetURL string `json:"target_url"`
	Secret    string `json:"secret"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createSubscriptionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	subscription, err := h.service.Create(r.Context(), req.TargetURL, req.Secret)
	if err != nil {
		log.Printf("Create subscription error: %v", err)
		http.Error(w, "Failed to create subscription", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, subscription)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/subscriptions/")

	if id == "" {
		http.Error(w, "subscription ID is required", http.StatusBadRequest)
		return
	}

	subscription, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, subscription)
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	subscriptions, err := h.service.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, subscriptions)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/subscriptions/")

	if id == "" {
		http.Error(w, "subscription ID is required", http.StatusBadRequest)
		return
	}

	var req createSubscriptionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	subscription, err := h.service.Update(
		r.Context(),
		id,
		req.TargetURL,
		req.Secret,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, subscription)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/subscriptions/")

	if id == "" {
		http.Error(w, "subscription ID is required", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
