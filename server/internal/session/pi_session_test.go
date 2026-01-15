package session

import (
	"testing"
	"time"
)

// TestPiSession 测试 Pi 会话
func TestPiSession(t *testing.T) {
	manager := NewPiSessionManager()
	
	// 创建会话
	session := manager.CreateSession("test-session", "gpt-4")
	if session == nil {
		t.Fatal("Expected session, got nil")
	}
	
	if session.ID == "" {
		t.Error("Expected non-empty session ID")
	}
	
	if session.Title != "test-session" {
		t.Errorf("Expected title 'test-session', got '%s'", session.Title)
	}
}

// TestPiMessageChain 测试消息链
func TestPiMessageChain(t *testing.T) {
	manager := NewPiSessionManager()
	session := manager.CreateSession("test", "gpt-4")
	
	// 添加父消息
	parent := session.AddMessage("user", "Hello", nil)
	
	// 添加子消息
	child := session.AddMessage("assistant", "Hi there!", parent.ID)
	
	if child.ParentID != parent.ID {
		t.Errorf("Expected parent ID '%s', got '%s'", parent.ID, child.ParentID)
	}
}

// TestPiSessionExport 测试会话导出
func TestPiSessionExport(t *testing.T) {
	manager := NewPiSessionManager()
	session := manager.CreateSession("test", "gpt-4")
	
	// 添加消息
	session.AddMessage("user", "Test message", nil)
	session.AddMessage("assistant", "Test response", nil)
	
	// 导出为 JSON
	jsonData, err := session.ExportJSON()
	if err != nil {
		t.Errorf("ExportJSON failed: %v", err)
	}
	
	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON export")
	}
	
	// 导出为 Markdown
	mdData, err := session.ExportMarkdown()
	if err != nil {
		t.Errorf("ExportMarkdown failed: %v", err)
	}
	
	if len(mdData) == 0 {
		t.Error("Expected non-empty Markdown export")
	}
}

// TestPiSessionCompression 测试会话压缩
func TestPiSessionCompression(t *testing.T) {
	manager := NewPiSessionManager()
	session := manager.CreateSession("test", "gpt-4")
	
	// 添加多条消息
	for i := 0; i < 10; i++ {
		session.AddMessage("user", "Test message", nil)
		session.AddMessage("assistant", "Test response", nil)
	}
	
	// 获取压缩大小
	compressedSize := session.GetCompressedSize()
	if compressedSize == 0 {
		t.Error("Expected non-zero compressed size")
	}
}

// TestPiSessionManager 测试会话管理器
func TestPiSessionManager(t *testing.T) {
	manager := NewPiSessionManager()
	
	// 创建多个会话
	session1 := manager.CreateSession("session1", "gpt-4")
	session2 := manager.CreateSession("session2", "claude-3")
	
	// 列出所有会话
	sessions := manager.ListSessions()
	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}
	
	// 获取单个会话
	retrieved := manager.GetSession(session1.ID)
	if retrieved == nil {
		t.Error("Expected to retrieve session")
	}
	
	// 删除会话
	err := manager.DeleteSession(session2.ID)
	if err != nil {
		t.Errorf("DeleteSession failed: %v", err)
	}
	
	// 验证删除
	sessions = manager.ListSessions()
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session after delete, got %d", len(sessions))
	}
}
