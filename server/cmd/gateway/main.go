package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/tortoise/server/internal/gateway"
	"github.com/tortoise/server/internal/session"
	"github.com/tortoise/server/internal/channel"
	"github.com/tortoise/server/internal/plugin"
	"github.com/tortoise/server/internal/memory"
	pb "github.com/tortoise/protocol/go"
)

// TortoiseGateway - 主 Gateway 服务实现
type TortoiseGateway struct {
	pb.UnimplementedTortoiseServiceServer

	// 核心组件
	sessions  *session.Manager
	channels  *channel.Manager
	plugins   *plugin.Manager
	memory    *memory.Manager

	// Gateway 配置
	config    *gateway.Config
	server    *grpc.Server
	listener  net.Listener

	// 状态
	mu        sync.RWMutex
	running   bool
	startTime time.Time

	// 连接管理
	conns     map[string]*ClientConnection
}

type ClientConnection struct {
	ID        string
	UserID    string
	Conn      *grpc.ClientConn
	Stream    pb.TortoiseService_SubscribeClient
	CreatedAt time.Time
}

// NewTortoiseGateway creates a new gateway instance
func NewTortoiseGateway(cfg *gateway.Config) *TortoiseGateway {
	return &TortoiseGateway{
		config:   cfg,
		sessions: session.NewManager(cfg.MaxSessions),
		channels: channel.NewManager(),
		plugins:  plugin.NewManager(),
		memory:   memory.NewManager(),
		conns:    make(map[string]*ClientConnection),
	}
}

// Start 启动 Gateway 服务
func (g *TortoiseGateway) Start() error {
	g.mu.Lock()
	if g.running {
		g.mu.Unlock()
		return fmt.Errorf("gateway already running")
	}

	// 创建监听器
	addr := fmt.Sprintf("%s:%d", g.config.BindAddress, g.config.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		g.mu.Unlock()
		return fmt.Errorf("failed to listen: %w", err)
	}
	g.listener = lis

	// 创建 gRPC 服务器
	var opts []grpc.ServerOption
	if !g.config.TLSEnabled {
		opts = append(opts, grpc.Creds(insecure.NewCredentials()))
	}

	g.server = grpc.NewServer(opts...)
	g.running = true
	g.startTime = time.Now()
	g.mu.Unlock()

	// 注册服务
	pb.RegisterTortoiseServiceServer(g.server, g)

	// 启动渠道连接管理器
	go g.channels.Start()

	log.Printf("Gateway starting on %s", addr)
	return g.server.Serve(lis)
}

// Stop 停止 Gateway 服务
func (g *TortoiseGateway) Stop() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.running {
		return nil
	}

	g.running = false
	g.server.GracefulStop()

	// 关闭所有渠道连接
	g.channels.Stop()

	// 关闭所有客户端连接
	for _, conn := range g.conns {
		conn.Conn.Close()
	}

	log.Println("Gateway stopped")
	return nil
}

// HealthCheck 健康检查
func (g *TortoiseGateway) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	g.mu.RLock()
	running := g.running
	startTime := g.startTime
	g.mu.RUnlock()

	uptime := time.Since(startTime).Seconds()
	components := []*pb.HealthCheckComponent{
		{Name: "gateway", Healthy: running, Message: "Gateway is running"},
		{Name: "sessions", Healthy: true, Message: fmt.Sprintf("%d active sessions", g.sessions.Count())},
		{Name: "channels", Healthy: true, Message: "Channels operational"},
		{Name: "plugins", Healthy: true, Message: "Plugins loaded"},
		{Name: "memory", Healthy: true, Message: "Memory system operational"},
	}

	healthy := running
	for _, c := range components {
		if !c.Healthy {
			healthy = false
			break
		}
	}

	return &pb.HealthCheckResponse{
		Healthy:        healthy,
		Version:        "0.1.0",
		UptimeSeconds:  int64(uptime),
		Components:     components,
	}, nil
}

// CreateSession 创建新会话
func (g *TortoiseGateway) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	sess := g.sessions.Create(req.UserId, req.Metadata)
	return &pb.CreateSessionResponse{
		Session: &pb.Session{
			Id:        sess.ID,
			UserId:    sess.UserID,
			State:     pb.SessionState_SESSION_STATE_ACTIVE,
			CreatedAt: sess.CreatedAt.Unix(),
			UpdatedAt: sess.UpdatedAt.Unix(),
			Metadata:  sess.Metadata,
		},
	}, nil
}

// GetSession 获取会话
func (g *TortoiseGateway) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.Session, error) {
	sess, err := g.sessions.Get(req.SessionId)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	return &pb.Session{
		Id:        sess.ID,
		UserId:    sess.UserID,
		State:     pb.SessionState_SESSION_STATE_ACTIVE,
		CreatedAt: sess.CreatedAt.Unix(),
		UpdatedAt: sess.UpdatedAt.Unix(),
		Metadata:  sess.Metadata,
	}, nil
}

