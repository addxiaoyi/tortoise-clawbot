package plugin

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// Plugin 插件接口
type Plugin interface {
	Info() PluginInfo
	Initialize(config json.RawMessage) error
	Execute(tool string, args map[string]interface{}) (interface{}, error)
	Shutdown() error
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID          string
	Name        string
	Version     string
	Description string
	Author      string
	Tools       []Tool
	Resources   []Resource
	Prompts     []Prompt
}

// Tool 工具定义
type Tool struct {
	Name        string
	Description string
	Parameters  []Parameter
}

// Parameter 参数定义
type Parameter struct {
	Name     string
	Type     string
	Required bool
	Default  interface{}
}

// Resource 资源定义
type Resource struct {
	URI         string
	Name        string
	Description string
	MimeType   string
}

// Prompt 提示定义
type Prompt struct {
	Name        string
	Description string
	Arguments   []Parameter
}

// Config 插件主机配置
type Config struct {
	SandboxEnabled bool
	MaxPlugins    int
	Timeout      time.Duration
}

// Host 插件主机
type Host struct {
	config Config

	// 插件注册
	plugins map[string]Plugin

	// 插件信息缓存
	infos map[string]*PluginInfo

	// 统计
	stats Stats

	// 控制
	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// Stats 插件统计
type Stats struct {
	PluginsLoaded    atomic.Int64
	PluginsActive   atomic.Int64
	ToolsRegistered atomic.Int64
	Executions     atomic.Int64
	ExecutionsMs   atomic.Int64
	Errors         atomic.Int64
}

// NewHost 创建插件主机
func NewHost(cfg Config) *Host {
	if cfg.MaxPlugins == 0 {
		cfg.MaxPlugins = 100
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Host{
		config:  cfg,
		plugins: make(map[string]Plugin),
		infos:   make(map[string]*PluginInfo),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Register 注册插件
func (h *Host) Register(plugin Plugin) error {
	info := plugin.Info()

	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.plugins) >= h.config.MaxPlugins {
		return ErrPluginLimitExceeded
	}

	if _, exists := h.plugins[info.ID]; exists {
		return ErrPluginAlreadyExists
	}

	if err := plugin.Initialize(nil); err != nil {
		return err
	}

	h.plugins[info.ID] = plugin
	h.infos[info.ID] = &info
	h.stats.PluginsLoaded.Add(1)
	h.stats.PluginsActive.Add(1)
	h.stats.ToolsRegistered.Add(int64(len(info.Tools)))

	return nil
}

// Unregister 注销插件
func (h *Host) Unregister(pluginID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	plugin, ok := h.plugins[pluginID]
	if !ok {
		return ErrPluginNotFound
	}

	if err := plugin.Shutdown(); err != nil {
		return err
	}

	delete(h.plugins, pluginID)
	delete(h.infos, pluginID)
	h.stats.PluginsActive.Add(-1)

	return nil
}

// Execute 执行插件工具
func (h *Host) Execute(pluginID, tool string, args map[string]interface{}) (interface{}, error) {
	h.mu.RLock()
	plugin, ok := h.plugins[pluginID]
	h.mu.RUnlock()

	if !ok {
		return nil, ErrPluginNotFound
	}

	start := time.Now()
	result, err := plugin.Execute(tool, args)
	h.stats.Executions.Add(1)
	h.stats.ExecutionsMs.Store(time.Since(start).Milliseconds())

	if err != nil {
		h.stats.Errors.Add(1)
		return nil, err
	}

	return result, nil
}

// GetPlugins 获取所有插件
func (h *Host) GetPlugins() []*PluginInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	infos := make([]*PluginInfo, 0, len(h.infos))
	for _, info := range h.infos {
		infos = append(infos, info)
	}
	return infos
}

// GetPlugin 获取单个插件
func (h *Host) GetPlugin(pluginID string) (*PluginInfo, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	info, ok := h.infos[pluginID]
	return info, ok
}

// ListTools 列出所有工具
func (h *Host) ListTools() []Tool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	tools := make([]Tool, 0)
	for _, info := range h.infos {
		tools = append(tools, info.Tools...)
	}
	return tools
}

// Stats 获取统计
func (h *Host) Stats() Stats {
	return Stats{
		PluginsLoaded:    h.stats.PluginsLoaded,
		PluginsActive:   h.stats.PluginsActive,
		ToolsRegistered: h.stats.ToolsRegistered,
		Executions:     h.stats.Executions,
		ExecutionsMs:   h.stats.ExecutionsMs,
		Errors:         h.stats.Errors,
	}
}

// Stop 停止主机
func (h *Host) Stop() {
	h.cancel()

	h.mu.Lock()
	defer h.mu.Unlock()

	for _, plugin := range h.plugins {
		plugin.Shutdown()
	}
}

// Errors
var (
	ErrPluginNotFound       = &PluginError{Code: "PLUGIN_NOT_FOUND", Message: "插件未找到"}
	ErrPluginAlreadyExists  = &PluginError{Code: "PLUGIN_EXISTS", Message: "插件已存在"}
	ErrPluginLimitExceeded = &PluginError{Code: "PLUGIN_LIMIT", Message: "插件数量已达上限"}
)

// PluginError 插件错误
type PluginError struct {
	Code    string
	Message string
}

func (e *PluginError) Error() string {
	return e.Code + ": " + e.Message
}

// ============ 示例插件实现 ============

type ExamplePlugin struct {
	info PluginInfo
}

func NewExamplePlugin() *ExamplePlugin {
	return &ExamplePlugin{
		info: PluginInfo{
			ID:          "example",
			Name:        "Example Plugin",
			Version:     "1.0.0",
			Description: "示例插件",
			Author:      "Tortoise",
			Tools: []Tool{
				{
					Name:        "greet",
					Description: "打招呼",
					Parameters:  []Parameter{},
				},
				{
					Name:        "calculate",
					Description: "计算",
					Parameters: []Parameter{
						{Name: "a", Type: "number", Required: true},
						{Name: "b", Type: "number", Required: true},
					},
				},
			},
		},
	}
}

func (p *ExamplePlugin) Info() PluginInfo {
	return p.info
}

func (p *ExamplePlugin) Initialize(config json.RawMessage) error {
	return nil
}

func (p *ExamplePlugin) Execute(tool string, args map[string]interface{}) (interface{}, error) {
	switch tool {
	case "greet":
		return map[string]string{"message": "Hello!"}, nil
	case "calculate":
		a := args["a"].(float64)
		b := args["b"].(float64)
		return map[string]float64{
			"sum":    a + b,
			"product": a * b,
		}, nil
	default:
		return nil, ErrToolNotFound
	}
}

func (p *ExamplePlugin) Shutdown() error {
	return nil
}

var ErrToolNotFound = &PluginError{Code: "TOOL_NOT_FOUND", Message: "工具未找到"}
