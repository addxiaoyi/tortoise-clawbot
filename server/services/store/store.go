package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PluginStore 插件商店
type PluginStore struct {
	registryURL string
	client     *http.Client
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Homepage    string            `json:"homepage"`
	Repository  string            `json:"repository"`
	Keywords    []string          `json:"keywords"`
	Channels    []string          `json:"channels"`
	Downloads   int               `json:"downloads"`
	Stars       int               `json:"stars"`
	Rating      float64           `json:"rating"`
	Verified    bool              `json:"verified"`
	License     string            `json:"license"`
	Size        string            `json:"size"`
	PublishedAt time.Time        `json:"published_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Manifest    *PluginManifest   `json:"manifest,omitempty"`
}

// PluginManifest 插件清单
type PluginManifest struct {
	Functions []FunctionDef `json:"functions"`
	Config    map[string]ConfigField `json:"config"`
}

// FunctionDef 函数定义
type FunctionDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// ConfigField 配置字段
type ConfigField struct {
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Default     any      `json:"default,omitempty"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Plugins    []*PluginInfo `json:"plugins"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
}

// Review 评价
type Review struct {
	ID        string    `json:"id"`
	PluginID  string    `json:"plugin_id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

// NewPluginStore 创建插件商店
func NewPluginStore(registryURL string) *PluginStore {
	return &PluginStore{
		registryURL: registryURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Search 搜索插件
func (s *PluginStore) Search(ctx context.Context, query string, page, perPage int) (*SearchResult, error) {
	url := fmt.Sprintf("%s/plugins/search?q=%s&page=%d&per_page=%d",
		s.registryURL, query, page, perPage)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %d", resp.StatusCode)
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetPlugin 获取插件详情
func (s *PluginStore) GetPlugin(ctx context.Context, id string) (*PluginInfo, error) {
	url := fmt.Sprintf("%s/plugins/%s", s.registryURL, id)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("plugin not found: %d", resp.StatusCode)
	}

	var plugin PluginInfo
	if err := json.NewDecoder(resp.Body).Decode(&plugin); err != nil {
		return nil, err
	}

	return &plugin, nil
}

// ListFeatured 列出推荐插件
func (s *PluginStore) ListFeatured(ctx context.Context, limit int) ([]*PluginInfo, error) {
	url := fmt.Sprintf("%s/plugins/featured?limit=%d", s.registryURL, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var plugins []*PluginInfo
	if err := json.NewDecoder(resp.Body).Decode(&plugins); err != nil {
		return nil, err
	}

	return plugins, nil
}

// ListPopular 列出热门插件
func (s *PluginStore) ListPopular(ctx context.Context, category string, limit int) ([]*PluginInfo, error) {
	url := fmt.Sprintf("%s/plugins/popular?category=%s&limit=%d",
		s.registryURL, category, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var plugins []*PluginInfo
	if err := json.NewDecoder(resp.Body).Decode(&plugins); err != nil {
		return nil, err
	}

	return plugins, nil
}

// Download 下载插件
func (s *PluginStore) Download(ctx context.Context, id, version, destPath string) error {
	url := fmt.Sprintf("%s/plugins/%s/%s/download", s.registryURL, id, version)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %d", resp.StatusCode)
	}

	// 保存到文件
	file, err := createFile(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// GetReviews 获取评价
func (s *PluginStore) GetReviews(ctx context.Context, pluginID string, page, perPage int) ([]*Review, error) {
	url := fmt.Sprintf("%s/plugins/%s/reviews?page=%d&per_page=%d",
		s.registryURL, pluginID, page, perPage)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var reviews []*Review
	if err := json.NewDecoder(resp.Body).Decode(&reviews); err != nil {
		return nil, err
	}

	return reviews, nil
}

// SubmitReview 提交评价
func (s *PluginStore) SubmitReview(ctx context.Context, review *Review) error {
	url := fmt.Sprintf("%s/plugins/%s/reviews", s.registryURL, review.PluginID)

	data, err := json.Marshal(review)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("review submission failed: %d", resp.StatusCode)
	}

	return nil
}

// GetCategories 获取分类
func (s *PluginStore) GetCategories(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("%s/plugins/categories", s.registryURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var categories []string
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		return nil, err
	}

	return categories, nil
}

// 辅助函数
func createFile(path string) (*os.File, error) {
	return os.Create(path)
}
