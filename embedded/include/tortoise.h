/**
 * @file tortoise.h
 * @brief Tortoise Embedded C SDK
 * 
 * A lightweight C SDK for embedded devices and IoT integration.
 */

#ifndef TORTOISE_H
#define TORTOISE_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stdbool.h>
#include <stddef.h>

// Version
#define TORTOISE_VERSION_MAJOR 0
#define TORTOISE_VERSION_MINOR 1
#define TORTOISE_VERSION_PATCH 0

// Error codes
typedef enum {
    TORTOISE_OK = 0,
    TORTOISE_ERROR_INVALID_PARAM = -1,
    TORTOISE_ERROR_NO_MEMORY = -2,
    TORTOISE_ERROR_NOT_CONNECTED = -3,
    TORTOISE_ERROR_TIMEOUT = -4,
    TORTOISE_ERROR_PROTOCOL = -5,
    TORTOISE_ERROR_AUTH = -6,
} tortoise_error_t;

// Connection state
typedef enum {
    TORTOISE_STATE_DISCONNECTED = 0,
    TORTOISE_STATE_CONNECTING = 1,
    TORTOISE_STATE_CONNECTED = 2,
    TORTOISE_STATE_ERROR = 3,
} tortoise_state_t;

// Message role
typedef enum {
    TORTOISE_ROLE_SYSTEM = 0,
    TORTOISE_ROLE_USER = 1,
    TORTOISE_ROLE_ASSISTANT = 2,
} tortoise_role_t;

// Thinking mode
typedef enum {
    TORTOISE_THINK_FAST = 0,
    TORTOISE_THINK_BALANCED = 1,
    TORTOISE_THINK_DEEP = 2,
    TORTOISE_THINK_RESEARCH = 3,
    TORTOISE_THINK_CREATIVE = 4,
} tortoise_think_mode_t;

// Configuration
typedef struct {
    char* gateway_url;
    char* api_key;
    char* device_id;
    uint16_t port;
    uint32_t timeout_ms;
    size_t max_memory_kb;
    bool auto_reconnect;
    void (*log_callback)(const char* level, const char* message);
} tortoise_config_t;

// Message
typedef struct {
    tortoise_role_t role;
    char* content;
    size_t content_len;
} tortoise_message_t;

// Response
typedef struct {
    char* content;
    size_t content_len;
    tortoise_error_t error_code;
    char* error_message;
} tortoise_response_t;

// Memory item
typedef struct {
    char* id;
    char* content;
    float importance;
    uint64_t created_at;
} tortoise_memory_t;

// Context handle
typedef void* tortoise_ctx_t;

/**
 * @brief Initialize the Tortoise SDK
 * @param config Configuration
 * @return Context handle or NULL on error
 */
tortoise_ctx_t tortoise_init(tortoise_config_t* config);

/**
 * @brief Destroy the Tortoise SDK context
 * @param ctx Context handle
 */
void tortoise_destroy(tortoise_ctx_t ctx);

/**
 * @brief Connect to the Tortoise gateway
 * @param ctx Context handle
 * @return TORTOISE_OK on success
 */
tortoise_error_t tortoise_connect(tortoise_ctx_t ctx);

/**
 * @brief Disconnect from the Tortoise gateway
 * @param ctx Context handle
 * @return TORTOISE_OK on success
 */
tortoise_error_t tortoise_disconnect(tortoise_ctx_t ctx);

/**
 * @brief Get connection state
 * @param ctx Context handle
 * @return Current state
 */
tortoise_state_t tortoise_get_state(tortoise_ctx_t ctx);

/**
 * @brief Send a chat message
 * @param ctx Context handle
 * @param messages Array of messages
 * @param message_count Number of messages
 * @param mode Thinking mode
 * @param response Output response
 * @return TORTOISE_OK on success
 */
tortoise_error_t tortoise_chat(
    tortoise_ctx_t ctx,
    tortoise_message_t* messages,
    size_t message_count,
    tortoise_think_mode_t mode,
    tortoise_response_t* response
);

/**
 * @brief Free a response
 * @param response Response to free
 */
void tortoise_response_free(tortoise_response_t* response);

/**
 * @brief Remember information
 * @param ctx Context handle
 * @param content Content to remember
 * @param importance Importance (0.0 - 1.0)
 * @return Memory ID or NULL on error
 */
char* tortoise_remember(tortoise_ctx_t ctx, const char* content, float importance);

/**
 * @brief Recall memories
 * @param ctx Context handle
 * @param query Query string
 * @param memories Output array
 * @param max_count Maximum memories to return
 * @return Number of memories returned
 */
size_t tortoise_recall(
    tortoise_ctx_t ctx,
    const char* query,
    tortoise_memory_t* memories,
    size_t max_count
);

/**
 * @brief Get memory statistics
 * @param ctx Context handle
 * @param short_term Pointer to store short-term count
 * @param medium_term Pointer to store medium-term count
 * @param long_term Pointer to store long-term count
 * @return TORTOISE_OK on success
 */
tortoise_error_t tortoise_memory_stats(
    tortoise_ctx_t ctx,
    size_t* short_term,
    size_t* medium_term,
    size_t* long_term
);

/**
 * @brief Get SDK version string
 * @return Version string
 */
const char* tortoise_version(void);

#ifdef __cplusplus
}
#endif

#endif /* TORTOISE_H */
