package api

import (
	"encoding/json"
	"net/http"

	"github.com/tortoise/cloud/internal/vector"
)

type VectorHandler struct {
	client *vector.Client
}

func (h *VectorHandler) Embed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	embedding, err := h.client.Embed(req.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"embedding": embedding,
	})
}

func (h *VectorHandler) Search(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query     string    `json:"query"`
		Limit     int       `json:"limit"`
		Threshold float64   `json:"threshold"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	results, err := h.client.Search(req.Query, req.Limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
	})
}
