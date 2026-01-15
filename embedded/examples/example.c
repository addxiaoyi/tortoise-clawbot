/**
 * @file example.c
 * @brief Tortoise Embedded SDK Example
 */

#include "tortoise.h"
#include <stdio.h>
#include <string.h>

void log_callback(const char* level, const char* message) {
    printf("[%s] %s\n", level, message);
}

int main() {
    printf("Tortoise Embedded SDK v%s\n", tortoise_version());
    
    // Initialize configuration
    tortoise_config_t config = {
        .gateway_url = "http://localhost:18789",
        .api_key = "your-api-key",
        .device_id = "my-device",
        .port = 18789,
        .timeout_ms = 30000,
        .max_memory_kb = 1024,
        .auto_reconnect = true,
        .log_callback = log_callback
    };
    
    // Initialize SDK
    tortoise_ctx_t ctx = tortoise_init(&config);
    if (!ctx) {
        fprintf(stderr, "Failed to initialize Tortoise SDK\n");
        return 1;
    }
    
    printf("SDK initialized successfully\n");
    
    // Connect to gateway
    tortoise_error_t err = tortoise_connect(ctx);
    if (err != TORTOISE_OK) {
        fprintf(stderr, "Failed to connect: %d\n", err);
        tortoise_destroy(ctx);
        return 1;
    }
    
    printf("Connected to gateway\n");
    
    // Check state
    tortoise_state_t state = tortoise_get_state(ctx);
    printf("Connection state: %d\n", state);
    
    // Send chat message
    tortoise_message_t messages[] = {
        { TORTOISE_ROLE_USER, "Hello, Tortoise!", 14 }
    };
    
    tortoise_response_t response;
    err = tortoise_chat(ctx, messages, 1, TORTOISE_THINK_BALANCED, &response);
    
    if (err == TORTOISE_OK) {
        printf("Response: %s\n", response.content);
        tortoise_response_free(&response);
    } else {
        fprintf(stderr, "Chat failed: %d\n", err);
    }
    
    // Remember something
    char* memory_id = tortoise_remember(ctx, "This is important data", 0.9f);
    if (memory_id) {
        printf("Remembered with ID: %s\n", memory_id);
        free(memory_id);
    }
    
    // Get memory stats
    size_t short_term, medium_term, long_term;
    err = tortoise_memory_stats(ctx, &short_term, &medium_term, &long_term);
    if (err == TORTOISE_OK) {
        printf("Memory - Short: %zu, Medium: %zu, Long: %zu\n", 
               short_term, medium_term, long_term);
    }
    
    // Disconnect
    tortoise_disconnect(ctx);
    printf("Disconnected\n");
    
    // Cleanup
    tortoise_destroy(ctx);
    printf("Cleanup complete\n");
    
    return 0;
}
