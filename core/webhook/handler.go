package webhook

import (
	"encoding/json"
	"io"
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

func (h *Handler) Ingest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(parts) != 2 || parts[0] != "ingest" || parts[1] == "" {
		http.Error(w, "invalid subscription ID", http.StatusBadRequest)
		return
	}

	subscriptionID := parts[1]

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	if len(payload) == 0 {
		http.Error(w, "request body is required", http.StatusBadRequest)
		return
	}

	var jsonData interface{}

	if err := json.Unmarshal(payload, &jsonData); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	webhook, err := h.service.Create(
		r.Context(),
		subscriptionID,
		payload,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     webhook.ID,
		"status": webhook.Status,
	})
}