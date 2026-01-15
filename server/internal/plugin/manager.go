package plugin

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ToolParameter - 工具参数
type ToolParameter struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Default     string
}

// ToolDefinition - 工具定义
type ToolDefinition struct {
	Name                string
	Description         string
	Parameters          []ToolParameter
	RequireConfirmation bool
	Category            string
}

// PluginInfo - 插件信息
type PluginInfo struct {
	ID          string
	Name        string
	Version     string
	Description string
	Author      string
	License     string
	Homepage    string
	Repository  string
}

// Plugin - 插件
type Plugin struct {
	Info        PluginInfo
	Tools       []ToolDefinition
	Config      map[string]string
	State       PluginState
	InstalledAt time.Time
}

// PluginState - 插件状态
type PluginState int

const (
	PluginStateInstalled PluginState = 1
	PluginStateLoaded   PluginState = 2
	PluginStateRunning  PluginState = 3
	PluginStateDisabled PluginState = 4
	PluginStateError    PluginState = 5
)

// Manager - 插件管理器
type Manager struct {
	plugins map[string]*Plugin
	tools   map[string]string // tool name -> plugin id
	mu      sync.RWMutex
}

// NewManager 创建插件管理器
func NewManager() *Manager {
	m := &Manager{
		plugins: make(map[string]*Plugin),
		tools:   make(map[string]string),
	}
	// 注册内置工具
	m.registerBuiltinTools()
	return m
}

// Install 安装插件
func (m *Manager) Install(source string, config map[string]string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 模拟安装
	id := uuid.New().String()
	plugin := &Plugin{
		Info: PluginInfo{
			ID:          id,
			Name:        "Plugin-" + id[:8],
			Version:     "1.0.0",
			Description: "A plugin",
			Author:      "Tortoise Team",
			License:     "MIT",
		},
		Tools:       []ToolDefinition{},
		Config:      config,
		State:       PluginStateInstalled,
		InstalledAt: time.Now(),
	}

	m.plugins[id] = plugin
	return id, nil
}

// Uninstall 卸载插件
func (m *Manager) Uninstall(id string, force bool) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, ok := m.plugins[id]
	if !ok {
		return false
	}

	// 移除工具注册
	for _, tool := range plugin.Tools {
		delete(m.tools, tool.Name)
	}

	delete(m.plugins, id)
	return true
}

// Get 获取插件
func (m *Manager) Get(id string) *Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.plugins[id]
}

// List 列出所有插件
func (m *Manager) List() []*Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]*Plugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// ListTools 列出所有工具
func (m *Manager) ListTools() []ToolDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools := make([]ToolDefinition, 0)
	for _, plugin := range m.plugins {
		tools = append(tools, plugin.Tools...)
	}
	return tools
}

// Execute 执行工具
func (m *Manager) Execute(pluginID, toolName, args string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, ok := m.plugins[pluginID]
	if !ok {
		return "", ErrPluginNotFound
	}

	// 查找工具
	var tool *ToolDefinition
	for i := range plugin.Tools {
		if plugin.Tools[i].Name == toolName {
			tool = &plugin.Tools[i]
			break
		}
	}

	if tool == nil {
		return "", ErrToolNotFound
	}

	// 模拟执行
	var argsMap map[string]interface{}
	json.Unmarshal([]byte(args), &argsMap)

	result := map[string]interface{}{
		"success": true,
		"tool":    toolName,
		"result":  "Tool executed successfully",
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// registerBuiltinTools 注册内置工具
func (m *Manager) registerBuiltinTools() {
	builtinTools := []ToolDefinition{
		{
			Name:        "web_search",
			Description: "Search the web for information",
			Category:    "search",
			Parameters: []ToolParameter{
				{Name: "query", Type: "string", Description: "Search query", Required: true},
				{Name: "limit", Type: "number", Description: "Maximum results", Required: false, Default: "10"},
			},
		},
		{
			Name:        "calculator",
			Description: "Perform mathematical calculations",
			Category:    "utility",
			Parameters: []ToolParameter{
				{Name: "expression", Type: "string", Description: "Math expression", Required: true},
			},
		},
		{
			Name:        "file_read",
			Description: "Read a file from the filesystem",
			Category:    "filesystem",
			Parameters: []ToolParameter{
				{Name: "path", Type: "string", Description: "File path", Required: true},
			},
		},
		{
			Name:        "file_write",
			Description: "Write content to a file",
			Category:    "filesystem",
			Parameters: []ToolParameter{
				{Name: "path", Type: "string", Description: "File path", Required: true},
				{Name: "content", Type: "string", Description: "Content to write", Required: true},
			},
		},
	}

	builtinPlugin := &Plugin{
		Info: PluginInfo{
			ID:          "builtin",
			Name:        "builtin",
			Version:     "1.0.0",
			Description: "Built-in tools",
			Author:      "Tortoise Team",
			License:     "MIT",
		},
		Tools:       builtinTools,
		State:       PluginStateRunning,
		InstalledAt: time.Now(),
	}

	m.plugins["builtin"] = builtinPlugin
	for _, tool := range builtinTools {
		m.tools[tool.Name] = "builtin"
	}
}
