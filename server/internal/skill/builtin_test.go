package skill

import (
	"context"
	"testing"
)

func TestWebSearchSkill(t *testing.T) {
	skill := &WebSearchSkill{}
	
	if skill.Name() != "web_search" {
		t.Errorf("Expected name 'web_search', got '%s'", skill.Name())
	}
	
	params := map[string]interface{}{
		"query": "test query",
		"limit": float64(5),
	}
	
	if err := skill.Validate(params); err != nil {
		t.Errorf("Validation failed: %v", err)
	}
	
	result, err := skill.Execute(context.Background(), params)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
	
	if result == nil {
		t.Error("Expected result, got nil")
	}
}

func TestCalculatorSkill(t *testing.T) {
	skill := &CalculatorSkill{}
	
	if skill.Name() != "calculator" {
		t.Errorf("Expected name 'calculator', got '%s'", skill.Name())
	}
	
	tests := []struct {
		expr    string
		wantErr bool
	}{
		{"2 + 2", false},
		{"10 - 5", false},
		{"3 * 4", false},
		{"20 / 4", false},
		{"2 ^ 8", false},
		{"sqrt(16)", false},
		{"sin(0)", false},
		{"invalid", true},
	}
	
	for _, tt := range tests {
		params := map[string]interface{}{
			"expression": tt.expr,
		}
		
		result, err := skill.Execute(context.Background(), params)
		
		if tt.wantErr {
			if err == nil && result != nil && !result.Success {
				continue // Expected failure
			}
			if err == nil {
				t.Errorf("Expected error for '%s', got nil", tt.expr)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for '%s': %v", tt.expr, err)
			}
			if result != nil && !result.Success {
				t.Errorf("Calculation failed for '%s': %v", tt.expr, result.Error)
			}
		}
	}
}

func TestFileSystemSkill(t *testing.T) {
	skill := &FileSystemSkill{}
	
	if skill.Name() != "file_system" {
		t.Errorf("Expected name 'file_system', got '%s'", skill.Name())
	}
	
	// Test read operation
	params := map[string]interface{}{
		"operation": "read",
		"path":      "test.txt",
	}
	
	result, err := skill.Execute(context.Background(), params)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
	
	if result == nil {
		t.Error("Expected result, got nil")
	}
}

func TestUnitConverterSkill(t *testing.T) {
	skill := &UnitConverterSkill{}
	
	if skill.Name() != "unit_converter" {
		t.Errorf("Expected name 'unit_converter', got '%s'", skill.Name())
	}
	
	tests := []struct {
		from    string
		to      string
		value    float64
		expected float64
	}{
		{"km", "m", 1, 1000},
		{"m", "km", 1000, 1},
		{"kg", "g", 1, 1000},
		{"celsius", "fahrenheit", 0, 32},
	}
	
	for _, tt := range tests {
		params := map[string]interface{}{
			"from":  tt.from,
			"to":    tt.to,
			"value": tt.value,
		}
		
		result, err := skill.Execute(context.Background(), params)
		if err != nil {
			t.Errorf("Execute failed for %s to %s: %v", tt.from, tt.to, err)
			continue
		}
		
		if result == nil || !result.Success {
			t.Errorf("Conversion failed for %s to %s", tt.from, tt.to)
		}
	}
}

func TestDateTimeSkill(t *testing.T) {
	skill := &DateTimeSkill{}
	
	if skill.Name() != "datetime" {
		t.Errorf("Expected name 'datetime', got '%s'", skill.Name())
	}
	
	params := map[string]interface{}{
		"operation": "now",
	}
	
	result, err := skill.Execute(context.Background(), params)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
	
	if result == nil {
		t.Error("Expected result, got nil")
	}
}

func TestTextProcessingSkill(t *testing.T) {
	skill := &TextProcessingSkill{}
	
	if skill.Name() != "text_processing" {
		t.Errorf("Expected name 'text_processing', got '%s'", skill.Name())
	}
	
	tests := []struct {
		op      string
		text    string
		wantErr bool
	}{
		{"upper", "hello", false},
		{"lower", "HELLO", false},
		{"count", "hello world", false},
		{"reverse", "hello", false},
		{"length", "hello", false},
		{"unknown", "hello", true},
	}
	
	for _, tt := range tests {
		params := map[string]interface{}{
			"operation": tt.op,
			"text":      tt.text,
		}
		
		result, err := skill.Execute(context.Background(), params)
		
		if tt.wantErr {
			if err == nil && result != nil && result.Success {
				t.Errorf("Expected error for op '%s', got success", tt.op)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for op '%s': %v", tt.op, err)
			}
			if result != nil && !result.Success {
				t.Errorf("Operation '%s' failed: %v", tt.op, result.Error)
			}
		}
	}
}

func TestRegisterBuiltInSkills(t *testing.T) {
	registry := NewRegistry()
	RegisterBuiltInSkills(registry)
	
	skills := registry.List()
	if len(skills) < 8 {
		t.Errorf("Expected at least 8 skills, got %d", len(skills))
	}
	
	// Verify specific skills exist
	expectedSkills := []string{
		"web_search",
		"calculator",
		"file_system",
		"unit_converter",
		"datetime",
		"text_processing",
	}
	
	for _, name := range expectedSkills {
		if !registry.Exists(name) {
			t.Errorf("Expected skill '%s' to be registered", name)
		}
	}
}

func TestSkillParameters(t *testing.T) {
	skill := &WebSearchSkill{}
	params := skill.Parameters()
	
	if len(params) == 0 {
		t.Error("Expected parameters, got none")
	}
	
	// Find required query parameter
	found := false
	for _, p := range params {
		if p.Name == "query" && p.Required {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected 'query' parameter to be required")
	}
}