// DeleteSession 删除会话
func (g *TortoiseGateway) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.Status, error) {
	ok := g.sessions.Delete(req.SessionId)
	return &pb.Status{
		Success: ok,
		Message: "Session deleted",
	}, nil
}

// ListSessions 列出所有会话
func (g *TortoiseGateway) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	sessions := g.sessions.List(req.UserId)
	pbSessions := make([]*pb.Session, len(sessions))
	for i, s := range sessions {
		pbSessions[i] = &pb.Session{
			Id:        s.ID,
			UserId:    s.UserID,
			State:     pb.SessionState_SESSION_STATE_ACTIVE,
			CreatedAt: s.CreatedAt.Unix(),
			UpdatedAt: s.UpdatedAt.Unix(),
			Metadata:  s.Metadata,
		}
	}
	return &pb.ListSessionsResponse{
		Sessions: pbSessions,
		Total:    int32(len(sessions)),
	}, nil
}

// SendMessage 发送消息
func (g *TortoiseGateway) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// 获取会话
	sess, err := g.sessions.Get(req.SessionId)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// 创建消息
	msg := &session.Message{
		ID:        uuid.New().String(),
		SessionID: sess.ID,
		Role:      "user",
		Content:   req.Content,
		Format:    req.Format.String(),
		Type:      "text",
		Timestamp: time.Now(),
		Metadata:  req.Metadata,
	}

	// 添加到会话
	sess.AddMessage(msg)

	return &pb.SendMessageResponse{
		MessageId: msg.ID,
		SessionId: sess.ID,
		Streaming: req.Stream,
	}, nil
}

