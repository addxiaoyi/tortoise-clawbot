package plugin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ============ Plugin Marketplace ============

// MarketplaceConfig 市场配置
type MarketplaceConfig struct {
	MarketURL    string            // 市场服务器 URL
	APIVersion   string           // API 版本
	CacheDir     string           // 插件缓存目录
	CacheTimeout time.Duration    // 缓存超时
	VerifyHash   bool            // 验证哈希
	MirrorURLs   []string        // 镜像 URL
}

// Marketplace 插件市场
type Marketplace struct {
	config     *MarketplaceConfig
	plugins    map[string]*MarketplacePlugin
	cache      map[string]*cacheEntry
	installed  map[string]*InstalledPlugin
	mu         sync.RWMutex
	httpClient *http.Client
}

// MarketplacePlugin 市场插件
type MarketplacePlugin struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Version     string           `json:"version"`
	DisplayName string           `json:"display_name"`
	Description string           `json:"description"`
	Author      string           `json:"author"`
	AuthorURL   string           `json:"author_url"`
	Homepage    string           `json:"homepage"`
	Repository  string           `json:"repository"`
	License     string           `json:"license"`
	Category    PluginCategory   `json:"category"`
	Tags        []string        `json:"tags"`
	IconURL     string           `json:"icon_url"`
	Screenshots []string        `json:"screenshots"`
	
	// 版本信息
	Versions    []PluginVersion  `json:"versions"`
	LatestVersion string         `json:"latest_version"`
	
	// 统计
	Downloads   int64           `json:"downloads"`
	Stars       int64           `json:"stars"`
	Rating      float64         `json:"rating"`
	Reviews     int             `json:"reviews"`
	
	// 安全
	Signed      bool             `json:"signed"`
	SHA256      string          `json:"sha256"`
	
	// 依赖
	Dependencies []string       `json:"dependencies"`
	Dependents   []string       `json:"dependents"`
	
	// 元数据
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	PublishedAt time.Time       `json:"published_at"`
}

// PluginVersion 插件版本
type PluginVersion struct {
	Version     string           `json:"version"`
	SemVer      string           `json:"semver"`
	Description string           `json:"description"`
	Changelog   string          `json:"changelog"`
	DownloadURL string          `json:"download_url"`
	SHA256      string          `json:"sha256"`
	FileSize    int64           `json:"file_size"`
	MinAPIVersion string       `json:"min_api_version"`
	MaxAPIVersion string       `json:"max_api_version"`
	Platforms   []string        `json:"platforms"`
	CreatedAt   time.Time       `json:"created_at"`
}

// InstalledPlugin 已安装插件
type InstalledPlugin struct {
	PluginID    string           `json:"plugin_id"`
	Name        string           `json:"name"`
	Version     string           `json:"version"`
	Path        string           `json:"path"`
	Enabled     bool             `json:"enabled"`
	AutoUpdate  bool             `json:"auto_update"`
	
	// 运行时信息
	LoadedAt    time.Time       `json:"loaded_at"`
	ErrorCount  int             `json:"error_count"`
	LastError   string          `json:"last_error,omitempty"`
	
	// 配置
	Config      map[string]interface{} `json:"config"`
}

// cacheEntry 缓存条目
type cacheEntry struct {
	plugin   *MarketplacePlugin
	expiresAt time.Time
}

