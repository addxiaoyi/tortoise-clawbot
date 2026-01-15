package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	_ "github.com/mattn/go-sqlite3"
)

// ============ PersistentStore 持久化存储 ============

// PersistentStore 支持 SQLite 和 Redis 的持久化存储
type PersistentStore struct {
	db    *sql.DB
	redis *redis.Client
	typ   string // "sqlite" | "redis"
}

// NewPersistentStore 创建持久化存储
func NewPersistentStore(typ, connectionString string) (*PersistentStore, error) {
	store := &PersistentStore{typ: typ}

	switch typ {
	case "sqlite":
		return store.initSQLite(connectionString)
	case "redis":
		return store.initRedis(connectionString)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", typ)
	}
}

// ============ SQLite 实现 ============

func (s *PersistentStore) initSQLite(path string) (*PersistentStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite: %w", err)
	}

	// 创建表
	if err := s.createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	s.db = db
	s.typ = "sqlite"
	log.Printf("✅ SQLite 存储已初始化: %s", path)
	return s, nil
}

func (s *PersistentStore) createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			user_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			message_count INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
		`CREATE TABLE IF NOT EXISTS memories (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			content TEXT NOT NULL,
			importance INTEGER DEFAULT 0,
			tags TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS plugins (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			version TEXT,
			description TEXT,
			author TEXT,
			enabled INTEGER DEFAULT 1,
			status TEXT DEFAULT 'inactive',
			config TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// ============ Redis 实现 ============

func (s *PersistentStore) initRedis(url string) (*PersistentStore, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	s.redis = client
	s.typ = "redis"
	log.Printf("✅ Redis 存储已初始化: %s", url)
	return s, nil
}

// ============ 会话操作 ============

// SessionStore 持久化会话存储
type PersistentSessionStore struct {
	store *PersistentStore
}

// SaveSession 保存会话
func (s *PersistentStore) SaveSession(session *Session) error {
	if s.typ == "sqlite" {
		_, err := s.db.Exec(`
			INSERT OR REPLACE INTO sessions (id, name, user_id, created_at, updated_at, message_count)
			VALUES (?, ?, ?, ?, ?, ?)
		`, session.ID, session.Name, session.UserID, session.CreatedAt, session.UpdatedAt, session.MessageCount)
		return err
	}
	// Redis 实现
	key := fmt.Sprintf("session:%s", session.ID)
	data, _ := json.Marshal(session)
	return s.redis.Set(context.Background(), key, data, 0).Err()
}

// GetSession 获取会话
func (s *PersistentStore) GetSession(id string) (*Session, error) {
	if s.typ == "sqlite" {
		var session Session
		err := s.db.QueryRow(`
			SELECT id, name, user_id, created_at, updated_at, message_count
			FROM sessions WHERE id = ?
		`, id).Scan(&session.ID, &session.Name, &session.UserID, &session.CreatedAt, &session.UpdatedAt, &session.MessageCount)
		if err != nil {
			return nil, err
		}
		return &session, nil
	}
	// Redis 实现
	key := fmt.Sprintf("session:%s", id)
	data, err := s.redis.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, err
	}
	var session Session
	json.Unmarshal(data, &session)
	return &session, nil
}

// ListSessions 列出所有会话
func (s *PersistentStore) ListSessions() ([]*Session, error) {
	if s.typ == "sqlite" {
		rows, err := s.db.Query(`
			SELECT id, name, user_id, created_at, updated_at, message_count
			FROM sessions ORDER BY updated_at DESC
		`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var sessions []*Session
		for rows.Next() {
			var session Session
			rows.Scan(&session.ID, &session.Name, &session.UserID, &session.CreatedAt, &session.UpdatedAt, &session.MessageCount)
			sessions = append(sessions, &session)
		}
		return sessions, nil
	}
	// Redis 实现
	keys, err := s.redis.Keys(context.Background(), "session:*").Result()
	if err != nil {
		return nil, err
	}
	var sessions []*Session
	for _, key := range keys {
		data, _ := s.redis.Get(context.Background(), key).Bytes()
		var session Session
		json.Unmarshal(data, &session)
		sessions = append(sessions, &session)
	}
	return sessions, nil
}

// DeleteSession 删除会话
func (s *PersistentStore) DeleteSession(id string) error {
	if s.typ == "sqlite" {
		_, err := s.db.Exec("DELETE FROM sessions WHERE id = ?", id)
		return err
	}
	// Redis 实现
	key := fmt.Sprintf("session:%s", id)
	return s.redis.Del(context.Background(), key).Err()
}

// ============ 消息操作 ============

// SaveMessage 保存消息
func (s *PersistentStore) SaveMessage(msg *Message) error {
	if s.typ == "sqlite" {
		metadata, _ := json.Marshal(msg.Metadata)
		_, err := s.db.Exec(`
			INSERT OR REPLACE INTO messages (id, session_id, role, content, metadata, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, msg.ID, msg.SessionID, msg.Role, msg.Content, metadata, msg.CreatedAt)
		
		// 更新会话消息计数
		if err == nil {
			s.db.Exec("UPDATE sessions SET message_count = message_count + 1 WHERE id = ?", msg.SessionID)
		}
		return err
	}
	// Redis 实现
	key := fmt.Sprintf("message:%s", msg.ID)
	data, _ := json.Marshal(msg)
	s.redis.Set(context.Background(), key, data, 0)
	s.redis.SAdd(context.Background(), fmt.Sprintf("session:%s:messages", msg.SessionID), msg.ID)
	return nil
}

