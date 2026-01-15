package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"tortoise/plugin"
)

// MarketplaceHandler 插件市场处理器
type MarketplaceHandler struct {
	marketplace *plugin.Marketplace
}

// NewMarketplaceHandler 创建市场处理器
func NewMarketplaceHandler(marketplace *plugin.Marketplace) *MarketplaceHandler {
	return &MarketplaceHandler{marketplace: marketplace}
}

// RegisterRoutes 注册路由
func (h *MarketplaceHandler) RegisterRoutes(r *gin.RouterGroup) {
	// 搜索和发现
	r.GET("/plugins", h.SearchPlugins)
	r.GET("/plugins/featured", h.GetFeatured)
	r.GET("/plugins/popular", h.GetPopular)
	r.GET("/plugins/recommended", h.GetRecommended)
	r.GET("/plugins/categories", h.ListCategories)
	
	// 插件详情
	r.GET("/plugins/:id", h.GetPlugin)
	
	// 安装管理
	r.POST("/plugins/:id/install", h.InstallPlugin)
	r.POST("/plugins/:id/update", h.UpdatePlugin)
	r.DELETE("/plugins/:id", h.UninstallPlugin)
	
	// 已安装插件
	r.GET("/installed", h.ListInstalled)
	r.POST("/installed/:id/enable", h.EnablePlugin)
	r.POST("/installed/:id/disable", h.DisablePlugin)
	
	// 安全扫描
	r.GET("/plugins/:id/security", h.ScanSecurity)
	
	// 开发者
	r.POST("/plugins", h.SubmitPlugin)
	r.POST("/plugins/:id/releases", h.PublishRelease)
}

// SearchPlugins 搜索插件
func (h *MarketplaceHandler) SearchPlugins(c *gin.Context) {
	query := c.Query("q")
	
	opts := &plugin.SearchOptions{
		SortBy:   c.DefaultQuery("sort", "downloads"),
		Page:     0,
		PageSize: 20,
	}
	
	if category := c.Query("category"); category != "" {
		opts.Category = plugin.PluginCategory(category)
	}
	
	plugins, err := h.marketplace.Search(query, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"plugins": plugins,
		"total":   len(plugins),
	})
}

// GetFeatured 获取精选插件
func (h *MarketplaceHandler) GetFeatured(c *gin.Context) {
	plugins, err := h.marketplace.GetFeatured()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"plugins": plugins})
}

// GetPopular 获取热门插件
func (h *MarketplaceHandler) GetPopular(c *gin.Context) {
	plugins, err := h.marketplace.GetPopular()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"plugins": plugins})
}

// GetRecommended 获取推荐插件
func (h *MarketplaceHandler) GetRecommended(c *gin.Context) {
	plugins, err := h.marketplace.GetRecommended()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"plugins": plugins})
}

// ListCategories 列出分类
func (h *MarketplaceHandler) ListCategories(c *gin.Context) {
	categories, err := h.marketplace.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// GetPlugin 获取插件详情
func (h *MarketplaceHandler) GetPlugin(c *gin.Context) {
	pluginID := c.Param("id")
	
	pluginInfo, err := h.marketplace.GetPlugin(pluginID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "插件不存在"})
		return
	}
	
	c.JSON(http.StatusOK, pluginInfo)
}

// InstallPlugin 安装插件
func (h *MarketplaceHandler) InstallPlugin(c *gin.Context) {
	pluginID := c.Param("id")
	version := c.Query("version")
	
	var req struct {
		Version string `json:"version"`
	}
	c.ShouldBindJSON(&req)
	
	if req.Version != "" {
		version = req.Version
	}
	
	installed, err := h.marketplace.Install(context.Background(), pluginID, version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "插件安装成功",
		"plugin": installed,
	})
}

// UpdatePlugin 更新插件
func (h *MarketplaceHandler) UpdatePlugin(c *gin.Context) {
	pluginID := c.Param("id")
	
	installed, err := h.marketplace.Update(context.Background(), pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "插件更新成功",
		"plugin": installed,
	})
}

// UninstallPlugin 卸载插件
func (h *MarketplaceHandler) UninstallPlugin(c *gin.Context) {
	pluginID := c.Param("id")
	
	if err := h.marketplace.Uninstall(pluginID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "插件卸载成功",
	})
}

// ListInstalled 列出已安装插件
func (h *MarketplaceHandler) ListInstalled(c *gin.Context) {
	plugins := h.marketplace.ListInstalled()
	
	c.JSON(http.StatusOK, gin.H{
		"plugins": plugins,
		"total":   len(plugins),
	})
}

