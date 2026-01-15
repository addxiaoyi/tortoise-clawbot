/**
 * Tortoise Embedded C++ Wrapper
 */

#ifndef TORTOISE_EMBEDDED_HPP
#define TORTOISE_EMBEDDED_HPP

#include "tortoise_embedded.h"
#include <string>
#include <functional>
#include <memory>
#include <stdexcept>

namespace tortoise {

/**
 * 异常类
 */
class TortoiseException : public std::runtime_error {
public:
    TortoiseException(tortoise_error_t error, const std::string& message)
        : std::runtime_error(message), error_(error) {}
    
    tortoise_error_t error() const { return error_; }
    
private:
    tortoise_error_t error_;
};

/**
 * 消息类
 */
class Message {
public:
    Message() : impl_() {}
    
    Message(const std::string& content, 
            const std::string& sender = "",
            const std::string& channel = "")
        : impl_({content.c_str(), content.size(), 
                  sender.empty() ? nullptr : sender.c_str(),
                  channel.empty() ? nullptr : channel.c_str()}) {}
    
    const std::string& content() const { return content_; }
    const std::string& sender() const { return sender_; }
    const std::string& channel() const { return channel_; }
    
    tortoise_message_t* native() { return &impl_; }
    const tortoise_message_t* native() const { return &impl_; }
    
private:
    std::string content_;
    std::string sender_;
    std::string channel_;
    tortoise_message_t impl_;
};

/**
 * 配置类
 */
class Config {
public:
    Config() {
        tortoise_config_init_default(&impl_);
    }
    
    Config& setServerHost(const std::string& host) {
        impl_.server_host = host.c_str();
        return *this;
    }
    
    Config& setServerPort(uint16_t port) {
        impl_.server_port = port;
        return *this;
    }
    
    Config& setConnectTimeout(uint32_t ms) {
        impl_.connect_timeout_ms = ms;
        return *this;
    }
    
    Config& setReadTimeout(uint32_t ms) {
        impl_.read_timeout_ms = ms;
        return *this;
    }
    
    Config& setWriteTimeout(uint32_t ms) {
        impl_.write_timeout_ms = ms;
        return *this;
    }
    
    Config& setBufferSize(size_t size) {
        impl_.buffer_size = size;
        return *this;
    }
    
    Config& setDebugMode(bool enable) {
        impl_.debug_mode = enable;
        return *this;
    }
    
    Config& setSSL(bool enable) {
        impl_.use_ssl = enable;
        return *this;
    }
    
    Config& setCACertPath(const std::string& path) {
        impl_.ca_cert_path = path.c_str();
        return *this;
    }
    
    Config& setAutoReconnect(bool enable) {
        impl_.auto_reconnect = enable;
        return *this;
    }
    
    Config& setReconnectInterval(uint32_t ms) {
        impl_.reconnect_interval_ms = ms;
        return *this;
    }
    
    Config& setLogCallback(std::function<void(int, const std::string&)> callback) {
        log_callback_ = callback;
        impl_.log_callback = [](int level, const char* msg) {};
        return *this;
    }
    
    tortoise_config_t* native() { return &impl_; }
    
private:
    tortoise_config_t impl_;
    std::function<void(int, const std::string&)> log_callback_;
};

/**
 * 客户端类
 */
class Client {
public:
    Client() : client_(nullptr), owned_(false) {}
    
    Client(const Config& config) : client_(nullptr), owned_(true) {
        tortoise_error_t err = tortoise_client_create(
            config.native(), &client_);
        if (err != TORTOISE_OK) {
            throw TortoiseException(err, 
                tortoise_error_to_string(err));
        }
    }
    
    ~Client() {
        if (owned_ && client_) {
            tortoise_client_destroy(client_);
        }
    }
    
    // 禁用拷贝
    Client(const Client&) = delete;
    Client& operator=(const Client&) = delete;
    
    // 启用移动
    Client(Client&& other) noexcept 
        : client_(other.client_), owned_(other.owned_) {
        other.client_ = nullptr;
        other.owned_ = false;
    }
    
    Client& operator=(Client&& other) noexcept {
        if (this != &other) {
            if (owned_ && client_) {
                tortoise_client_destroy(client_);
            }
            client_ = other.client_;
            owned_ = other.owned_;
            other.client_ = nullptr;
            other.owned_ = false;
        }
        return *this;
    }
    
    void connect() {
        checkError(tortoise_client_connect(client_));
    }
    
    void disconnect() {
        checkError(tortoise_client_disconnect(client_));
    }
    
    bool isConnected() const {
        bool connected = false;
        checkError(tortoise_client_is_connected(client_, &connected));
        return connected;
    }
    
    Message sendMessage(const Message& msg, uint32_t timeout_ms = 5000) {
        tortoise_message_t response;
        checkError(tortoise_client_send_message(
            client_, msg.native(), &response, timeout_ms));
        return Message(response.content, 
                     response.sender_id ?: "",
                     response.channel ?: "");
    }
    
    void sendMessageAsync(
        const Message& msg,
        std::function<void(tortoise_error_t, const Message&)> callback) {
        
        checkError(tortoise_client_send_message_async(
            client_,
            msg.native(),
            [](tortoise_error_t error, 
               const tortoise_message_t* resp, 
               void* user_data) {
                auto* cb = static_cast<
                    std::function<void(tortoise_error_t, const Message&)>*>(
                    user_data);
                Message msg;
                if (resp) {
                    msg = Message(resp->content,
                                 resp->sender_id ?: "",
                                 resp->channel ?: "");
                }
                (*cb)(error, msg);
            },
            &callback));
    }
    
    void setConnectionCallback(
        std::function<void(bool)> callback) {
        
        checkError(tortoise_client_set_connection_callback(
            client_,
            [](bool connected, void* user_data) {
                auto* cb = static_cast<
                    std::function<void(bool)>*>(user_data);
                (*cb)(connected);
            },
            &callback));
    }
    
    std::string getErrorMessage() const {
        return tortoise_client_get_error_message(client_);
    }
    
    tortoise_client_t* native() { return client_; }
    
private:
    void checkError(tortoise_error_t error) {
        if (error != TORTOISE_OK) {
            throw TortoiseException(error, getErrorMessage());
        }
    }
    
    tortoise_client_t* client_;
    bool owned_;
};

/**
 * 获取 SDK 版本
 */
inline void getVersion(int* major, int* minor, int* patch) {
    tortoise_get_version(major, minor, patch);
}

} // namespace tortoise

#endif // TORTOISE_EMBEDDED_HPP
