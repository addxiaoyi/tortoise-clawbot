package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
)

type MeshHandler struct {
	logger zerolog.Logger
}

type MeshNode struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	Capabilities []string `json:"capabilities"`
	Status      string   `json:"status"`
}

func (h *MeshHandler) ListNodes(w http.ResponseWriter, r *http.Request) {
	// TODO: List mesh nodes
	nodes := []MeshNode{
		{
			ID:           "node-1",
			Name:         "local-node",
			Address:      "127.0.0.1:8080",
			Capabilities: []string{"messaging", "delegation"},
			Status:       "online",
		},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"nodes": nodes,
	})
}

func (h *MeshHandler) Connect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Address string `json:"address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	h.logger.Info().Str("address", req.Address).Msg("connecting to mesh node")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "connected",
	})
}

func (h *MeshHandler) Delegate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NodeID   string `json:"node_id"`
		Task     string `json:"task"`
		Priority string `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	h.logger.Info().
		Str("node_id", req.NodeID).
		Str("task", req.Task).
		Msg("delegating task")

	json.NewEncoder(w).Encode(map[string]string{
		"message": "task delegated",
	})
}
