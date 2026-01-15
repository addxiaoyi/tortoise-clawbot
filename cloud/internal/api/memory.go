package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
)

type MemoryHandler struct {
	logger zerolog.Logger
}

type MemoryEntry struct {
	Key    string      `json:"key"`
	Value  interface{} `json:"value"`
	Type   string      `json:"type"`
	TTL    int         `json:"ttl,omitempty"`
}

func (h *MemoryHandler) List(w http.ResponseWriter, r *http.Request) {
	// TODO: List memory entries
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": []MemoryEntry{},
	})
}

func (h *MemoryHandler) Store(w http.ResponseWriter, r *http.Request) {
	var req MemoryEntry

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	h.logger.Info().Str("key", req.Key).Msg("memory stored")

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "stored",
	})
}

func (h *MemoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	// key := chi.URLParam(r, "key")

	entry := MemoryEntry{
		Key:   "test-key",
		Value: "test-value",
		Type:  "episodic",
	}

	json.NewEncoder(w).Encode(entry)
}

func (h *MemoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// key := chi.URLParam(r, "key")

	w.WriteHeader(http.StatusNoContent)
}

func (h *MemoryHandler) Search(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
		Type  string `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// TODO: Search memory
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": []MemoryEntry{},
	})
}
