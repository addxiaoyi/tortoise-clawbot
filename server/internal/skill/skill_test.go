package skill

import (
	"context"
	"testing"
)

// TestWebSearchSkill 测试网络搜索技能
func TestWebSearchSkill(t *testing.T) {
	skill := &WebSearchSkill{}
	
	// 测试名称和描述
	if skill.Name() != "web_search" {
		t.Errorf("Expected name 'web_search', got '%s'", skill.Name())
	}
	
	if skill.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", skill.Version())
	}
	
	// 测试参数
	params := skill.Parameters()
	if len(params) != 3 {
		t.Errorf("Expected 3 parameters, got %d", len(params))
	}
	
	// 测试执行
	result, err := skill.Execute(context.Background(), map[string]interface{}{
		"query": "test query",
	})
	
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
	
	if result == nil {
		t.Error("Expected result, got nil")
	}
}

// TestCalculatorSkill 测试计算器技能
func TestCalculatorSkill(t *testing.T) {
	skill := &CalculatorSkill{}
	
	// 测试基本计算
	result, err := skill.Execute(context.Background(), map[string]interface{}{
		"expression": "2 + 2",
	})
	
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
	
	if result.Success {
		t.Error("Expected calculation error for invalid expression")
	}
}

// TestSkillRegistry 测试技能注册表
func TestSkillRegistry(t *testing.T) {
	registry := NewSkillRegistry()
	
	// 注册测试技能
	testSkill := &WebSearchSkill{}
	registry.Register(testSkill)
	
	// 获取技能
	retrieved := registry.Get("web_search")
	if retrieved == nil {
		t.Error("Expected to retrieve web_search skill")
	}
	
	// 获取不存在的技能
	missing := registry.Get("nonexistent")
	if missing != nil {
		t.Error("Expected nil for nonexistent skill")
	}
	
	// 列出所有技能
	skills := registry.ListAll()
	if len(skills) == 0 {
		t.Error("Expected at least one skill")
	}
}
