package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type AgentHandler struct {
	logger zerolog.Logger
}

type Agent struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Model       string                 `json:"model"`
	State       string                 `json:"state"`
	Skills      []string               `json:"skills"`
	Permissions []string               `json:"permissions"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

func (h *AgentHandler) List(w http.ResponseWriter, r *http.Request) {
	// TODO: List agents from storage
	agents := []Agent{
		{
			ID:    uuid.New().String(),
			Name:  "default-agent",
			Model: "gpt-4",
			State: "running",
		},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"agents": agents,
	})
}

func (h *AgentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string                 `json:"name"`
		Model   string                 `json:"model"`
		Skills  []string               `json:"skills"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	agent := Agent{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Model:    req.Model,
		State:    "created",
		Skills:   req.Skills,
		Metadata: req.Metadata,
	}

	h.logger.Info().Str("agent_id", agent.ID).Str("name", agent.Name).Msg("agent created")

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(agent)
}

func (h *AgentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// TODO: Fetch from storage
	agent := Agent{
		ID:    id,
		Name:  "test-agent",
		Model: "gpt-4",
		State: "running",
	}

	json.NewEncoder(w).Encode(agent)
}

func (h *AgentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		Name    string `json:"name"`
		State   string `json:"state"`
		Model   string `json:"model"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	h.logger.Info().Str("agent_id", id).Msg("agent updated")

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "updated",
	})
}

func (h *AgentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	h.logger.Info().Str("agent_id", id).Msg("agent deleted")

	w.WriteHeader(http.StatusNoContent)
}
