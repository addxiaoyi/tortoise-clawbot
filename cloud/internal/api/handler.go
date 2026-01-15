package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/tortoise/cloud/internal/auth"
	"github.com/tortoise/cloud/internal/vector"
)

type HandlerConfig struct {
	Auth     *auth.Service
	VectorDB *vector.Client
	Logger   zerolog.Logger
}

type Handler struct {
	config HandlerConfig
	Auth   *AuthHandler
	Agents *AgentHandler
	Memory *MemoryHandler
	Vector *VectorHandler
	Mesh   *MeshHandler
}

func NewHandler(cfg *HandlerConfig) *Handler {
	h := &Handler{config: *cfg}

	h.Auth = &AuthHandler{service: cfg.Auth}
	h.Agents = &AgentHandler{logger: cfg.Logger}
	h.Memory = &MemoryHandler{logger: cfg.Logger}
	h.Vector = &VectorHandler{client: cfg.VectorDB}
	h.Mesh = &MeshHandler{logger: cfg.Logger}

	return h
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	json.NewResponse(w).OK(map[string]string{
		"status":  "healthy",
		"version": "0.1.0",
	})
}

// Response helper
type Response struct{}

func (Response) OK(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func (Response) Created(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}

func (Response) Error(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

import (
	"encoding/json"
	"net/http"
)
