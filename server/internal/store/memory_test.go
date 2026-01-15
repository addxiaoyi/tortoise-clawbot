package store

import (
	"testing"
	"time"
)

// ============ MemoryStore Tests ============

func TestMemoryStore_Create(t *testing.T) {
	store := NewMemoryStore()
	
	mem := store.Create("semantic", "测试记忆内容", 5, []string{"测试", "单元测试"})
	
	if mem == nil {
		t.Fatal("Create returned nil")
	}
	
	if mem.ID == "" {
		t.Error("Memory ID should not be empty")
	}
	
	if mem.Type != "semantic" {
		t.Errorf("Expected type 'semantic', got '%s'", mem.Type)
	}
	
	if mem.Content != "测试记忆内容" {
		t.Errorf("Expected content '测试记忆内容', got '%s'", mem.Content)
	}
	
	if mem.Importance != 5 {
		t.Errorf("Expected importance 5, got %d", mem.Importance)
	}
}

func TestMemoryStore_List(t *testing.T) {
	store := NewMemoryStore()
	
	// 创建多个记忆
	store.Create("working", "工作记忆1", 3, []string{"工作"})
	store.Create("semantic", "语义记忆1", 5, []string{"语义"})
	store.Create("episodic", "情景记忆1", 7, []string{"情景"})
	
	memories := store.List()
	
	if len(memories) != 3 {
		t.Errorf("Expected 3 memories, got %d", len(memories))
	}
}

func TestMemoryStore_Get(t *testing.T) {
	store := NewMemoryStore()
	
	mem := store.Create("semantic", "测试记忆", 5, []string{"测试"})
	
	// 测试获取存在的记忆
	found := store.Get(mem.ID)
	if found == nil {
		t.Fatal("Get returned nil for existing memory")
	}
	if found.Content != "测试记忆" {
		t.Errorf("Expected content '测试记忆', got '%s'", found.Content)
	}
	
	// 测试获取不存在的记忆
	notFound := store.Get("non-existent-id")
	if notFound != nil {
		t.Error("Get should return nil for non-existent memory")
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	store := NewMemoryStore()
	
	mem := store.Create("semantic", "测试记忆", 5, nil)
	
	// 删除存在的记忆
	deleted := store.Delete(mem.ID)
	if !deleted {
		t.Error("Delete should return true for existing memory")
	}
	
	// 验证已删除
	if store.Get(mem.ID) != nil {
		t.Error("Memory should not exist after delete")
	}
	
	// 删除不存在的记忆
	deleted = store.Delete("non-existent-id")
	if deleted {
		t.Error("Delete should return false for non-existent memory")
	}
}

func TestMemoryStore_Search(t *testing.T) {
	store := NewMemoryStore()
	
	store.Create("semantic", "这是一个关于狗的记忆", 5, []string{"动物", "狗"})
	store.Create("semantic", "这是一个关于猫的记忆", 5, []string{"动物", "猫"})
	store.Create("working", "工作内容", 3, []string{"工作"})
	
	// 搜索内容
	results := store.Search("狗")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for '狗', got %d", len(results))
	}
	
	// 搜索标签
	results = store.Search("动物")
	if len(results) != 2 {
		t.Errorf("Expected 2 results for '动物', got %d", len(results))
	}
}

func TestMemoryStore_Update(t *testing.T) {
	store := NewMemoryStore()
	
	mem := store.Create("semantic", "原始内容", 3, []string{"标签1"})
	
	// 更新记忆
	updated := store.Update(mem.ID, "更新内容", 8, []string{"新标签"})
	
	if updated == nil {
		t.Fatal("Update returned nil")
	}
	
	if updated.Content != "更新内容" {
		t.Errorf("Expected content '更新内容', got '%s'", updated.Content)
	}
	
	if updated.Importance != 8 {
		t.Errorf("Expected importance 8, got %d", updated.Importance)
	}
}

func TestMemoryStore_Count(t *testing.T) {
	store := NewMemoryStore()
	
	if store.Count() != 0 {
		t.Errorf("Expected 0 initially, got %d", store.Count())
	}
	
	store.Create("semantic", "记忆1", 5, nil)
	store.Create("working", "记忆2", 3, nil)
	store.Create("episodic", "记忆3", 7, nil)
	
	if store.Count() != 3 {
		t.Errorf("Expected 3 after adding 3 memories, got %d", store.Count())
	}
}

// ============ SessionStore Tests ============

func TestSessionStore_Create(t *testing.T) {
	store := NewSessionStore()
	
	session := store.Create("测试会话", "user123")
	
	if session == nil {
		t.Fatal("Create returned nil")
	}
	
	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}
	
	if session.Name != "测试会话" {
		t.Errorf("Expected name '测试会话', got '%s'", session.Name)
	}
	
	if session.UserID != "user123" {
		t.Errorf("Expected userId 'user123', got '%s'", session.UserID)
	}
}

func TestSessionStore_List(t *testing.T) {
	store := NewSessionStore()
	
	store.Create("会话1", "user1")
	store.Create("会话2", "user2")
	store.Create("会话3", "user1")
	
	sessions := store.List()
	
	if len(sessions) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(sessions))
	}
}