// EnablePlugin 启用插件
func (h *MarketplaceHandler) EnablePlugin(c *gin.Context) {
	pluginID := c.Param("id")
	
	if err := h.marketplace.Enable(pluginID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "插件已启用",
	})
}

// DisablePlugin 禁用插件
func (h *MarketplaceHandler) DisablePlugin(c *gin.Context) {
	pluginID := c.Param("id")
	
	if err := h.marketplace.Disable(pluginID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "插件已禁用",
	})
}

// ScanSecurity 安全扫描
func (h *MarketplaceHandler) ScanSecurity(c *gin.Context) {
	pluginID := c.Param("id")
	
	report, err := h.marketplace.ScanSecurity(pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, report)
}

// SubmitPlugin 提交插件
func (h *MarketplaceHandler) SubmitPlugin(c *gin.Context) {
	var pluginData struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
		Author      string `json:"author"`
		Category    string `json:"category"`
	}
	
	if err := c.ShouldBindJSON(&pluginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	submission := &plugin.MarketplacePlugin{
		ID:          uuid.New().String(),
		Name:        pluginData.Name,
		Version:     pluginData.Version,
		Description: pluginData.Description,
		Author:      pluginData.Author,
		Category:    plugin.PluginCategory(pluginData.Category),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		PublishedAt: time.Now(),
	}
	
	if err := h.marketplace.SubmitPlugin(submission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "插件提交成功",
		"plugin":  submission,
	})
}

// PublishRelease 发布新版本
func (h *MarketplaceHandler) PublishRelease(c *gin.Context) {
	pluginID := c.Param("id")
	
	var releaseData struct {
		Version     string `json:"version"`
		Description string `json:"description"`
		Changelog   string `json:"changelog"`
		DownloadURL string `json:"download_url"`
		SHA256      string `json:"sha256"`
		FileSize    int64  `json:"file_size"`
	}
	
	if err := c.ShouldBindJSON(&releaseData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	release := &plugin.PluginVersion{
		Version:     releaseData.Version,
		Description: releaseData.Description,
		Changelog:   releaseData.Changelog,
		DownloadURL: releaseData.DownloadURL,
		SHA256:      releaseData.SHA256,
		FileSize:    releaseData.FileSize,
		CreatedAt:   time.Now(),
	}
	
	if err := h.marketplace.PublishRelease(pluginID, release); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "版本发布成功",
		"release": release,
	})
}

// ============ 请求/响应结构 ============

type PluginSearchRequest struct {
	Query    string `form:"q"`
	Category string `form:"category"`
	SortBy   string `form:"sort"` // downloads, stars, rating, recent
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type PluginInstallRequest struct {
	Version string `json:"version"`
}

type PluginSubmitRequest struct {
	Name        string `json:"name" binding:"required"`
	Version     string `json:"version" binding:"required"`
	DisplayName string `json:"display_name"`
	Description string `json:"description" binding:"required"`
	Author      string `json:"author" binding:"required"`
	Category    string `json:"category" binding:"required"`
	Tags        []string `json:"tags"`
	License     string `json:"license"`
	Repository  string `json:"repository"`
}

type PluginResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Category    string                 `json:"category"`
	Downloads   int64                 `json:"downloads"`
	Rating      float64                `json:"rating"`
	Versions    []*plugin.PluginVersion `json:"versions,omitempty"`
	Installed   *plugin.InstalledPlugin `json:"installed,omitempty"`
}

func (h *MarketplaceHandler) SearchPluginsJSON(c *gin.Context) {
	var req PluginSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	opts := &plugin.SearchOptions{
		SortBy:   req.SortBy,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	
	if req.Category != "" {
		opts.Category = plugin.PluginCategory(req.Category)
	}
	
	plugins, err := h.marketplace.Search(req.Query, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// 转换响应
	response := make([]PluginResponse, 0, len(plugins))
	for _, p := range plugins {
		response = append(response, PluginResponse{
			ID:          p.ID,
			Name:        p.Name,
			Version:     p.Version,
			Description: p.Description,
			Author:      p.Author,
			Category:    string(p.Category),
			Downloads:   p.Downloads,
			Rating:      p.Rating,
			Versions:    p.Versions,
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"plugins": response,
		"total":   len(response),
		"page":    req.Page,
		"page_size": req.PageSize,
	})
}

// MarshalJSON 实现 json.Marshaler
func (p *PluginResponse) MarshalJSON() ([]byte, error) {
	type Alias PluginResponse
	return json.Marshal((*Alias)(p))
}