// SendMessageStream 流式发送消息
func (g *TortoiseGateway) SendMessageStream(req *pb.SendMessageRequest, stream pb.TortoiseService_SendMessageStreamServer) error {
	// 获取会话
	sess, err := g.sessions.Get(req.SessionId)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// 创建消息
	msg := &session.Message{
		ID:        uuid.New().String(),
		SessionID: sess.ID,
		Role:      "user",
		Content:   req.Content,
		Format:    req.Format.String(),
		Type:      "text",
		Timestamp: time.Now(),
		Metadata:  req.Metadata,
	}

	// 模拟流式响应
	words := []string{"Thinking", "processing", "your", "request", "..."}
	for i, word := range words {
		chunk := &pb.MessageChunk{
			MessageId: msg.ID,
			Content:   word,
			IsFinal:   i == len(words)-1,
			Delta:     word,
		}
		if err := stream.Send(chunk); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// GetMessages 获取消息历史
func (g *TortoiseGateway) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	sess, err := g.sessions.Get(req.SessionId)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	messages := sess.GetMessages(req.Limit, req.Offset)
	pbMessages := make([]*pb.Message, len(messages))
	for i, m := range messages {
		pbMessages[i] = &pb.Message{
			Id:        m.ID,
			SessionId: m.SessionID,
			Role:      m.Role,
			Content:   m.Content,
			Format:    pb.ContentFormat_FORMAT_PLAIN,
			Type:      pb.MessageType_MESSAGE_TYPE_TEXT,
			Timestamp: m.Timestamp.Unix(),
			Metadata:  m.Metadata,
		}
	}

	return &pb.GetMessagesResponse{
		Messages: pbMessages,
		Total:    int32(len(pbMessages)),
	}, nil
}

// ListTools 列出所有工具
func (g *TortoiseGateway) ListTools(req *emptypb.Empty, stream pb.TortoiseService_ListToolsServer) error {
	tools := g.plugins.ListTools()
	for _, tool := range tools {
		pbTool := &pb.ToolDefinition{
			Name:                tool.Name,
			Description:         tool.Description,
			RequireConfirmation: tool.RequireConfirmation,
			Category:            tool.Category,
		}
		for _, p := range tool.Parameters {
			pbTool.Parameters = append(pbTool.Parameters, &pb.ToolParameter{
				Name:        p.Name,
				Type:        p.Type,
				Description: p.Description,
				Required:    p.Required,
				Default:     p.Default,
			})
		}
		if err := stream.Send(pbTool); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteTool 执行工具
func (g *TortoiseGateway) ExecuteTool(ctx context.Context, req *pb.ExecuteToolRequest) (*pb.ExecuteToolResponse, error) {
	result, err := g.plugins.Execute(req.PluginId, req.ToolName, req.Arguments)
	if err != nil {
		return &pb.ExecuteToolResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.ExecuteToolResponse{
		Success:     true,
		Result:      result,
		DurationMs:  100,
	}, nil
}

// QueryMemory 查询记忆
func (g *TortoiseGateway) QueryMemory(ctx context.Context, req *pb.MemoryQuery) (*pb.MemoryQueryResult, error) {
	result := g.memory.Query(req.Query, req.MemoryType.String(), int(req.Limit))
	
	pbMemories := make([]*pb.Memory, len(result))
	for i, m := range result {
		pbMemories[i] = &pb.Memory{
			Id:         m.ID,
			MemoryType: pb.MemoryType_MEMORY_TYPE_SEMANTIC,
			Content:    m.Content,
			Importance: m.Importance,
			CreatedAt:  m.CreatedAt.Unix(),
			AccessedAt: m.AccessedAt.Unix(),
			Metadata:   m.Metadata,
		}
	}

	scores := make([]float32, len(result))
	return &pb.MemoryQueryResult{
		Memories: pbMemories,
		Scores:   scores,
	}, nil
}

// SaveMemory 保存记忆
func (g *TortoiseGateway) SaveMemory(ctx context.Context, req *pb.SaveMemoryRequest) (*pb.SaveMemoryResponse, error) {
	id := g.memory.Save(req.Type.String(), req.Content, req.Importance, req.Metadata)
	return &pb.SaveMemoryResponse{
		MemoryId: id,
	}, nil
}

// DeleteMemory 删除记忆
func (g *TortoiseGateway) DeleteMemory(ctx context.Context, req *pb.DeleteMemoryRequest) (*pb.Status, error) {
	ok := g.memory.Delete(req.MemoryId)
	return &pb.Status{
		Success: ok,
		Message: "Memory deleted",
	}, nil
}

// ListPlugins 列出所有插件
func (g *TortoiseGateway) ListPlugins(req *emptypb.Empty, stream pb.TortoiseService_ListPluginsServer) error {
	plugins := g.plugins.List()
	for _, p := range plugins {
		pbPlugin := &pb.Plugin{
			Id:          p.Info.ID,
			Name:        p.Info.Name,
			Version:     p.Info.Version,
			Description: p.Info.Description,
			State:       pb.PluginState_PLUGIN_STATE_LOADED,
			InstalledAt: p.InstalledAt.Unix(),
		}
		for _, t := range p.Tools {
			pbPlugin.Tools = append(pbPlugin.Tools, &pb.ToolDefinition{
				Name:        t.Name,
				Description: t.Description,
				Category:    t.Category,
			})
		}
		if err := stream.Send(pbPlugin); err != nil {
			return err
		}
	}
	return nil
}

// InstallPlugin 安装插件
func (g *TortoiseGateway) InstallPlugin(ctx context.Context, req *pb.InstallPluginRequest) (*pb.Plugin, error) {
	id, err := g.plugins.Install(req.Source, req.Config)
	if err != nil {
		return nil, err
	}

	p := g.plugins.Get(id)
	return &pb.Plugin{
		Id:          p.Info.ID,
		Name:        p.Info.Name,
		Version:     p.Info.Version,
		Description: p.Info.Description,
		State:       pb.PluginState_PLUGIN_STATE_INSTALLED,
		InstalledAt: p.InstalledAt.Unix(),
	}, nil
}

// UninstallPlugin 卸载插件
func (g *TortoiseGateway) UninstallPlugin(ctx context.Context, req *pb.UninstallPluginRequest) (*pb.Status, error) {
	ok := g.plugins.Uninstall(req.PluginId, req.Force)
	return &pb.Status{
		Success: ok,
		Message: "Plugin uninstalled",
	}, nil
}

// ListChannels 列出所有渠道
func (g *TortoiseGateway) ListChannels(req *emptypb.Empty, stream pb.TortoiseService_ListChannelsServer) error {
	channels := g.channels.List()
	for _, c := range channels {
		pbChannel := &pb.Channel{
			Id:        c.ID,
			Type:      pb.ChannelType(c.Type),
			Name:      c.Name,
			State:     pb.ChannelState(c.State),
			Config:    c.Config,
			CreatedAt: c.CreatedAt.Unix(),
		}
		if err := stream.Send(pbChannel); err != nil {
			return err
		}
	}
	return nil
}

// ConnectChannel 连接渠道
func (g *TortoiseGateway) ConnectChannel(ctx context.Context, req *pb.ConnectChannelRequest) (*pb.Channel, error) {
	c, err := g.channels.Connect(int(req.Type), req.Credentials, req.Config)
	if err != nil {
		return nil, err
	}

	return &pb.Channel{
		Id:        c.ID,
		Type:      pb.ChannelType(c.Type),
		Name:      c.Name,
		State:     pb.ChannelState(c.State),
		Config:    c.Config,
		CreatedAt: c.CreatedAt.Unix(),
	}, nil
}

// DisconnectChannel 断开渠道连接
func (g *TortoiseGateway) DisconnectChannel(ctx context.Context, req *pb.DisconnectChannelRequest) (*pb.Status, error) {
	ok := g.channels.Disconnect(req.ChannelId)
	return &pb.Status{
		Success: ok,
		Message: "Channel disconnected",
	}, nil
}

// GetConfig 获取配置
func (g *TortoiseGateway) GetConfig(ctx context.Context, req *emptypb.Empty) (*pb.Config, error) {
	return &pb.Config{
		Version: "0.1.0",
		Gateway: &pb.GatewayConfig{
			BindAddress:           g.config.BindAddress,
			Port:                  int32(g.config.Port),
			TlsEnabled:            g.config.TLSEnabled,
			MaxConnections:        int32(g.config.MaxConnections),
			ConnectionTimeoutMs:   int32(g.config.ConnectionTimeout / time.Millisecond),
		},
	}, nil
}

// UpdateConfig 更新配置
func (g *TortoiseGateway) UpdateConfig(ctx context.Context, req *pb.UpdateConfigRequest) (*pb.Status, error) {
	if req.Config != nil {
		if req.Config.Gateway != nil {
			g.config.BindAddress = req.Config.Gateway.BindAddress
			g.config.Port = int(req.Config.Gateway.Port)
			g.config.TLSEnabled = req.Config.Gateway.TlsEnabled
		}
	}
	return &pb.Status{
		Success: true,
		Message: "Config updated",
	}, nil
}

// StartGateway 启动 Gateway
func (g *TortoiseGateway) StartGateway(ctx context.Context, req *pb.StartGatewayRequest) (*pb.Status, error) {
	if req.Config != nil {
		g.config.BindAddress = req.Config.Gateway.BindAddress
		g.config.Port = int(req.Config.Gateway.Port)
		g.config.TLSEnabled = req.Config.Gateway.TlsEnabled
	}

	go func() {
		if err := g.Start(); err != nil {
			log.Printf("Gateway error: %v", err)
		}
	}()

	return &pb.Status{
		Success: true,
		Message: "Gateway started",
	}, nil
}

// StopGateway 停止 Gateway
func (g *TortoiseGateway) StopGateway(ctx context.Context, req *pb.StopGatewayRequest) (*pb.Status, error) {
	err := g.Stop()
	return &pb.Status{
		Success: err == nil,
		Message: "Gateway stopped",
	}, nil
}

// GetGatewayStatus 获取 Gateway 状态
func (g *TortoiseGateway) GetGatewayStatus(ctx context.Context, req *emptypb.Empty) (*pb.GatewayStatus, error) {
	g.mu.RLock()
	running := g.running
	port := g.config.Port
	startTime := g.startTime
	g.mu.RUnlock()

	return &pb.GatewayStatus{
		Running:            running,
		Port:               int32(port),
		ActiveConnections:  int32(len(g.conns)),
		UptimeSeconds:      int64(time.Since(startTime).Seconds()),
	}, nil
}

// GetMetrics 获取系统指标
func (g *TortoiseGateway) GetMetrics(ctx context.Context, req *pb.MetricsRequest) (*pb.MetricsResponse, error) {
	return &pb.MetricsResponse{
		Timestamp: time.Now().Unix(),
		System: &pb.SystemMetrics{
			ActiveConnections: int32(len(g.conns)),
			Goroutines:        100,
		},
	}, nil
}

// Subscribe 订阅事件
func (g *TortoiseGateway) Subscribe(req *pb.SubscribeRequest, stream pb.TortoiseService_SubscribeClient) error {
	// 创建客户端连接
	connID := uuid.New().String()
	conn := &ClientConnection{
		ID:        connID,
		Conn:      nil,
		Stream:    stream,
		CreatedAt: time.Now(),
	}

	g.mu.Lock()
	g.conns[connID] = conn
	g.mu.Unlock()

	defer func() {
		g.mu.Lock()
		delete(g.conns, connID)
		g.mu.Unlock()
	}()

	// 保持连接，直到客户端断开
	<-stream.Context().Done()
	return nil
}

func main() {
	cfg := &gateway.Config{
		BindAddress:         "0.0.0.0",
		Port:                18792,
		TLSEnabled:          false,
		MaxConnections:      10000,
		ConnectionTimeout:  30 * time.Second,
		ReadTimeout:         30 * time.Second,
		WriteTimeout:        30 * time.Second,
		MaxSessions:        10000,
	}

	gw := NewTortoiseGateway(cfg)
	
	if err := gw.Start(); err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}
}
