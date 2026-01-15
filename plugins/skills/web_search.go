// Skills 包 - 技能插件
package skills

import (
	"context"
	"encoding/json"
)

// SearchResult 搜索结果
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// WebSearchSkill Web 搜索技能
type WebSearchSkill struct {
	apiKey string
}

// NewWebSearchSkill 创建 Web 搜索技能
func NewWebSearchSkill(apiKey string) *WebSearchSkill {
	return &WebSearchSkill{apiKey: apiKey}
}

// Execute 执行技能
func (s *WebSearchSkill) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	
	// 模拟搜索
	results := []SearchResult{
		{Title: "Result 1", URL: "https://example.com/1", Snippet: "Sample result 1"},
		{Title: "Result 2", URL: "https://example.com/2", Snippet: "Sample result 2"},
	}
	
	return json.Marshal(map[string]interface{}{
		"query":   params.Query,
		"results": results,
	})
}

// Metadata 返回技能元数据
func (s *WebSearchSkill) Metadata() *SkillMetadata {
	return &SkillMetadata{
		Name:        "web_search",
		Description: "Search the web for information",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The search query",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results",
					"default":     5,
				},
			},
			"required": []string{"query"},
		},
	}
}

// SkillMetadata 技能元数据
type SkillMetadata struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}
