package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"tortoise/config"
	"tortoise/database"
	"tortoise/handlers"
	"tortoise/middleware"
	"tortoise/services/ai"
	"tortoise/services/channel"
	"tortoise/services/plugin"
	"tortoise/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 初始化服务
	aiService := ai.NewAIService(cfg.AI)
	channelService := channel.NewChannelService(cfg.Channels)
	pluginService := plugin.NewPluginService(cfg.Plugins)
	wsHub := websocket.NewHub()

	// 初始化处理器
	h := handlers.NewHandler(handlers.Dependencies{
		DB:             db,
		AIService:      aiService,
		ChannelService: channelService,
		PluginService:  pluginService,
		WSHub:          wsHub,
	})

	// 初始化新处理器
	marketplaceHandler := handlers.NewMarketplaceHandler(nil)      // TODO: 注入 marketplace
	orchestratorHandler := handlers.NewOrchestratorHandler()
	enterpriseHandler := handlers.NewEnterpriseHandler()
	clusterHandler := handlers.NewClusterHandler()
	memoryHandler := handlers.NewMemoryHandler()

	// 设置 Gin 模式
	if cfg.App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 引擎
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	// 健康检查
	r.GET("/health", h.HealthCheck)

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 认证中间件
		api.Use(middleware.Auth(cfg.App.SecretKey))

		// 会话管理
		sessions := api.Group("/sessions")
		{
			sessions.GET("", h.ListSessions)
			sessions.POST("", h.CreateSession)
			sessions.GET("/:id", h.GetSession)
			sessions.DELETE("/:id", h.DeleteSession)
			sessions.GET("/:id/messages", h.ListMessages)
		}

		// AI 聊天
		chat := api.Group("/chat")
		{
			chat.POST("/completions", h.ChatCompletions)
			chat.POST("/completions/stream", h.ChatCompletionsStream)
		}

		// 渠道管理
		channels := api.Group("/channels")
		{
			channels.GET("", h.ListChannels)
			channels.POST("", h.CreateChannel)
			channels.GET("/:id", h.GetChannel)
			channels.PUT("/:id", h.UpdateChannel)
			channels.DELETE("/:id", h.DeleteChannel)
			channels.POST("/:id/connect", h.ConnectChannel)
			channels.POST("/:id/disconnect", h.DisconnectChannel)
		}

		// 记忆管理
		memory := api.Group("/memory")
		{
			memory.GET("", h.ListMemories)
			memory.POST("", h.CreateMemory)
			memory.GET("/:id", h.GetMemory)
			memory.PUT("/:id", h.UpdateMemory)
			memory.DELETE("/:id", h.DeleteMemory)
			memory.GET("/search", h.SearchMemory)
		}

		// 插件管理
		plugins := api.Group("/plugins")
		{
			plugins.GET("", h.ListPlugins)
			plugins.POST("/install", h.InstallPlugin)
			plugins.POST("/:id/enable", h.EnablePlugin)
			plugins.POST("/:id/disable", h.DisablePlugin)
			plugins.DELETE("/:id", h.UninstallPlugin)
		}

		// 配置管理
		config := api.Group("/config")
		{
			config.GET("", h.GetConfig)
			config.PUT("", h.UpdateConfig)
		}
	}

	// WebSocket
	r.GET("/ws", h.WebSocket)

	// 启动 WebSocket Hub
	go wsHub.Run()

	// 启动渠道服务
	go channelService.Start()

	// 启动插件服务
	go pluginService.LoadAll()

	// 启动服务器
	go func() {
		addr := cfg.App.Address
		log.Printf("Server starting on %s", addr)
		if err := r.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	channelService.Stop()
	pluginService.UnloadAll()
	log.Println("Server stopped")
}