// NewMarketplace 创建市场
func NewMarketplace(config *MarketplaceConfig) *Marketplace {
	if config.CacheTimeout == 0 {
		config.CacheTimeout = 1 * time.Hour
	}
	if config.CacheDir == "" {
		config.CacheDir = "./plugins/cache"
	}
	
	m := &Marketplace{
		config:    config,
		plugins:   make(map[string]*MarketplacePlugin),
		cache:    make(map[string]*cacheEntry),
		installed: make(map[string]*InstalledPlugin),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	
	// 确保缓存目录存在
	os.MkdirAll(config.CacheDir, 0755)
	
	// 加载已安装插件
	m.loadInstalled()
	
	return m
}

// loadInstalled 加载已安装插件
func (m *Marketplace) loadInstalled() {
	// 从文件系统扫描已安装插件
	pluginDirs, _ := filepath.Glob(filepath.Join(m.config.CacheDir, "..", "*"))
	
	for _, dir := range pluginDirs {
		manifestFile := filepath.Join(dir, "manifest.json")
		if _, err := os.Stat(manifestFile); err == nil {
			data, _ := os.ReadFile(manifestFile)
			var plugin MarketplacePlugin
			if json.Unmarshal(data, &plugin) == nil {
				m.installed[plugin.ID] = &InstalledPlugin{
					PluginID: plugin.ID,
					Name:     plugin.Name,
					Version:  plugin.Version,
					Path:     dir,
					Enabled:  true,
				}
			}
		}
	}
}

// Search 搜索插件
func (m *Marketplace) Search(query string, opts *SearchOptions) ([]*MarketplacePlugin, error) {
	// 检查缓存
	m.mu.RLock()
	if entry, ok := m.cache["search:"+query]; ok && time.Now().Before(entry.expiresAt) {
		m.mu.RUnlock()
		return m.pluginsSlice(opts), nil
	}
	m.mu.RUnlock()
	
	// 从市场 API 获取
	url := fmt.Sprintf("%s/api/v1/plugins/search?q=%s", m.config.MarketURL, url.QueryEscape(query))
	
	resp, err := m.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}
	defer resp.Body.Close()
	
	var plugins []*MarketplacePlugin
	if err := json.NewDecoder(resp.Body).Decode(&plugins); err != nil {
		return nil, fmt.Errorf("解析失败: %w", err)
	}
	
	// 更新缓存
	m.mu.Lock()
	for _, p := range plugins {
		m.plugins[p.ID] = p
	}
	m.cache["search:"+query] = &cacheEntry{
		plugin:    nil,
		expiresAt: time.Now().Add(m.config.CacheTimeout),
	}
	m.mu.Unlock()
	
	return m.pluginsSlice(opts), nil
}

// SearchOptions 搜索选项
type SearchOptions struct {
	Category  PluginCategory
	Tags      []string
	Author    string
	Platform  string
	MinRating float64
	SortBy    string // downloads, stars, rating, recent
	Page      int
	PageSize  int
}

// ListCategories 列出分类
func (m *Marketplace) ListCategories() ([]*CategoryInfo, error) {
	url := fmt.Sprintf("%s/api/v1/categories", m.config.MarketURL)
	
	resp, err := m.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var categories []*CategoryInfo
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		return nil, err
	}
	
	return categories, nil
}

