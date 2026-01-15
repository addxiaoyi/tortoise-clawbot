package database

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	db, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// 测试数据库连接
	if db.db == nil {
		t.Fatal("Database connection is nil")
	}
}

func TestSession_CRUD(t *testing.T) {
	db, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// 创建会话
	session := &Session{
		ID:         "test-session-1",
		UserID:     "test-user",
		Title:      "Test Session",
		AIProvider: "openai",
		Model:      "gpt-4",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := db.CreateSession(session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 获取会话
	retrieved, err := db.GetSession("test-session-1")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrieved.Title != "Test Session" {
		t.Errorf("Expected title 'Test Session', got '%s'", retrieved.Title)
	}

	// 列出会话
	sessions, err := db.ListSessions("test-user", 10, 0)
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessions))
	}

	// 删除会话
	if err := db.DeleteSession("test-session-1"); err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}
}

func TestMessage_CRUD(t *testing.T) {
	db, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// 先创建会话
	session := &Session{
		ID:         "test-session-msg",
		UserID:     "test-user",
		Title:      "Message Test",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	db.CreateSession(session)

	// 创建消息
	message := &Message{
		ID:        "test-message-1",
		SessionID: "test-session-msg",
		Role:      "user",
		Content:   "Hello, world!",
		CreatedAt: time.Now(),
	}

	if err := db.CreateMessage(message); err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	// 获取消息
	retrieved, err := db.GetMessage("test-message-1")
	if err != nil {
		t.Fatalf("Failed to get message: %v", err)
	}

	if retrieved.Content != "Hello, world!" {
		t.Errorf("Expected content 'Hello, world!', got '%s'", retrieved.Content)
	}

	// 列出消息
	messages, err := db.ListMessages("test-session-msg", 10, 0)
	if err != nil {
		t.Fatalf("Failed to list messages: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	// 删除消息
	if err := db.DeleteMessage("test-message-1"); err != nil {
		t.Fatalf("Failed to delete message: %v", err)
	}
}

func TestMessage_Metadata(t *testing.T) {
	msg := &Message{
		ID:        "test-msg-meta",
		SessionID: "test-session",
		Role:      "assistant",
		Content:   "Response content",
		CreatedAt: time.Now(),
	}

	// 设置元数据
	msg.SetMetadata(map[string]interface{}{
		"tokens":     100,
		"model":      "gpt-4",
		"finishReason": "stop",
	})

	// 获取元数据
	meta := msg.GetMetadata()
	if meta["tokens"] != 100 {
		t.Errorf("Expected tokens 100, got %v", meta["tokens"])
	}
}

func TestMemory_CRUD(t *testing.T) {
	db, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// 创建记忆
	memory := &Memory{
		ID:        "test-memory-1",
		UserID:    "test-user",
		Title:     "Test Memory",
		Content:   "Important information",
		Type:      "fact",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.CreateMemory(memory); err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// 获取记忆
	retrieved, err := db.GetMemory("test-memory-1")
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}

	if retrieved.Title != "Test Memory" {
		t.Errorf("Expected title 'Test Memory', got '%s'", retrieved.Title)
	}

	// 列出记忆
	memories, err := db.ListMemories("test-user", 10, 0)
	if err != nil {
		t.Fatalf("Failed to list memories: %v", err)
	}

	if len(memories) != 1 {
		t.Errorf("Expected 1 memory, got %d", len(memories))
	}

	// 删除记忆
	if err := db.DeleteMemory("test-memory-1"); err != nil {
		t.Fatalf("Failed to delete memory: %v", err)
	}
}

func TestMemory_Tags(t *testing.T) {
	m := &Memory{
		ID:        "test-memory-tags",
		UserID:    "test-user",
		Title:     "Tags Test",
		Content:   "Content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 设置标签
	m.SetTags([]string{"important", "work", "project"})

	// 获取标签
	tags := m.GetTags()
	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}

	expectedTags := map[string]bool{
		"important": true,
		"work":      true,
		"project":   true,
	}

	for _, tag := range tags {
		if !expectedTags[tag] {
			t.Errorf("Unexpected tag: %s", tag)
		}
	}
}

func TestChannel_CRUD(t *testing.T) {
	db, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// 创建渠道
	channel := &Channel{
		ID:        "test-channel-1",
		UserID:    "test-user",
		Type:      "telegram",
		Name:      "Test Channel",
		Status:    "disconnected",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.CreateChannel(channel); err != nil {
		t.Fatalf("Failed to create channel: %v", err)
	}

	// 获取渠道
	retrieved, err := db.GetChannel("test-channel-1")
	if err != nil {
		t.Fatalf("Failed to get channel: %v", err)
	}

	if retrieved.Type != "telegram" {
		t.Errorf("Expected type 'telegram', got '%s'", retrieved.Type)
	}

	// 列出渠道
	channels, err := db.ListChannels("test-user")
	if err != nil {
		t.Fatalf("Failed to list channels: %v", err)
	}

	if len(channels) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(channels))
	}

	// 删除渠道
	if err := db.DeleteChannel("test-channel-1"); err != nil {
		t.Fatalf("Failed to delete channel: %v", err)
	}
}

func TestConfig_CRUD(t *testing.T) {
	db, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// 设置配置
	if err := db.SetConfig("test-user", "theme", "dark"); err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	// 获取配置
	value, err := db.GetConfig("test-user", "theme")
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	if value != "dark" {
		t.Errorf("Expected value 'dark', got '%s'", value)
	}
}

func TestChannel_Config(t *testing.T) {
	ch := &Channel{
		ID:        "test-channel-config",
		UserID:    "test-user",
		Type:      "discord",
		Name:      "Discord Bot",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 设置配置
	ch.SetChannelConfig(map[string]interface{}{
		"token":     "secret-token",
		"guildId":   "123456",
		"channelId": "789012",
	})

	// 获取配置
	config := ch.GetChannelConfig()
	if config["guildId"] != "123456" {
		t.Errorf("Expected guildId '123456', got '%v'", config["guildId"])
	}
}
