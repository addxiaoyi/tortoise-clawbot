package main

import (
	"context"
	"fmt"
	"log"

	tortoise "github.com/tortoise/sdk-go"
)

func main() {
	// Create client
	client := tortoise.NewClient(
		tortoise.WithAPIKey("your-api-key"),
		tortoise.WithBaseURL("http://localhost:18792"),
	)

	ctx := context.Background()

	// Connect
	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect()

	// Create session
	session, err := client.Sessions.Create(ctx, &tortoise.CreateSessionRequest{
		UserID: "user@example.com",
		Config: &tortoise.SessionConfig{
			Model:       "gpt-4o",
			Temperature: 0.7,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Session: %s\n", session.ID)

	// Send message
	resp, err := client.Messages.Send(ctx, session.ID, &tortoise.MessageRequest{
		Content: "Hello!",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response: %s\n", resp.Content)

	// List tools
	tools, err := client.Tools.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Tools: %v\n", tools)

	// Invoke tool
	result, err := client.Tools.Invoke(ctx, "calculator", &tortoise.InvokeToolRequest{
		Arguments: map[string]interface{}{
			"expression": "42 * 2",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Calculator result: %v\n", result.Result)
}
