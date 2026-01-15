package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

// ============ Plugin System 插件系统 ============

// Plugin 插件接口
type Plugin interface {
	// 元数据
	Metadata() *PluginMetadata
	
	// 初始化
	Initialize(ctx context.Context, config json.RawMessage) error
	
	// 关闭
	Shutdown() error
	
	// 获取工具
	GetTools() []*Tool
	
	// 执行工具
	Execute(ctx context.Context, tool string, args map[string]interface{}) (interface{}, error)
}

// PluginMetadata 插件元数据
type PluginMetadata struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags,omitempty"`
	Homepage    string   `json:"homepage,omitempty"`
	Repository  string   `json:"repository,omitempty"`
	License     string   `json:"license,omitempty"`
}

// Tool 工具定义
type Tool struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Parameters  []ToolParameter     `json:"parameters,omitempty"`
	Output      *ToolOutput        `json:"output,omitempty"`
}

// ToolParameter 工具参数
type ToolParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // string, number, boolean, object, array
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
}

// ToolOutput 工具输出
type ToolOutput struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ============ Plugin Manager 插件管理器 ============

// Manager 插件管理器
type Manager struct {
	mu      sync.RWMutex
	plugins map[string]*PluginInstance
	httpServer *http.Server
}

// PluginInstance 插件实例
type PluginInstance struct {
	ID       string
	Plugin   Plugin
	Metadata *PluginMetadata
	Enabled  bool
	Config   json.RawMessage
}

// NewManager 创建插件管理器
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]*PluginInstance),
	}
}

// Register 注册插件
func (m *Manager) Register(plugin Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	metadata := plugin.Metadata()
	if metadata == nil {
		return fmt.Errorf("plugin metadata is nil")
	}

	if _, exists := m.plugins[metadata.ID]; exists {
		return fmt.Errorf("plugin %s already registered", metadata.ID)
	}

	instance := &PluginInstance{
		ID:       metadata.ID,
		Plugin:   plugin,
		Metadata: metadata,
		Enabled:  true,
	}

	m.plugins[metadata.ID] = instance
	log.Printf("✅ 插件已注册: %s v%s", metadata.Name, metadata.Version)

	return nil
}

// Unregister 注销插件
func (m *Manager) Unregister(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.plugins[id]
	if !exists {
		return fmt.Errorf("plugin %s not found", id)
	}

	if err := instance.Plugin.Shutdown(); err != nil {
		log.Printf("⚠️ 关闭插件 %s 失败: %v", id, err)
	}

	delete(m.plugins, id)
	log.Printf("🗑️ 插件已注销: %s", id)
	return nil
}

// Get 获取插件
func (m *Manager) Get(id string) *PluginInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.plugins[id]
}

// List 列出所有插件
func (m *Manager) List() []*PluginInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instances := make([]*PluginInstance, 0, len(m.plugins))
	for _, p := range m.plugins {
		instances = append(instances, p)
	}
	return instances
}

// Enable 启用插件
func (m *Manager) Enable(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.plugins[id]
	if !exists {
		return fmt.Errorf("plugin %s not found", id)
	}

	instance.Enabled = true
	log.Printf("✅ 插件已启用: %s", id)
	return nil
}

// Disable 禁用插件
func (m *Manager) Disable(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.plugins[id]
	if !exists {
		return fmt.Errorf("plugin %s not found", id)
	}

	instance.Enabled = false
	log.Printf("⏸️ 插件已禁用: %s", id)
	return nil
}

// Execute 执行工具
func (m *Manager) Execute(ctx context.Context, pluginID, tool string, args map[string]interface{}) (interface{}, error) {
	m.mu.RLock()
	instance, exists := m.plugins[pluginID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	if !instance.Enabled {
		return nil, fmt.Errorf("plugin %s is disabled", pluginID)
	}

	return instance.Plugin.Execute(ctx, tool, args)
}

// GetTools 获取所有工具
func (m *Manager) GetTools() []*Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var tools []*Tool
	for _, instance := range m.plugins {
		if instance.Enabled {
			tools = append(tools, instance.Plugin.GetTools()...)
		}
	}
	return tools
}

