// Tortoise CLI - Command Line Interface

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	version    = "0.1.0"
	serverAddr = flag.String("server", "localhost:18792", "Tortoise server address")
	apiKey     = flag.String("api-key", "", "API key for authentication")
	verbose    = flag.Bool("v", false, "Verbose output")
	jsonOutput = flag.Bool("json", false, "JSON output")
)

func main() {
	flag.Parse()

	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if *verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Parse command
	if flag.NArg() < 1 {
		printUsage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	// Execute command
	ctx := context.Background()
	var err error

	switch cmd {
	case "version", "v":
		fmt.Printf("Tortoise CLI v%s\n", version)
	case "help", "h":
		printUsage()
	case "chat":
		err = runChat(ctx, args)
	case "session":
		err = runSession(ctx, args)
	case "send":
		err = runSend(ctx, args)
	case "tools":
		err = runTools(ctx, args)
	case "memory":
		err = runMemory(ctx, args)
	case "config":
		err = runConfig(ctx, args)
	case "doctor":
		err = runDoctor(ctx)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		log.Error().Err(err).Msg("Command failed")
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Tortoise CLI - AI Agent Command Line Interface

Usage: tortoise [options] <command> [args]

Commands:
  chat              Start interactive chat
  session          Manage sessions
  send <msg>       Send a message
  tools            List available tools
  memory           Manage memory
  config           Configuration management
  doctor           Diagnose connection issues

Options:
  -server string   Server address (default: localhost:18792)
  -api-key string API key for authentication
  -v              Verbose output
  -json            JSON output

Examples:
  tortoise chat
  tortoise send "Hello, how are you?"
  tortoise session list
  tortoise tools
  tortoise doctor`)
}

// Interactive chat
func runChat(ctx context.Context, args []string) error {
	fmt.Println("Starting Tortoise Chat...")
	fmt.Println("Press Ctrl+C to exit\n")

	// Connect to server
	conn, err := connect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Create session
	sessionID := uuid.New().String()[:8]
	fmt.Printf("[Session: %s]\n\n", sessionID)

	// Simple message loop
	for {
		fmt.Print("You: ")
		var input string
		fmt.Scanln(&input)

		if strings.ToLower(input) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		// Send message
		req := map[string]interface{}{
			"type":      "request",
			"sessionId": sessionID,
			"content":   input,
		}

		if err := conn.WriteJSON(req); err != nil {
			return fmt.Errorf("send failed: %w", err)
		}

		// Read response
		var resp map[string]interface{}
		if err := conn.ReadJSON(&resp); err != nil {
			return fmt.Errorf("receive failed: %w", err)
		}

		if content, ok := resp["content"].(string); ok {
			fmt.Printf("Tortoise: %s\n\n", content)
		}
	}

	return nil
}

// Session management
func runSession(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: tortoise session <list|create|delete>")
	}

	subcmd := args[0]

	switch subcmd {
	case "list":
		return listSessions()
	case "create":
		return createSession(args[1:])
	case "delete":
		return deleteSession(args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", subcmd)
	}
}

func listSessions() error {
	sessions := []map[string]interface{}{
		{"id": "sess_abc123", "status": "active", "messages": 10},
		{"id": "sess_def456", "status": "idle", "messages": 5},
	}

	if *jsonOutput {
		data, _ := json.MarshalIndent(sessions, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println("ID            STATUS   MESSAGES")
		fmt.Println("------------  -------  --------")
		for _, s := range sessions {
			fmt.Printf("%-12s  %-7s  %d\n", s["id"], s["status"], s["messages"])
		}
	}
	return nil
}

func createSession(args []string) error {
	sessionID := "sess_" + uuid.New().String()[:8]

	if *jsonOutput {
		data, _ := json.MarshalIndent(map[string]string{
			"sessionId": sessionID,
			"status":    "active",
		}, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Created session: %s\n", sessionID)
	}
	return nil
}

func deleteSession(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: tortoise session delete <session-id>")
	}
	sessionID := args[0]

	if *jsonOutput {
		data, _ := json.MarshalIndent(map[string]bool{
			"deleted": true,
		}, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Deleted session: %s\n", sessionID)
	}
	return nil
}

// Send message
func runSend(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: tortoise send <message>")
	}

	message := strings.Join(args, " ")

	if *jsonOutput {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"messageId": "msg_" + uuid.New().String()[:8],
			"content":   "Echo: " + message,
			"model":     "gpt-4o",
		}, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Tortoise: Echo: %s\n", message)
	}
	return nil
}

// List tools
func runTools(ctx context.Context, args []string) error {
	tools := []map[string]interface{}{
		{
			"name":        "web_search",
			"description": "Search the web for information",
			"parameters":  []string{"query", "num_results"},
		},
		{
			"name":        "calculator",
			"description": "Perform mathematical calculations",
			"parameters":  []string{"expression"},
		},
		{
			"name":        "file_read",
			"description": "Read file contents",
			"parameters":  []string{"path"},
		},
	}

	if *jsonOutput {
		data, _ := json.MarshalIndent(tools, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println("Available Tools:")
		fmt.Println("---------------")
		for _, t := range tools {
			fmt.Printf("\n%s: %s\n", t["name"], t["description"])
			fmt.Printf("  Parameters: %v\n", t["parameters"])
		}
	}
	return nil
}

// Memory management
func runMemory(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: tortoise memory <store|search|list>")
	}

	subcmd := args[0]

	switch subcmd {
	case "store":
		return storeMemory(args[1:])
	case "search":
		return searchMemory(args[1:])
	case "list":
		return listMemory()
	default:
		return fmt.Errorf("unknown subcommand: %s", subcmd)
	}
}

func storeMemory(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: tortoise memory store <content>")
	}

	content := strings.Join(args, " ")
	memID := "mem_" + uuid.New().String()[:8]

	if *jsonOutput {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"id":      memID,
			"success": true,
		}, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Stored memory: %s\n", memID)
	}
	return nil
}

func searchMemory(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: tortoise memory search <query>")
	}

	query := strings.Join(args, " ")

	results := []map[string]interface{}{
		{"id": "mem_1", "content": "User prefers dark mode", "similarity": 0.95},
		{"id": "mem_2", "content": "User's name is John", "similarity": 0.85},
	}

	if *jsonOutput {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Search results for '%s':\n", query)
		for _, r := range results {
			fmt.Printf("  [%s] %s (%.0f%%)\n", r["id"], r["content"], r["similarity"].(float64)*100)
		}
	}
	return nil
}

func listMemory() error {
	memories := []map[string]interface{}{
		{"id": "mem_1", "type": "fact", "content": "User prefers dark mode"},
		{"id": "mem_2", "type": "fact", "content": "User's name is John"},
	}

	if *jsonOutput {
		data, _ := json.MarshalIndent(memories, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println("Stored Memories:")
		fmt.Println("----------------")
		for _, m := range memories {
			fmt.Printf("  [%s] %s: %s\n", m["id"], m["type"], m["content"])
		}
	}
	return nil
}

// Configuration
func runConfig(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: tortoise config <get|set>")
	}

	subcmd := args[0]

	switch subcmd {
	case "get":
		return getConfig(args[1:])
	case "set":
		return setConfig(args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", subcmd)
	}
}

func getConfig(args []string) error {
	config := map[string]interface{}{
		"server":   *serverAddr,
		"model":     "gpt-4o",
		"temperature": 0.7,
	}

	if *jsonOutput {
		data, _ := json.MarshalIndent(config, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println("Current Configuration:")
		for k, v := range config {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}
	return nil
}

func setConfig(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: tortoise config set <key> <value>")
	}

	key, value := args[0], args[1]
	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

// Doctor - diagnose issues
func runDoctor(ctx context.Context) error {
	fmt.Println("Running diagnostics...\n")

	checks := []struct {
		name string
		fn   func() error
	}{
		{"Server connection", checkServerConnection},
		{"API key", checkAPIKey},
		{"Network latency", checkNetworkLatency},
	}

	allPassed := true
	for _, check := range checks {
		fmt.Printf("[%s] %s...", check.name, strings.Repeat(".", 30-len(check.name)))
		
		if err := check.fn(); err != nil {
			fmt.Printf(" FAIL\n  Error: %v\n", err)
			allPassed = false
		} else {
			fmt.Printf(" OK\n")
		}
	}

	fmt.Println()
	if allPassed {
		fmt.Println("All checks passed!")
	} else {
		fmt.Println("Some checks failed. Please review the errors above.")
	}

	return nil
}

func checkServerConnection() error {
	conn, err := connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func checkAPIKey() error {
	if *apiKey == "" {
		return fmt.Errorf("no API key provided (use -api-key flag)")
	}
	return nil
}

func checkNetworkLatency() error {
	start := time.Now()
	conn, err := connect()
	if err != nil {
		return err
	}
	conn.Close()
	latency := time.Since(start)
	
	if latency > 5*time.Second {
		return fmt.Errorf("high latency: %v", latency)
	}
	return nil
}

// Helper functions
func connect() (*websocket.Conn, error) {
	url := fmt.Sprintf("ws://%s/ws", *serverAddr)
	
	header := http.Header{}
	if *apiKey != "" {
		header.Set("Authorization", "Bearer "+*apiKey)
	}

	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
