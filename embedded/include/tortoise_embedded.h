/**
 * Tortoise Embedded SDK
 * 
 * 用于嵌入式设备的轻量级 AI 代理 SDK
 */

#ifndef TORTOISE_EMBEDDED_H
#define TORTOISE_EMBEDDED_H

#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

// === 版本信息 ===

#define TORTOISE_EMBEDDED_VERSION_MAJOR 0
#define TORTOISE_EMBEDDED_VERSION_MINOR 1
#define TORTOISE_EMBEDDED_VERSION_PATCH 0

// === 错误码 ===

typedef enum {
    TORTOISE_OK = 0,
    TORTOISE_ERROR_INVALID_PARAM = -1,
    TORTOISE_ERROR_OUT_OF_MEMORY = -2,
    TORTOISE_ERROR_NOT_INITIALIZED = -3,
    TORTOISE_ERROR_ALREADY_INITIALIZED = -4,
    TORTOISE_ERROR_CONNECTION_FAILED = -5,
    TORTOISE_ERROR_TIMEOUT = -6,
    TORTOISE_ERROR_NOT_FOUND = -7,
    TORTOISE_ERROR_PERMISSION_DENIED = -8,
    TORTOISE_ERROR_UNKNOWN = -99,
} tortoise_error_t;

// === 配置 ===

typedef struct {
    // 服务器地址
    const char* server_host;
    uint16_t server_port;
    
    // 连接超时 (毫秒)
    uint32_t connect_timeout_ms;
    
    // 读写超时 (毫秒)
    uint32_t read_timeout_ms;
    uint32_t write_timeout_ms;
    
    // 缓冲区大小
    size_t buffer_size;
    
    // 调试模式
    bool debug_mode;
    
    // SSL/TLS
    bool use_ssl;
    const char* ca_cert_path;
    
    // 自动重连
    bool auto_reconnect;
    uint32_t reconnect_interval_ms;
    
    // 日志回调
    void (*log_callback)(int level, const char* message);
} tortoise_config_t;

// === 客户端 ===

typedef struct tortoise_client tortoise_client_t;

// === 消息结构 ===

typedef struct {
    const char* content;
    size_t content_length;
    const char* sender_id;
    const char* channel;
} tortoise_message_t;

// === API 函数 ===

/**
 * 创建客户端实例
 * @param config 配置
 * @param client 输出客户端指针
 * @return 错误码
 */
tortoise_error_t tortoise_client_create(
    const tortoise_config_t* config,
    tortoise_client_t** client
);

/**
 * 销毁客户端实例
 * @param client 客户端指针
 * @return 错误码
 */
tortoise_error_t tortoise_client_destroy(
    tortoise_client_t* client
);

/**
 * 连接到服务器
 * @param client 客户端指针
 * @return 错误码
 */
tortoise_error_t tortoise_client_connect(
    tortoise_client_t* client
);

/**
 * 断开连接
 * @param client 客户端指针
 * @return 错误码
 */
tortoise_error_t tortoise_client_disconnect(
    tortoise_client_t* client
);

/**
 * 检查连接状态
 * @param client 客户端指针
 * @param connected 输出连接状态
 * @return 错误码
 */
tortoise_error_t tortoise_client_is_connected(
    tortoise_client_t* client,
    bool* connected
);

/**
 * 发送消息
 * @param client 客户端指针
 * @param message 消息
 * @param response 输出响应
 * @param timeout_ms 超时时间
 * @return 错误码
 */
tortoise_error_t tortoise_client_send_message(
    tortoise_client_t* client,
    const tortoise_message_t* message,
    tortoise_message_t* response,
    uint32_t timeout_ms
);

/**
 * 发送消息 (异步)
 * @param client 客户端指针
 * @param message 消息
 * @param callback 回调函数
 * @return 错误码
 */
typedef void (*tortoise_async_callback_t)(
    tortoise_error_t error,
    const tortoise_message_t* response,
    void* user_data
);

tortoise_error_t tortoise_client_send_message_async(
    tortoise_client_t* client,
    const tortoise_message_t* message,
    tortoise_async_callback_t callback,
    void* user_data
);

/**
 * 获取最后错误信息
 * @param client 客户端指针
 * @return 错误信息
 */
const char* tortoise_client_get_error_message(
    tortoise_client_t* client
);

/**
 * 设置连接状态回调
 * @param client 客户端指针
 * @param callback 回调函数
 * @param user_data 用户数据
 * @return 错误码
 */
typedef void (*tortoise_connection_callback_t)(
    bool connected,
    void* user_data
);

tortoise_error_t tortoise_client_set_connection_callback(
    tortoise_client_t* client,
    tortoise_connection_callback_t callback,
    void* user_data
);

/**
 * 获取 SDK 版本
 */
void tortoise_get_version(
    int* major,
    int* minor,
    int* patch
);

// === 工具函数 ===

/**
 * 初始化默认配置
 * @param config 输出配置
 * @return 错误码
 */
tortoise_error_t tortoise_config_init_default(
    tortoise_config_t* config
);

/**
 * 获取错误信息
 * @param error 错误码
 * @return 错误信息
 */
const char* tortoise_error_to_string(
    tortoise_error_t error
);

#ifdef __cplusplus
}
#endif

#endif // TORTOISE_EMBEDDED_H