func TestSessionStore_Delete(t *testing.T) {
	store := NewSessionStore()
	
	session := store.Create("测试会话", "user1")
	
	deleted := store.Delete(session.ID)
	if !deleted {
		t.Error("Delete should return true")
	}
	
	if store.Get(session.ID) != nil {
		t.Error("Session should not exist after delete")
	}
}

// ============ MessageStore Tests ============

func TestMessageStore_Create(t *testing.T) {
	store := NewMessageStore()
	
	msg := &Message{
		SessionID: "session1",
		Role:     "user",
		Content:  "测试消息",
	}
	
	created := store.Create(msg)
	
	if created.ID == "" {
		t.Error("Message ID should be generated")
	}
	
	if created.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestMessageStore_GetBySession(t *testing.T) {
	store := NewMessageStore()
	
	sessionID := "session123"
	
	store.Create(&Message{
		SessionID: sessionID,
		Role:     "user",
		Content:  "消息1",
	})
	store.Create(&Message{
		SessionID: sessionID,
		Role:     "assistant",
		Content:  "回复1",
	})
	store.Create(&Message{
		SessionID: "other-session",
		Role:     "user",
		Content:  "其他消息",
	})
	
	messages := store.GetBySession(sessionID)
	
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages for session %s, got %d", sessionID, len(messages))
	}
}

// ============ PluginStore Tests ============

func TestPluginStore_List(t *testing.T) {
	store := NewPluginStore()
	
	plugins := store.List()
	
	// 应该包含默认插件
	if len(plugins) < 2 {
		t.Errorf("Expected at least 2 default plugins, got %d", len(plugins))
	}
	
	// 查找 search 插件
	found := false
	for _, p := range plugins {
		if p.ID == "search" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("search plugin should be in default plugins")
	}
}

func TestPluginStore_Toggle(t *testing.T) {
	store := NewPluginStore()
	
	plugin := store.Get("search")
	if plugin == nil {
		t.Fatal("search plugin not found")
	}
	
	initialState := plugin.Enabled
	
	store.Toggle("search")
	
	plugin = store.Get("search")
	if plugin.Enabled == initialState {
		t.Error("Plugin state should have changed after toggle")
	}
}

// ============ ConfigStore Tests ============

func TestConfigStore_GetConfig(t *testing.T) {
	store := NewConfigStore()
	
	cfg := store.GetConfig()
	
	if cfg == nil {
		t.Fatal("GetConfig returned nil")
	}
	
	if len(cfg.AI.Providers) == 0 {
		t.Error("Should have default AI providers")
	}
	
	// 检查敏感字段是否被过滤
	if cfg.Security.JWTSecret != "" {
		t.Error("JWT Secret should be masked")
	}
}

func TestConfigStore_UpdateConfig(t *testing.T) {
	store := NewConfigStore()
	
	updates := map[string]interface{}{
		"ai": map[string]interface{}{
			"routing": "cost",
		},
	}
	
	err := store.UpdateConfig(updates)
	if err != nil {
		t.Errorf("UpdateConfig failed: %v", err)
	}
	
	cfg := store.GetConfig()
	if cfg.AI.Routing != "cost" {
		t.Errorf("Expected routing 'cost', got '%s'", cfg.AI.Routing)
	}
}

// ============ Integration Tests ============

func TestMemory_CRUD(t *testing.T) {
	store := NewMemoryStore()
	
	// Create
	mem := store.Create("semantic", "测试", 5, []string{"测试"})
	id := mem.ID
	
	// Read
	if store.Get(id) == nil {
		t.Fatal("Failed to read created memory")
	}
	
	// Update
	updated := store.Update(id, "更新内容", 8, []string{"新标签"})
	if updated == nil {
		t.Fatal("Failed to update memory")
	}
	
	// Delete
	if !store.Delete(id) {
		t.Fatal("Failed to delete memory")
	}
	
	// Verify deleted
	if store.Get(id) != nil {
		t.Fatal("Memory should not exist after delete")
	}
}

func TestSession_CRUD(t *testing.T) {
	store := NewSessionStore()
	
	// Create
	session := store.Create("测试会话", "user1")
	id := session.ID
	
	// Read
	if store.Get(id) == nil {
		t.Fatal("Failed to read created session")
	}
	
	// Update
	store.Update(id)
	
	// Delete
	if !store.Delete(id) {
		t.Fatal("Failed to delete session")
	}
	
	// Verify deleted
	if store.Get(id) != nil {
		t.Fatal("Session should not exist after delete")
	}
}

// ============ Benchmark Tests ============

func BenchmarkMemoryStore_Create(b *testing.B) {
	store := NewMemoryStore()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Create("semantic", "测试内容", 5, []string{"标签"})
	}
}

func BenchmarkMemoryStore_Search(b *testing.B) {
	store := NewMemoryStore()
	
	// 添加测试数据
	for i := 0; i < 100; i++ {
		store.Create("semantic", "这是测试内容"+string(rune('0'+i)), 5, []string{"标签"})
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Search("测试")
	}
}

func BenchmarkSessionStore_List(b *testing.B) {
	store := NewSessionStore()
	
	// 添加测试数据
	for i := 0; i < 100; i++ {
		store.Create("会话", "user1")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.List()
	}
}
