package store

import (
	"sync"

	"github.com/google/uuid"
)

// Tool 工具模型
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  []Parameter            `json:"parameters,omitempty"`
}

// Parameter 参数模型
type Parameter struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// Plugin 插件模型
type Plugin struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Enabled     bool     `json:"enabled"`
	Tools       []Tool   `json:"tools"`
	Status      string   `json:"status"` // active, inactive, error
}

// PluginStore 插件存储
type PluginStore struct {
	plugins map[string]*Plugin
	mu      sync.RWMutex
}

// NewPluginStore 创建插件存储
func NewPluginStore() *PluginStore {
	store := &PluginStore{
		plugins: make(map[string]*Plugin),
	}
	store.createSampleData()
	return store
}

func (p *PluginStore) createSampleData() {
	plugins := []*Plugin{
		{
			ID:          uuid.New().String(),
			Name:        "Search Plugin",
			Version:     "1.0.0",
			Description: "提供网络搜索能力，帮助查找最新信息",
			Author:      "Tortoise Team",
			Enabled:     true,
			Tools: []Tool{
				{
					Name:        "search",
					Description: "搜索网络内容",
					Parameters: []Parameter{
						{Name: "query", Type: "string", Required: true},
					},
				},
			},
			Status: "active",
		},
		{
			ID:          uuid.New().String(),
			Name:        "Calculator",
			Version:     "1.2.0",
			Description: "数学计算插件，支持复杂表达式",
			Author:      "Tortoise Team",
			Enabled:     true,
			Tools: []Tool{
				{
					Name:        "calculate",
					Description: "执行数学计算",
					Parameters: []Parameter{
						{Name: "expression", Type: "string", Required: true},
					},
				},
			},
			Status: "active",
		},
		{
			ID:          uuid.New().String(),
			Name:        "Weather",
			Version:     "0.9.0",
			Description: "天气查询插件",
			Author:      "Community",
			Enabled:     false,
			Tools: []Tool{
				{
					Name:        "get_weather",
					Description: "获取天气信息",
					Parameters: []Parameter{
						{Name: "city", Type: "string", Required: true},
					},
				},
			},
			Status: "inactive",
		},
		{
			ID:          uuid.New().String(),
			Name:        "File Manager",
			Version:     "2.1.0",
			Description: "文件管理插件",
			Author:      "Tortoise Team",
			Enabled:     true,
			Tools: []Tool{
				{
					Name:        "read_file",
					Description: "读取文件内容",
					Parameters: []Parameter{
						{Name: "path", Type: "string", Required: true},
					},
				},
				{
					Name:        "write_file",
					Description: "写入文件内容",
					Parameters: []Parameter{
						{Name: "path", Type: "string", Required: true},
						{Name: "content", Type: "string", Required: true},
					},
				},
			},
			Status: "active",
		},
	}

	for _, plugin := range plugins {
		p.plugins[plugin.ID] = plugin
	}
}

// GetPlugins 获取所有插件
func (p *PluginStore) GetPlugins() []*Plugin {
	p.mu.RLock()
	defer p.mu.RUnlock()

	plugins := make([]*Plugin, 0, len(p.plugins))
	for _, plugin := range p.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// GetPlugin 获取单个插件
func (p *PluginStore) GetPlugin(id string) (*Plugin, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	plugin, ok := p.plugins[id]
	return plugin, ok
}

// TogglePlugin 切换插件启用状态
func (p *PluginStore) TogglePlugin(id string, enabled bool) (*Plugin, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	plugin, ok := p.plugins[id]
	if !ok {
		return nil, false
	}

	plugin.Enabled = enabled
	if enabled {
		plugin.Status = "active"
	} else {
		plugin.Status = "inactive"
	}
	return plugin, true
}

// DeletePlugin 删除插件
func (p *PluginStore) DeletePlugin(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.plugins[id]; ok {
		delete(p.plugins, id)
		return true
	}
	return false
}
