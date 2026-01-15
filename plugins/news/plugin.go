// News Plugin for Tortoise

package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tortoise/server/plugin"
)

// NewsPlugin provides news aggregation
type NewsPlugin struct {
	apiKey string
}

func (p *NewsPlugin) Name() string {
	return "news"
}

func (p *NewsPlugin) Version() string {
	return "1.0.0"
}

func (p *NewsPlugin) Description() string {
	return "Get latest news and articles"
}

func (p *NewsPlugin) Init(ctx context.Context, config map[string]interface{}) error {
	return nil
}

func (p *NewsPlugin) Execute(ctx context.Context, req *plugin.Request) (*plugin.Response, error) {
	topic, _ := req.Arguments["topic"].(string)
	limit, _ := req.Arguments["limit"].(float64)
	if limit == 0 {
		limit = 5
	}

	// Return sample news data
	articles := []map[string]interface{}{
		{
			"title":       "Breaking: AI advances continue",
			"source":      "Tech News",
			"url":         "https://example.com/ai-advances",
			"publishedAt": "2024-01-15T10:00:00Z",
			"summary":     "Recent developments in AI technology...",
		},
		{
			"title":       "New framework released",
			"source":      "Dev Weekly",
			"url":         "https://example.com/new-framework",
			"publishedAt": "2024-01-14T15:30:00Z",
			"summary":     "A new development framework has been released...",
		},
	}

	data := map[string]interface{}{
		"topic":    topic,
		"count":    len(articles),
		"articles": articles,
	}

	return &plugin.Response{
		Success: true,
		Data:    data,
	}, nil
}

func (p *NewsPlugin) Tools() []plugin.ToolDefinition {
	return []plugin.ToolDefinition{
		{
			Name:        "get_news",
			Description: "Get latest news on a topic",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"topic": map[string]interface{}{
						"type":        "string",
						"description": "News topic or category",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of articles to return",
						"default":    5,
					},
				},
				"required": []string{"topic"},
			},
		},
	}
}

func (p *NewsPlugin) Cleanup() error {
	return nil
}

// ExportPlugin exports the plugin
func ExportPlugin() plugin.Plugin {
	return &NewsPlugin{}
}

func main() {}

// Fetch news from external API
func fetchNews(topic string, limit int) ([]map[string]interface{}, error) {
	// Placeholder - would call external news API
	return []map[string]interface{}{}, nil
}

func fetchJSON(url string) (map[string]interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	return data, nil
}