// CategoryInfo 分类信息
type CategoryInfo struct {
	ID          PluginCategory `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	IconURL     string        `json:"icon_url"`
	PluginCount int           `json:"plugin_count"`
}

// pluginsSlice 返回插件列表
func (m *Marketplace) pluginsSlice(opts *SearchOptions) []*MarketplacePlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	plugins := make([]*MarketplacePlugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		plugins = append(plugins, p)
	}
	
	// 过滤
	if opts != nil {
		if opts.Category != "" {
			filtered := make([]*MarketplacePlugin, 0)
			for _, p := range plugins {
				if p.Category == opts.Category {
					filtered = append(filtered, p)
				}
			}
			plugins = filtered
		}
		
		if opts.MinRating > 0 {
			filtered := make([]*MarketplacePlugin, 0)
			for _, p := range plugins {
				if p.Rating >= opts.MinRating {
					filtered = append(filtered, p)
				}
			}
			plugins = filtered
		}
		
		// 排序
		switch opts.SortBy {
		case "downloads":
			sort.Slice(plugins, func(i, j int) bool {
				return plugins[i].Downloads > plugins[j].Downloads
			})
		case "stars":
			sort.Slice(plugins, func(i, j int) bool {
				return plugins[i].Stars > plugins[j].Stars
			})
		case "rating":
			sort.Slice(plugins, func(i, j int) bool {
				return plugins[i].Rating > plugins[j].Rating
			})
		case "recent":
			sort.Slice(plugins, func(i, j int) bool {
				return plugins[i].UpdatedAt.After(plugins[j].UpdatedAt)
			})
		}
		
		// 分页
		if opts.PageSize > 0 {
			start := opts.Page * opts.PageSize
			if start < len(plugins) {
				end := start + opts.PageSize
				if end > len(plugins) {
					end = len(plugins)
				}
				plugins = plugins[start:end]
			} else {
				plugins = plugins[:0]
			}
		}
	}
	
	return plugins
}

// GetPlugin 获取插件详情
func (m *Marketplace) GetPlugin(pluginID string) (*MarketplacePlugin, error) {
	// 检查本地缓存
	m.mu.RLock()
	if p, ok := m.plugins[pluginID]; ok {
		m.mu.RUnlock()
		return p, nil
	}
	m.mu.RUnlock()
	
	// 从 API 获取
	url := fmt.Sprintf("%s/api/v1/plugins/%s", m.config.MarketURL, pluginID)
	
	resp, err := m.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("插件不存在: %s", pluginID)
	}
	
	var plugin MarketplacePlugin
	if err := json.NewDecoder(resp.Body).Decode(&plugin); err != nil {
		return nil, err
	}
	
	// 更新缓存
	m.mu.Lock()
	m.plugins[plugin.ID] = &plugin
	m.mu.Unlock()
	
	return &plugin, nil
}

// Install 安装插件
func (m *Marketplace) Install(ctx context.Context, pluginID, version string) (*InstalledPlugin, error) {
	// 获取插件信息
	plugin, err := m.GetPlugin(pluginID)
	if err != nil {
		return nil, err
	}
	
	// 选择版本
	var downloadURL, sha256 string
	if version == "" {
		// 使用最新版本
		downloadURL = plugin.Versions[0].DownloadURL
		sha256 = plugin.Versions[0].SHA256
		version = plugin.Versions[0].Version
	} else {
		// 查找指定版本
		for _, v := range plugin.Versions {
			if v.Version == version {
				downloadURL = v.DownloadURL
				sha256 = v.SHA256
				break
			}
		}
		if downloadURL == "" {
			return nil, fmt.Errorf("版本不存在: %s", version)
		}
	}
	
	// 下载插件
	pluginPath := filepath.Join(m.config.CacheDir, pluginID, version)
	if err := os.MkdirAll(pluginPath, 0755); err != nil {
		return nil, err
	}
	
	downloadPath := filepath.Join(pluginPath, "plugin.tar.gz")
	if err := m.downloadFile(ctx, downloadURL, downloadPath); err != nil {
		return nil, fmt.Errorf("下载失败: %w", err)
	}
	
	// 验证哈希
	if m.config.VerifyHash {
		if err := m.verifyHash(downloadPath, sha256); err != nil {
			return nil, fmt.Errorf("验证失败: %w", err)
		}
	}
	
	// 解压
	if err := m.extractPlugin(downloadPath, pluginPath); err != nil {
		return nil, fmt.Errorf("解压失败: %w", err)
	}
	
	// 记录已安装
	installed := &InstalledPlugin{
		PluginID: pluginID,
		Name:     plugin.Name,
		Version:  version,
		Path:     pluginPath,
		Enabled:  true,
		LoadedAt: time.Now(),
	}
	
	m.mu.Lock()
	m.installed[pluginID] = installed
	m.mu.Unlock()
	
	// 保存 manifest
	manifest, _ := json.MarshalIndent(plugin, "", "  ")
	os.WriteFile(filepath.Join(pluginPath, "manifest.json"), manifest, 0644)
	
	log.Printf("✅ 插件已安装: %s@%s", plugin.Name, version)
	
	return installed, nil
}

// downloadFile 下载文件
func (m *Marketplace) downloadFile(ctx context.Context, url, dest string) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	
	_, err = io.Copy(out, resp.Body)
	return err
}

// verifyHash 验证哈希
func (m *Marketplace) verifyHash(filePath, expectedHash string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	
	hash := sha256.Sum256(data)
	actualHash := hex.EncodeToString(hash[:])
	
	if actualHash != expectedHash {
		return fmt.Errorf("哈希不匹配: expected %s, got %s", expectedHash, actualHash)
	}
	
	return nil
}

// extractPlugin 解压插件
func (m *Marketplace) extractPlugin(archive, dest string) error {
	// 简化实现 - 实际应使用 tar/gzip 库
	return nil
}

// Uninstall 卸载插件
func (m *Marketplace) Uninstall(pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	installed, ok := m.installed[pluginID]
	if !ok {
		return fmt.Errorf("插件未安装: %s", pluginID)
	}
	
	// 删除文件
	if err := os.RemoveAll(installed.Path); err != nil {
		return err
	}
	
	delete(m.installed, pluginID)
	
	log.Printf("✅ 插件已卸载: %s", pluginID)
	return nil
}

// Update 更新插件
func (m *Marketplace) Update(ctx context.Context, pluginID string) (*InstalledPlugin, error) {
	plugin, err := m.GetPlugin(pluginID)
	if err != nil {
		return nil, err
	}
	
	// 获取最新版本
	latestVersion := plugin.Versions[0].Version
	
	m.mu.RLock()
	installed, ok := m.installed[pluginID]
	m.mu.RUnlock()
	
	if !ok {
		return nil, fmt.Errorf("插件未安装: %s", pluginID)
	}
	
	if installed.Version == latestVersion {
		return installed, nil // 已是最新
	}
	
	// 重新安装
	return m.Install(ctx, pluginID, latestVersion)
}

// ListInstalled 列出已安装插件
func (m *Marketplace) ListInstalled() []*InstalledPlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	plugins := make([]*InstalledPlugin, 0, len(m.installed))
	for _, p := range m.installed {
		plugins = append(plugins, p)
	}
	
	return plugins
}

// Enable 启用插件
func (m *Marketplace) Enable(pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	installed, ok := m.installed[pluginID]
	if !ok {
		return fmt.Errorf("插件未安装: %s", pluginID)
	}
	
	installed.Enabled = true
	log.Printf("✅ 插件已启用: %s", pluginID)
	
	return nil
}

// Disable 禁用插件
func (m *Marketplace) Disable(pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	installed, ok := m.installed[pluginID]
	if !ok {
		return fmt.Errorf("插件未安装: %s", pluginID)
	}
	
	installed.Enabled = false
	log.Printf("✅ 插件已禁用: %s", pluginID)
	
	return nil
}

// GetFeatured 获取精选插件
func (m *Marketplace) GetFeatured() ([]*MarketplacePlugin, error) {
	url := fmt.Sprintf("%s/api/v1/plugins/featured", m.config.MarketURL)
	
	resp, err := m.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var plugins []*MarketplacePlugin
	if err := json.NewDecoder(resp.Body).Decode(&plugins); err != nil {
		return nil, err
	}
	
	return plugins, nil
}

// GetPopular 获取热门插件
func (m *Marketplace) GetPopular() ([]*MarketplacePlugin, error) {
	return m.Search("", &SearchOptions{SortBy: "downloads", PageSize: 20})
}

// GetRecommended 获取推荐插件
func (m *Marketplace) GetRecommended() ([]*MarketplacePlugin, error) {
	url := fmt.Sprintf("%s/api/v1/plugins/recommended", m.config.MarketURL)
	
	resp, err := m.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var plugins []*MarketplacePlugin
	if err := json.NewDecoder(resp.Body).Decode(&plugins); err != nil {
		return nil, err
	}
	
	return plugins, nil
}

// SubmitPlugin 提交插件到市场
func (m *Marketplace) SubmitPlugin(plugin *MarketplacePlugin) error {
	url := fmt.Sprintf("%s/api/v1/plugins", m.config.MarketURL)
	
	body, _ := json.Marshal(plugin)
	resp, err := m.httpClient.Post(url, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("提交失败: HTTP %d", resp.StatusCode)
	}
	
	return nil
}

// PublishRelease 发布新版本
func (m *Marketplace) PublishRelease(pluginID string, release *PluginVersion) error {
	url := fmt.Sprintf("%s/api/v1/plugins/%s/releases", m.config.MarketURL, pluginID)
	
	body, _ := json.Marshal(release)
	resp, err := m.httpClient.Post(url, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("发布失败: HTTP %d", resp.StatusCode)
	}
	
	return nil
}

// ValidateManifest 验证插件清单
func ValidateManifest(data []byte) (*MarketplacePlugin, error) {
	var manifest struct {
		Name        string          `json:"name"`
		Version     string          `json:"version"`
		Description string          `json:"description"`
		Author      string          `json:"author"`
		Category    PluginCategory  `json:"category"`
		Platforms   []string        `json:"platforms"`
		Dependencies []string       `json:"dependencies"`
	}
	
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("无效的 manifest: %w", err)
	}
	
	// 验证名称格式
	if !regexp.MustCompile(`^[a-z][a-z0-9-]*$`).MatchString(manifest.Name) {
		return nil, fmt.Errorf("无效的插件名称: %s", manifest.Name)
	}
	
	// 验证版本格式
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(manifest.Version) {
		return nil, fmt.Errorf("无效的版本格式: %s", manifest.Version)
	}
	
	return &MarketplacePlugin{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		Category:    manifest.Category,
	}, nil
}

// CreatePluginManifest 创建插件清单
func CreatePluginManifest(plugin *MarketplacePlugin) ([]byte, error) {
	manifest := map[string]interface{}{
		"schema_version": "1.0",
		"name":           plugin.Name,
		"version":        plugin.Version,
		"display_name":   plugin.DisplayName,
		"description":    plugin.Description,
		"author":         plugin.Author,
		"author_url":     plugin.AuthorURL,
		"homepage":       plugin.Homepage,
		"repository":     plugin.Repository,
		"license":        plugin.License,
		"category":       plugin.Category,
		"tags":           plugin.Tags,
		"platforms":      []string{"linux", "darwin", "windows"},
		"dependencies":   plugin.Dependencies,
		"api_version":    "1.0",
		"id":             uuid.New().String(),
	}
	
	return json.MarshalIndent(manifest, "", "  ")
}

// SecurityScan 安全扫描
type SecurityReport struct {
	PluginID    string      `json:"plugin_id"`
	ScanDate    time.Time   `json:"scan_date"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	Permissions []Permission     `json:"permissions"`
	Score       float64     `json:"score"`
	Passed      bool        `json:"passed"`
}

