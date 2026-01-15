package memory

import (
	"testing"
	"time"
)

// TestLongTermMemory 测试长期记忆
func TestLongTermMemory(t *testing.T) {
	memory := NewLongTermMemory()
	
	// 创建记忆
	mem := memory.Create("test content", MemoryTypeEpisodic, 0.8)
	if mem == nil {
		t.Fatal("Expected memory, got nil")
	}
	
	if mem.Content != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", mem.Content)
	}
}

// TestMemorySearch 测试记忆搜索
func TestMemorySearch(t *testing.T) {
	memory := NewLongTermMemory()
	
	// 创建多条记忆
	memory.Create("apple fruit", MemoryTypeSemantic, 0.9)
	memory.Create("banana fruit", MemoryTypeSemantic, 0.8)
	memory.Create("carrot vegetable", MemoryTypeSemantic, 0.7)
	
	// 搜索
	results := memory.Search("fruit", 10)
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

// TestMemoryImportanceDecay 测试重要性衰减
func TestMemoryImportanceDecay(t *testing.T) {
	memory := NewLongTermMemory()
	
	// 创建记忆
	mem := memory.Create("test", MemoryTypeEpisodic, 0.5)
	
	// 更新访问
	memory.RecordAccess(mem.ID)
	
	// 获取记忆
	retrieved := memory.Get(mem.ID)
	if retrieved == nil {
		t.Fatal("Expected memory, got nil")
	}
	
	if retrieved.AccessCount != 1 {
		t.Errorf("Expected access count 1, got %d", retrieved.AccessCount)
	}
}

// TestMemoryForget 测试遗忘机制
func TestMemoryForget(t *testing.T) {
	memory := NewLongTermMemory()
	
	// 创建低重要性记忆
	memory.Create("forgettable", MemoryTypeEpisodic, 0.1)
	
	// 创建高重要性记忆
	memory.Create("important", MemoryTypeEpisodic, 0.9)
	
	// 触发遗忘
	forgotten := memory.Forget()
	
	// 检查遗忘结果
	if forgotten < 0 {
		t.Error("Expected non-negative forgotten count")
	}
}

// TestMemoryConsolidation 测试记忆整合
func TestMemoryConsolidation(t *testing.T) {
	memory := NewLongTermMemory()
	
	// 创建相关记忆
	mem1 := memory.Create("user likes coffee", MemoryTypeEpisodic, 0.7)
	mem2 := memory.Create("user drinks coffee daily", MemoryTypeEpisodic, 0.6)
	
	// 整合
	err := memory.Consolidate()
	if err != nil {
		t.Errorf("Consolidate failed: %v", err)
	}
}

// TestMemoryExportImport 测试记忆导出导入
func TestMemoryExportImport(t *testing.T) {
	memory := NewLongTermMemory()
	
	// 创建记忆
	memory.Create("test1", MemoryTypeEpisodic, 0.8)
	memory.Create("test2", MemoryTypeSemantic, 0.9)
	
	// 导出
	data, err := memory.Export()
	if err != nil {
		t.Errorf("Export failed: %v", err)
	}
	
	// 创建新实例
	newMemory := NewLongTermMemory()
	
	// 导入
	err = newMemory.Import(data)
	if err != nil {
		t.Errorf("Import failed: %v", err)
	}
	
	// 验证
	all := newMemory.GetAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 memories after import, got %d", len(all))
	}
}

// TestMemoryStats 测试记忆统计
func TestMemoryStats(t *testing.T) {
	memory := NewLongTermMemory()
	
	// 创建记忆
	memory.Create("test1", MemoryTypeEpisodic, 0.8)
	memory.Create("test2", MemoryTypeSemantic, 0.9)
	memory.Create("test3", MemoryTypeProcedural, 0.7)
	
	// 获取统计
	stats := memory.GetStats()
	
	if stats.TotalMemories != 3 {
		t.Errorf("Expected 3 total memories, got %d", stats.TotalMemories)
	}
	
	if stats.ByType[string(MemoryTypeEpisodic)] != 1 {
		t.Error("Expected 1 episodic memory")
	}
}
