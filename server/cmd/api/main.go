package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"tortoise-server/internal/api"
	"tortoise-server/internal/store"
)

func main() {
	// 加载环境变量
	loadEnv()

	// 初始化存储
	memStore := store.NewMemoryStore()
	sessionStore := store.NewSessionStore()
	messageStore := store.NewMessageStore()
	pluginStore := store.NewPluginStore()
	configStore := store.NewConfigStore()

	// 尝试从环境变量加载配置
	loadConfigFromEnv(configStore)

	// 创建 API 服务器
	server := api.NewServer(memStore, sessionStore, messageStore, pluginStore, configStore)

	// 启动服务器
	go func() {
		addr := ":" + getEnv("PORT", "18792")
		log.Printf("🚀 Tortoise API Server 启动中...")
		log.Printf("📍 HTTP API: http://localhost%s/api/v1", addr)
		log.Printf("📍 WebSocket: ws://localhost%s/ws", addr)
		
		if err := server.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 正在关闭服务器...")
}

// loadEnv 加载 .env 文件（如果存在）
func loadEnv() {
	// 尝试加载 .env 文件
	envFile := ".env"
	if _, err := os.Stat(envFile); err == nil {
		log.Printf("📄 从 %s 加载环境变量", envFile)
		// 简单实现：在实际项目中可以使用 godotenv 库
	}
}

// loadConfigFromEnv 从环境变量加载配置
func loadConfigFromEnv(configStore *store.ConfigStore) {
	// AI 配置
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		log.Printf("✅ 检测到 OpenAI API Key")
	}

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		log.Printf("✅ 检测到 Anthropic API Key")
	}

	// 渠道配置
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		log.Printf("✅ 检测到 Telegram Bot Token")
	}

	if token := os.Getenv("DISCORD_BOT_TOKEN"); token != "" {
		log.Printf("✅ 检测到 Discord Bot Token")
	}

	// 数据库配置
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		log.Printf("✅ 检测到 Redis URL")
	}

	// 安全配置
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		log.Printf("✅ 检测到 JWT Secret")
	}
}

// getEnv 获取环境变量，带默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