// Vulnerability 漏洞
type Vulnerability struct {
	ID          string `json:"id"`
	Severity    string `json:"severity"` // critical, high, medium, low
	Description string `json:"description"`
	Recommendation string `json:"recommendation"`
}

// Permission 权限
type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Dangerous   bool   `json:"dangerous"`
}

// ScanSecurity 安全扫描
func (m *Marketplace) ScanSecurity(pluginID string) (*SecurityReport, error) {
	plugin, err := m.GetPlugin(pluginID)
	if err != nil {
		return nil, err
	}
	
	report := &SecurityReport{
		PluginID:    pluginID,
		ScanDate:    time.Now(),
		Vulnerabilities: make([]Vulnerability, 0),
		Permissions: make([]Permission, 0),
		Score:       100,
		Passed:      true,
	}
	
	// 检查签名
	if !plugin.Signed {
		report.Vulnerabilities = append(report.Vulnerabilities, Vulnerability{
			ID:          "UNSIGNED",
			Severity:    "medium",
			Description: "插件未签名",
		})
		report.Score -= 10
	}
	
	// 检查权限
	report.Permissions = append(report.Permissions, Permission{
		Name:        "network",
		Description: "网络访问",
		Dangerous:   true,
	})
	
	if report.Score < 70 {
		report.Passed = false
	}
	
	return report, nil
}
