package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"sync"

	"tortoise/config"
)

// PluginService 插件服务
type PluginService struct {
	plugins  map[string]*Plugin
	mu       sync.RWMutex
	cfg      config.PluginsConfig
	registry string
}

// Plugin 插件
type Plugin struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
	Instance    interface{}            `json:"-"`
}

// PluginInterface 插件接口
type PluginInterface interface {
	Init(config map[string]interface{}) error
	Start() error
	Stop() error
	Name() string
}

// NewPluginService 创建插件服务
func NewPluginService(cfg config.PluginsConfig) *PluginService {
	s := &PluginService{
		plugins:  make(map[string]*Plugin),
		cfg:      cfg,
		registry: cfg.Registry,
	}

	// 创建插件目录
	os.MkdirAll(cfg.Directory, 0755)

	return s
}

// LoadAll 加载所有插件
func (s *PluginService) LoadAll() {
	log.Println("Loading all plugins...")

	// 扫描插件目录
	dir := s.cfg.Directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Failed to read plugins directory: %v", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			manifestPath := filepath.Join(dir, entry.Name(), "manifest.json")
			if data, err := os.ReadFile(manifestPath); err == nil {
				var p Plugin
				if err := json.Unmarshal(data, &p); err == nil {
					s.plugins[p.ID] = &p
				}
			}
		}
	}

	log.Printf("Loaded %d plugins", len(s.plugins))
}

// ListPlugins 列出插件
func (s *PluginService) ListPlugins() []*Plugin {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Plugin, 0, len(s.plugins))
	for _, p := range s.plugins {
		result = append(result, p)
	}
	return result
}

// GetPlugin 获取插件
func (s *PluginService) GetPlugin(id string) (*Plugin, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.plugins[id]
	return p, ok
}

// Install 安装插件
func (s *PluginService) Install(id, version string) error {
	// 从注册表获取插件信息
	resp, err := http.Get(fmt.Sprintf("%s/plugins/%s/%s", s.registry, id, version))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to download plugin: %s", resp.Status)
	}

	// 下载插件
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 保存到插件目录
	pluginDir := filepath.Join(s.cfg.Directory, id)
	os.MkdirAll(pluginDir, 0755)

	manifestPath := filepath.Join(pluginDir, "manifest.json")
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return err
	}

	// 加载插件
	var p Plugin
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}

	s.mu.Lock()
	s.plugins[p.ID] = &p
	s.mu.Unlock()

	log.Printf("Installed plugin: %s", p.Name)
	return nil
}

// Uninstall 卸载插件
func (s *PluginService) Uninstall(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.plugins[id]
	if !ok {
		return fmt.Errorf("Plugin not found: %s", id)
	}

	// 停止插件
	if p.Enabled {
		if p.Instance != nil {
			if pi, ok := p.Instance.(PluginInterface); ok {
				pi.Stop()
			}
		}
	}

	// 删除插件文件
	pluginDir := filepath.Join(s.cfg.Directory, id)
	os.RemoveAll(pluginDir)

	delete(s.plugins, id)
	log.Printf("Uninstalled plugin: %s", id)
	return nil
}

// Enable 启用插件
func (s *PluginService) Enable(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.plugins[id]
	if !ok {
		return fmt.Errorf("Plugin not found: %s", id)
	}

	if p.Enabled {
		return nil
	}

	// 加载并启动插件
	pluginPath := filepath.Join(s.cfg.Directory, id, "plugin.so")
	if _, err := os.Stat(pluginPath); err == nil {
		p.Instance = s.loadPlugin(pluginPath)
		if p.Instance != nil {
			if pi, ok := p.Instance.(PluginInterface); ok {
				if err := pi.Init(p.Config); err != nil {
					return err
				}
				if err := pi.Start(); err != nil {
					return err
				}
			}
		}
	}

	p.Enabled = true
	log.Printf("Enabled plugin: %s", p.Name)
	return nil
}

// Disable 禁用插件
func (s *PluginService) Disable(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.plugins[id]
	if !ok {
		return fmt.Errorf("Plugin not found: %s", id)
	}

	if !p.Enabled {
		return nil
	}

	// 停止插件
	if p.Instance != nil {
		if pi, ok := p.Instance.(PluginInterface); ok {
			pi.Stop()
		}
	}

	p.Enabled = false
	log.Printf("Disabled plugin: %s", p.Name)
	return nil
}

// UnloadAll 卸载所有插件
func (s *PluginService) UnloadAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.plugins {
		if p.Enabled && p.Instance != nil {
			if pi, ok := p.Instance.(PluginInterface); ok {
				pi.Stop()
			}
		}
	}
}

// loadPlugin 加载插件 (使用 Go 插件系统)
func (s *PluginService) loadPlugin(path string) interface{} {
	p, err := plugin.Open(path)
	if err != nil {
		log.Printf("Failed to load plugin: %v", err)
		return nil
	}

	sym, err := p.Lookup("Plugin")
	if err != nil {
		log.Printf("Failed to lookup plugin symbol: %v", err)
		return nil
	}

	return sym
}

// ExecutePlugin 执行插件
func (s *PluginService) ExecutePlugin(id string, method string, args ...interface{}) (interface{}, error) {
	s.mu.RLock()
	p, ok := s.plugins[id]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("Plugin not found: %s", id)
	}

	if !p.Enabled {
		return nil, fmt.Errorf("Plugin not enabled: %s", id)
	}

	// 执行插件方法
	if p.Instance != nil {
		// 这里应该使用反射或接口来调用插件方法
		log.Printf("Executing plugin %s method %s", id, method)
	}

	return nil, nil
}

// UpdateConfig 更新插件配置
func (s *PluginService) UpdateConfig(id string, config map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.plugins[id]
	if !ok {
		return fmt.Errorf("Plugin not found: %s", id)
	}

	p.Config = config

	// 如果插件正在运行，重新初始化
	if p.Enabled && p.Instance != nil {
		if pi, ok := p.Instance.(PluginInterface); ok {
			if err := pi.Init(config); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetPluginsDir 获取插件目录
func (s *PluginService) GetPluginsDir() string {
	return s.cfg.Directory
}
