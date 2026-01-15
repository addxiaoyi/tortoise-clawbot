package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"

	"github.com/tortoise/cloud/internal/api"
	"github.com/tortoise/cloud/internal/auth"
	"github.com/tortoise/cloud/internal/config"
	"github.com/tortoise/cloud/internal/vector"
)

func main() {
	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Load configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	// Initialize auth
	authService, err := auth.NewService(cfg.Auth)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize auth")
	}

	// Initialize vector DB
	vectorDB, err := vector.NewClient(cfg.VectorDB)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize vector DB")
	}

	// Initialize API handlers
	apiHandler := api.NewHandler(&api.HandlerConfig{
		Auth:      authService,
		VectorDB:  vectorDB,
		Logger:    logger,
	})

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Server.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Health check
		r.Get("/health", apiHandler.Health)

		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", apiHandler.Auth.Register)
			r.Post("/login", apiHandler.Auth.Login)
			r.Post("/logout", apiHandler.Auth.Logout)
			r.Post("/refresh", apiHandler.Auth.Refresh)
			r.Get("/me", apiHandler.Auth.Me)
		})

		// Agent routes (protected)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware(authService))

			// Agents
			r.Route("/agents", func(r chi.Router) {
				r.Get("/", apiHandler.Agents.List)
				r.Post("/", apiHandler.Agents.Create)
				r.Get("/{id}", apiHandler.Agents.Get)
				r.Put("/{id}", apiHandler.Agents.Update)
				r.Delete("/{id}", apiHandler.Agents.Delete)
			})

			// Memory
			r.Route("/memory", func(r chi.Router) {
				r.Get("/", apiHandler.Memory.List)
				r.Post("/", apiHandler.Memory.Store)
				r.Get("/{key}", apiHandler.Memory.Get)
				r.Delete("/{key}", apiHandler.Memory.Delete)
				r.Post("/search", apiHandler.Memory.Search)
			})

			// Vector search
			r.Route("/vector", func(r chi.Router) {
				r.Post("/embed", apiHandler.Vector.Embed)
				r.Post("/search", apiHandler.Vector.Search)
			})

			// Mesh
			r.Route("/mesh", func(r chi.Router) {
				r.Get("/nodes", apiHandler.Mesh.ListNodes)
				r.Post("/connect", apiHandler.Mesh.Connect)
				r.Post("/delegate", apiHandler.Mesh.Delegate)
			})
		})
	})

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Info().
			Str("host", cfg.Server.Host).
			Int("port", cfg.Server.Port).
			Msg("starting server")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server error")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("server forced to shutdown")
	}

	logger.Info().Msg("server stopped")
}

func authMiddleware(authSvc *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := authSvc.ValidateToken(token)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