// GetMessages 获取会话消息
func (s *PersistentStore) GetMessages(sessionID string) ([]*Message, error) {
	if s.typ == "sqlite" {
		rows, err := s.db.Query(`
			SELECT id, session_id, role, content, metadata, created_at
			FROM messages WHERE session_id = ?
			ORDER BY created_at ASC
		`, sessionID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var messages []*Message
		for rows.Next() {
			var msg Message
			var metadata sql.NullString
			rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &metadata, &msg.CreatedAt)
			if metadata.Valid {
				json.Unmarshal([]byte(metadata.String), &msg.Metadata)
			}
			messages = append(messages, &msg)
		}
		return messages, nil
	}
	// Redis 实现
	ids, _ := s.redis.SMembers(context.Background(), fmt.Sprintf("session:%s:messages", sessionID)).Result()
	var messages []*Message
	for _, id := range ids {
		key := fmt.Sprintf("message:%s", id)
		data, _ := s.redis.Get(context.Background(), key).Bytes()
		var msg Message
		json.Unmarshal(data, &msg)
		messages = append(messages, &msg)
	}
	return messages, nil
}

// ============ 记忆操作 ============

// SaveMemory 保存记忆
func (s *PersistentStore) SaveMemory(mem *Memory) error {
	if s.typ == "sqlite" {
		tags, _ := json.Marshal(mem.Tags)
		_, err := s.db.Exec(`
			INSERT OR REPLACE INTO memories (id, type, content, importance, tags, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, mem.ID, mem.Type, mem.Content, mem.Importance, tags, mem.CreatedAt, mem.UpdatedAt)
		return err
	}
	// Redis 实现
	key := fmt.Sprintf("memory:%s", mem.ID)
	data, _ := json.Marshal(mem)
	return s.redis.Set(context.Background(), key, data, 0).Err()
}

// SearchMemories 搜索记忆
func (s *PersistentStore) SearchMemories(query string) ([]*Memory, error) {
	if s.typ == "sqlite" {
		rows, err := s.db.Query(`
			SELECT id, type, content, importance, tags, created_at, updated_at
			FROM memories WHERE content LIKE ?
			ORDER BY importance DESC
		`, "%"+query+"%")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var memories []*Memory
		for rows.Next() {
			var mem Memory
			var tags sql.NullString
			rows.Scan(&mem.ID, &mem.Type, &mem.Content, &mem.Importance, &tags, &mem.CreatedAt, &mem.UpdatedAt)
			if tags.Valid {
				json.Unmarshal([]byte(tags.String), &mem.Tags)
			}
			memories = append(memories, &mem)
		}
		return memories, nil
	}
	// Redis - 简化实现
	return nil, fmt.Errorf("Redis search not implemented")
}

// ============ 配置操作 ============

// SaveConfig 保存配置
func (s *PersistentStore) SaveConfig(key string, value interface{}) error {
	data, _ := json.Marshal(value)
	if s.typ == "sqlite" {
		_, err := s.db.Exec(`
			INSERT OR REPLACE INTO config (key, value, updated_at)
			VALUES (?, ?, ?)
		`, key, string(data), time.Now())
		return err
	}
	// Redis 实现
	return s.redis.Set(context.Background(), fmt.Sprintf("config:%s", key), data, 0).Err()
}

// GetConfig 获取配置
func (s *PersistentStore) GetConfig(key string, dest interface{}) error {
	if s.typ == "sqlite" {
		var value string
		err := s.db.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(value), dest)
	}
	// Redis 实现
	data, err := s.redis.Get(context.Background(), fmt.Sprintf("config:%s", key)).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Close 关闭连接
func (s *PersistentStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	if s.redis != nil {
		return s.redis.Close()
	}
	return nil
}