// ============ 内置插件 ============

// SearchPlugin 搜索插件
type SearchPlugin struct {
	metadata *PluginMetadata
}

func NewSearchPlugin() *SearchPlugin {
	return &SearchPlugin{
		metadata: &PluginMetadata{
			ID:          "search",
			Name:        "Web Search",
			Version:     "1.0.0",
			Description: "网页搜索工具，支持 Google、Bing 等搜索引擎",
			Author:      "Tortoise",
			Tags:        []string{"search", "web", "utility"},
			License:     "MIT",
		},
	}
}

func (p *SearchPlugin) Metadata() *PluginMetadata {
	return p.metadata
}

func (p *SearchPlugin) Initialize(ctx context.Context, config json.RawMessage) error {
	log.Printf("✅ 搜索插件初始化完成")
	return nil
}

func (p *SearchPlugin) Shutdown() error {
	return nil
}

func (p *SearchPlugin) GetTools() []*Tool {
	return []*Tool{
		{
			Name:        "search",
			Description: "搜索网页内容",
			Parameters: []ToolParameter{
				{
					Name:        "query",
					Type:        "string",
					Description: "搜索关键词",
					Required:    true,
				},
				{
					Name:        "limit",
					Type:        "number",
					Description: "返回结果数量限制",
					Default:     10,
				},
				{
					Name:        "engine",
					Type:        "string",
					Description: "搜索引擎",
					Default:     "google",
					Enum:        []string{"google", "bing", "duckduckgo"},
				},
			},
			Output: &ToolOutput{
				Type:        "array",
				Description: "搜索结果列表",
			},
		},
	}
}

func (p *SearchPlugin) Execute(ctx context.Context, tool string, args map[string]interface{}) (interface{}, error) {
	switch tool {
	case "search":
		return p.search(args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", tool)
	}
}

func (p *SearchPlugin) search(args map[string]interface{}) (interface{}, error) {
	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query is required")
	}

	// 模拟搜索结果
	return []map[string]interface{}{
		{
			"title":       fmt.Sprintf("关于 %s 的搜索结果 1", query),
			"url":         "https://example.com/result1",
			"description": "这是第一个搜索结果的描述...",
			"score":      0.95,
		},
		{
			"title":       fmt.Sprintf("关于 %s 的搜索结果 2", query),
			"url":         "https://example.com/result2",
			"description": "这是第二个搜索结果的描述...",
			"score":      0.85,
		},
	}, nil
}

// CalculatorPlugin 计算器插件
type CalculatorPlugin struct {
	metadata *PluginMetadata
}

func NewCalculatorPlugin() *CalculatorPlugin {
	return &CalculatorPlugin{
		metadata: &PluginMetadata{
			ID:          "calculator",
			Name:        "Calculator",
			Version:     "1.0.0",
			Description: "数学计算工具，支持基础运算和表达式求值",
			Author:      "Tortoise",
			Tags:        []string{"math", "calculator", "utility"},
			License:     "MIT",
		},
	}
}

func (p *CalculatorPlugin) Metadata() *PluginMetadata {
	return p.metadata
}

func (p *CalculatorPlugin) Initialize(ctx context.Context, config json.RawMessage) error {
	return nil
}

func (p *CalculatorPlugin) Shutdown() error {
	return nil
}

func (p *CalculatorPlugin) GetTools() []*Tool {
	return []*Tool{
		{
			Name:        "calculate",
			Description: "执行数学计算",
			Parameters: []ToolParameter{
				{
					Name:        "expression",
					Type:        "string",
					Description: "数学表达式，如 2+2*3",
					Required:    true,
				},
			},
			Output: &ToolOutput{
				Type:        "number",
				Description: "计算结果",
			},
		},
		{
			Name:        "convert",
			Description: "单位转换",
			Parameters: []ToolParameter{
				{
					Name:        "value",
					Type:        "number",
					Description: "要转换的数值",
					Required:    true,
				},
				{
					Name:        "from",
					Type:        "string",
					Description: "源单位",
					Required:    true,
				},
				{
					Name:        "to",
					Type:        "string",
					Description: "目标单位",
					Required:    true,
				},
			},
			Output: &ToolOutput{
				Type:        "number",
				Description: "转换结果",
			},
		},
	}
}

