// Tortoise Framework - Go Usage Examples

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/tortoise/cloud/internal/api"
	"github.com/tortoise/cloud/internal/auth"
)

func main() {
	fmt.Println("Tortoise Framework - Go Examples")
	fmt.Println("================================\n")

	// Example 1: Using the API client
	apiExample()

	// Example 2: Authentication
	authExample()
}

func apiExample() {
	fmt.Println("Example 1: API Client")
	fmt.Println("-------------------")

	// Create API handler
	handler := api.NewHandler(&api.HandlerConfig{
		Auth:    nil,
		VectorDB: nil,
		Logger:   zerolog.New(os.Stdout),
	})

	// Work with agents
	agents := []api.AgentInfo{
		{ID: "agent-1", Name: "assistant", State: "running"},
		{ID: "agent-2", Name: "coder", State: "paused"},
	}

	agentsJSON, _ := json.MarshalIndent(agents, "", "  ")
	fmt.Printf("Agents: %s\n", agentsJSON)

	// Memory operations
	memory := api.MemoryEntry{
		Key:   "user-preferences",
		Value: map[string]interface{}{"theme": "dark"},
		Type:  "semantic",
	}

	memoryJSON, _ := json.MarshalIndent(memory, "", "  ")
	fmt.Printf("Memory: %s\n", memoryJSON)

	fmt.Println()
}

func authExample() {
	fmt.Println("Example 2: Authentication")
	fmt.Println("----------------------")

	// Create auth service
	cfg := auth.Config{
		JWTSecret: "your-secret-key",
	}

	svc, err := auth.NewService(cfg)
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}

	// Register user
	user, err := svc.Register("user@example.com", "password123")
	if err != nil {
		log.Fatalf("Failed to register: %v", err)
	}
	fmt.Printf("Registered user: %s\n", user.Email)

	// Login
	token, err := svc.Login("user@example.com", "password123")
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	fmt.Printf("Token: %s\n", token[:20]+"...")

	// Validate token
	claims, err := svc.ValidateToken(token)
	if err != nil {
		log.Fatalf("Failed to validate: %v", err)
	}
	fmt.Printf("Validated user ID: %s\n", claims.UserID)

	fmt.Println()
}

import (
	"github.com/rs/zerolog"
	"os"
)
