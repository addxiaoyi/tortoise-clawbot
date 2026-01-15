/**
 * @file tortoise.c
 * @brief Tortoise Embedded C SDK Implementation
 */

#include "tortoise.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

// Internal context structure
typedef struct {
    tortoise_config_t config;
    tortoise_state_t state;
    int socket_fd;
} tortoise_context_t;

const char* tortoise_version(void) {
    return "0.1.0";
}

tortoise_ctx_t tortoise_init(tortoise_config_t* config) {
    if (!config || !config->gateway_url || !config->api_key) {
        return NULL;
    }

    tortoise_context_t* ctx = (tortoise_context_t*)malloc(sizeof(tortoise_context_t));
    if (!ctx) {
        return NULL;
    }

    memset(ctx, 0, sizeof(tortoise_context_t));
    
    // Copy configuration
    ctx->config.gateway_url = strdup(config->gateway_url);
    ctx->config.api_key = strdup(config->api_key);
    ctx->config.port = config->port ? config->port : 18789;
    ctx->config.timeout_ms = config->timeout_ms ? config->timeout_ms : 30000;
    ctx->config.auto_reconnect = config->auto_reconnect;
    ctx->config.log_callback = config->log_callback;

    if (config->device_id) {
        ctx->config.device_id = strdup(config->device_id);
    } else {
        // Generate default device ID
        ctx->config.device_id = strdup("embedded-device");
    }

    ctx->state = TORTOISE_STATE_DISCONNECTED;
    ctx->socket_fd = -1;

    return (tortoise_ctx_t)ctx;
}

void tortoise_destroy(tortoise_ctx_t ctx) {
    if (!ctx) return;
    
    tortoise_context_t* c = (tortoise_context_t*)ctx;
    
    tortoise_disconnect(ctx);
    
    if (c->config.gateway_url) free(c->config.gateway_url);
    if (c->config.api_key) free(c->config.api_key);
    if (c->config.device_id) free(c->config.device_id);
    
    free(c);
}

tortoise_error_t tortoise_connect(tortoise_ctx_t ctx) {
    if (!ctx) return TORTOISE_ERROR_INVALID_PARAM;
    
    tortoise_context_t* c = (tortoise_context_t*)ctx;
    
    if (c->state == TORTOISE_STATE_CONNECTED) {
        return TORTOISE_OK;
    }

    c->state = TORTOISE_STATE_CONNECTING;
    
    // Placeholder - would establish actual socket connection
    // For embedded devices, this would use MQTT, WebSocket, or HTTP
    c->state = TORTOISE_STATE_CONNECTED;
    c->socket_fd = 0;
    
    if (c->config.log_callback) {
        c->config.log_callback("INFO", "Connected to Tortoise gateway");
    }
    
    return TORTOISE_OK;
}

tortoise_error_t tortoise_disconnect(tortoise_ctx_t ctx) {
    if (!ctx) return TORTOISE_ERROR_INVALID_PARAM;
    
    tortoise_context_t* c = (tortoise_context_t*)ctx;
    
    if (c->socket_fd >= 0) {
        // Close socket
        c->socket_fd = -1;
    }
    
    c->state = TORTOISE_STATE_DISCONNECTED;
    
    if (c->config.log_callback) {
        c->config.log_callback("INFO", "Disconnected from Tortoise gateway");
    }
    
    return TORTOISE_OK;
}

tortoise_state_t tortoise_get_state(tortoise_ctx_t ctx) {
    if (!ctx) return TORTOISE_STATE_DISCONNECTED;
    return ((tortoise_context_t*)ctx)->state;
}

tortoise_error_t tortoise_chat(
    tortoise_ctx_t ctx,
    tortoise_message_t* messages,
    size_t message_count,
    tortoise_think_mode_t mode,
    tortoise_response_t* response
) {
    if (!ctx || !messages || !response) {
        return TORTOISE_ERROR_INVALID_PARAM;
    }
    
    tortoise_context_t* c = (tortoise_context_t*)ctx;
    
    if (c->state != TORTOISE_STATE_CONNECTED) {
        return TORTOISE_ERROR_NOT_CONNECTED;
    }

    // Initialize response
    memset(response, 0, sizeof(tortoise_response_t));
    
    // Placeholder - would send actual request to gateway
    // Format: JSON over socket/HTTP/MQTT
    
    response->content = strdup("Tortoise embedded response placeholder");
    response->content_len = strlen(response->content);
    response->error_code = TORTOISE_OK;
    
    if (c->config.log_callback) {
        c->config.log_callback("DEBUG", "Chat request sent");
    }
    
    return TORTOISE_OK;
}

void tortoise_response_free(tortoise_response_t* response) {
    if (!response) return;
    
    if (response->content) {
        free(response->content);
        response->content = NULL;
    }
    
    if (response->error_message) {
        free(response->error_message);
        response->error_message = NULL;
    }
}

char* tortoise_remember(tortoise_ctx_t ctx, const char* content, float importance) {
    if (!ctx || !content) return NULL;
    
    tortoise_context_t* c = (tortoise_context_t*)ctx;
    
    if (c->state != TORTOISE_STATE_CONNECTED) {
        return NULL;
    }
    
    // Placeholder - would send remember request to gateway
    // Return a generated memory ID
    char* id = (char*)malloc(64);
    snprintf(id, 64, "mem_%lu", (unsigned long)time(NULL));
    
    return id;
}

size_t tortoise_recall(
    tortoise_ctx_t ctx,
    const char* query,
    tortoise_memory_t* memories,
    size_t max_count
) {
    if (!ctx || !query || !memories) return 0;
    
    tortoise_context_t* c = (tortoise_context_t*)ctx;
    
    if (c->state != TORTOISE_STATE_CONNECTED) {
        return 0;
    }
    
    // Placeholder - would query gateway for memories
    // Return empty results for now
    (void)memories;
    (void)max_count;
    
    return 0;
}

tortoise_error_t tortoise_memory_stats(
    tortoise_ctx_t ctx,
    size_t* short_term,
    size_t* medium_term,
    size_t* long_term
) {
    if (!ctx) return TORTOISE_ERROR_INVALID_PARAM;
    
    tortoise_context_t* c = (tortoise_context_t*)ctx;
    
    if (c->state != TORTOISE_STATE_CONNECTED) {
        return TORTOISE_ERROR_NOT_CONNECTED;
    }
    
    // Placeholder - would query gateway for memory stats
    if (short_term) *short_term = 0;
    if (medium_term) *medium_term = 0;
    if (long_term) *long_term = 0;
    
    return TORTOISE_OK;
}