func (p *CalculatorPlugin) Execute(ctx context.Context, tool string, args map[string]interface{}) (interface{}, error) {
	switch tool {
	case "calculate":
		return p.calculate(args)
	case "convert":
		return p.convert(args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", tool)
	}
}

func (p *CalculatorPlugin) calculate(args map[string]interface{}) (interface{}, error) {
	expr, ok := args["expression"].(string)
	if !ok {
		return nil, fmt.Errorf("expression is required")
	}

	// 简化的计算实现
	// 实际应使用 expression-parser 库
	result := 0.0
	fmt.Sscanf(expr, "%f", &result)
	return result, nil
}

func (p *CalculatorPlugin) convert(args map[string]interface{}) (interface{}, error) {
	value, ok := args["value"].(float64)
	if !ok {
		return nil, fmt.Errorf("value is required")
	}

	from, _ := args["from"].(string)
	to, _ := args["to"].(string)

	// 简化转换
	if from == "km" && to == "mi" {
		return value * 0.621371, nil
	}
	if from == "mi" && to == "km" {
		return value * 1.60934, nil
	}

	return value, nil
}

// ImagePlugin 图片处理插件
type ImagePlugin struct {
	metadata *PluginMetadata
}

func NewImagePlugin() *ImagePlugin {
	return &ImagePlugin{
		metadata: &PluginMetadata{
			ID:          "image",
			Name:        "Image Processing",
			Version:     "1.0.0",
			Description: "图片处理工具，支持缩放、裁剪、格式转换等",
			Author:      "Tortoise",
			Tags:        []string{"image", "media", "utility"},
			License:     "MIT",
		},
	}
}

func (p *ImagePlugin) Metadata() *PluginMetadata {
	return p.metadata
}

func (p *ImagePlugin) Initialize(ctx context.Context, config json.RawMessage) error {
	return nil
}

func (p *ImagePlugin) Shutdown() error {
	return nil
}

func (p *ImagePlugin) GetTools() []*Tool {
	return []*Tool{
		{
			Name:        "resize",
			Description: "调整图片大小",
			Parameters: []ToolParameter{
				{
					Name:        "url",
					Type:        "string",
					Description: "图片 URL",
					Required:    true,
				},
				{
					Name:        "width",
					Type:        "number",
					Description: "目标宽度",
					Required:    true,
				},
				{
					Name:        "height",
					Type:        "number",
					Description: "目标高度",
					Required:    true,
				},
			},
		},
		{
			Name:        "thumbnail",
			Description: "生成缩略图",
			Parameters: []ToolParameter{
				{
					Name:        "url",
					Type:        "string",
					Description: "图片 URL",
					Required:    true,
				},
				{
					Name:        "size",
					Type:        "number",
					Description: "缩略图大小",
					Default:     200,
				},
			},
		},
	}
}

func (p *ImagePlugin) Execute(ctx context.Context, tool string, args map[string]interface{}) (interface{}, error) {
	switch tool {
	case "resize":
		return map[string]interface{}{
			"url":    args["url"],
			"width":  args["width"],
			"height": args["height"],
			"status": "resized",
		}, nil
	case "thumbnail":
		return map[string]interface{}{
			"url":  args["url"],
			"size": args["size"],
			"status": "thumbnail_created",
		}, nil
	default:
		return nil, fmt.Errorf("unknown tool: %s", tool)
	}
}

// ============ 内置插件注册 ============

// RegisterBuiltins 注册内置插件
func (m *Manager) RegisterBuiltins() {
	m.Register(NewSearchPlugin())
	m.Register(NewCalculatorPlugin())
	m.Register(NewImagePlugin())
	log.Printf("📦 已注册 %d 个内置插件", len(m.plugins))
}
