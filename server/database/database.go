package database

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB 数据库接口
type DB interface {
	Close() error

	// 会话操作
	CreateSession(session *Session) error
	GetSession(id string) (*Session, error)
	ListSessions(userID string, limit, offset int) ([]*Session, error)
	UpdateSession(session *Session) error
	DeleteSession(id string) error

	// 消息操作
	CreateMessage(message *Message) error
	GetMessage(id string) (*Message, error)
	ListMessages(sessionID string, limit, offset int) ([]*Message, error)
	DeleteMessage(id string) error

	// 记忆操作
	CreateMemory(memory *Memory) error
	GetMemory(id string) (*Memory, error)
	ListMemories(userID string, limit, offset int) ([]*Memory, error)
	SearchMemories(userID, query string, limit int) ([]*Memory, error)
	UpdateMemory(memory *Memory) error
	DeleteMemory(id string) error

	// 渠道操作
	CreateChannel(channel *Channel) error
	GetChannel(id string) (*Channel, error)
	ListChannels(userID string) ([]*Channel, error)
	UpdateChannel(channel *Channel) error
	DeleteChannel(id string) error

	// 配置操作
	GetConfig(userID, key string) (string, error)
	SetConfig(userID, key, value string) error
}

// SQLiteDB SQLite 数据库实现
type SQLiteDB struct {
	db *sql.DB
}

// Session 会话
type Session struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Title      string    `json:"title"`
	AIProvider string    `json:"ai_provider"`
	Model      string    `json:"model"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Message 消息 (DAG 结构支持 parentId)
type Message struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"session_id"`
	ParentID   string    `json:"parent_id,omitempty"` // 父消息 ID，用于 DAG 结构
	Role       string    `json:"role"`               // user, assistant, system
	Content    string    `json:"content"`
	Model      string    `json:"model,omitempty"`
	Tokens     int       `json:"tokens"`
	Metadata   string    `json:"metadata"` // JSON 字符串
	CreatedAt  time.Time `json:"created_at"`
}

// Session 会话 (支持分支)
type Session struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	ParentID    string    `json:"parent_id,omitempty"` // 父会话 ID，用于分支
	Title       string    `json:"title"`
	AIProvider  string    `json:"ai_provider"`
	Model       string    `json:"model"`
	IsBranch    bool      `json:"is_branch"`
	BranchCount int       `json:"branch_count"` // 分支数量
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Memory 记忆
type Memory struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Tags      string    `json:"tags"` // JSON 数组字符串
	Embedding []float64 `json:"-"`    // 向量，不存储在数据库
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Channel 渠道
type Channel struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Config    string    `json:"config"` // JSON 字符串
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// New 创建数据库连接
func New(cfg interface{}) (*SQLiteDB, error) {
	path := "./data/tortoise.db"
	
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, err
	}

	sqlDB := &SQLiteDB{db: db}
	
	// 初始化表
	if err := sqlDB.initTables(); err != nil {
		return nil, err
	}

	return sqlDB, nil
}

// initTables 初始化表
func (db *SQLiteDB) initTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		ai_provider TEXT,
		model TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		model TEXT,
		tokens INTEGER DEFAULT 0,
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS memories (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		type TEXT DEFAULT 'general',
		tags TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS channels (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		config TEXT,
		status TEXT DEFAULT 'disconnected',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT,
		UNIQUE(user_id, key)
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
	CREATE INDEX IF NOT EXISTS idx_memories_user_id ON memories(user_id);
	CREATE INDEX IF NOT EXISTS idx_channels_user_id ON channels(user_id);
	`

	_, err := db.db.Exec(schema)
	return err
}

// Close 关闭数据库
func (db *SQLiteDB) Close() error {
	return db.db.Close()
}

// ========== 会话操作 ==========

func (db *SQLiteDB) CreateSession(session *Session) error {
	_, err := db.db.Exec(
		`INSERT INTO sessions (id, user_id, title, ai_provider, model, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		session.ID, session.UserID, session.Title, session.AIProvider, session.Model,
		session.CreatedAt, session.UpdatedAt,
	)
	return err
}

func (db *SQLiteDB) GetSession(id string) (*Session, error) {
	s := &Session{}
	err := db.db.QueryRow(
		`SELECT id, user_id, title, ai_provider, model, created_at, updated_at 
		 FROM sessions WHERE id = ?`, id,
	).Scan(&s.ID, &s.UserID, &s.Title, &s.AIProvider, &s.Model, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (db *SQLiteDB) ListSessions(userID string, limit, offset int) ([]*Session, error) {
	rows, err := db.db.Query(
		`SELECT id, user_id, title, ai_provider, model, created_at, updated_at 
		 FROM sessions WHERE user_id = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		s := &Session{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.Title, &s.AIProvider, &s.Model, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (db *SQLiteDB) UpdateSession(session *Session) error {
	_, err := db.db.Exec(
		`UPDATE sessions SET title = ?, ai_provider = ?, model = ?, updated_at = ? WHERE id = ?`,
		session.Title, session.AIProvider, session.Model, session.UpdatedAt, session.ID,
	)
	return err
}

func (db *SQLiteDB) DeleteSession(id string) error {
	_, err := db.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// ========== 消息操作 ==========

func (db *SQLiteDB) CreateMessage(message *Message) error {
	_, err := db.db.Exec(
		`INSERT INTO messages (id, session_id, role, content, model, tokens, metadata, created_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		message.ID, message.SessionID, message.Role, message.Content, message.Model,
		message.Tokens, message.Metadata, message.CreatedAt,
	)
	return err
}

func (db *SQLiteDB) GetMessage(id string) (*Message, error) {
	m := &Message{}
	err := db.db.QueryRow(
		`SELECT id, session_id, role, content, model, tokens, metadata, created_at 
		 FROM messages WHERE id = ?`, id,
	).Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.Model, &m.Tokens, &m.Metadata, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (db *SQLiteDB) ListMessages(sessionID string, limit, offset int) ([]*Message, error) {
	rows, err := db.db.Query(
		`SELECT id, session_id, role, content, model, tokens, metadata, created_at 
		 FROM messages WHERE session_id = ? ORDER BY created_at ASC LIMIT ? OFFSET ?`,
		sessionID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		m := &Message{}
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.Model, &m.Tokens, &m.Metadata, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (db *SQLiteDB) DeleteMessage(id string) error {
	_, err := db.db.Exec(`DELETE FROM messages WHERE id = ?`, id)
	return err
}

// ========== 记忆操作 ==========

func (db *SQLiteDB) CreateMemory(memory *Memory) error {
	_, err := db.db.Exec(
		`INSERT INTO memories (id, user_id, title, content, type, tags, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		memory.ID, memory.UserID, memory.Title, memory.Content, memory.Type,
		memory.Tags, memory.CreatedAt, memory.UpdatedAt,
	)
	return err
}

func (db *SQLiteDB) GetMemory(id string) (*Memory, error) {
	m := &Memory{}
	err := db.db.QueryRow(
		`SELECT id, user_id, title, content, type, tags, created_at, updated_at 
		 FROM memories WHERE id = ?`, id,
	).Scan(&m.ID, &m.UserID, &m.Title, &m.Content, &m.Type, &m.Tags, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (db *SQLiteDB) ListMemories(userID string, limit, offset int) ([]*Memory, error) {
	rows, err := db.db.Query(
		`SELECT id, user_id, title, content, type, tags, created_at, updated_at 
		 FROM memories WHERE user_id = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		m := &Memory{}
		if err := rows.Scan(&m.ID, &m.UserID, &m.Title, &m.Content, &m.Type, &m.Tags, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		memories = append(memories, m)
	}
	return memories, nil
}

func (db *SQLiteDB) SearchMemories(userID, query string, limit int) ([]*Memory, error) {
	rows, err := db.db.Query(
		`SELECT id, user_id, title, content, type, tags, created_at, updated_at 
		 FROM memories WHERE user_id = ? AND (title LIKE ? OR content LIKE ?) 
		 ORDER BY updated_at DESC LIMIT ?`,
		userID, "%"+query+"%", "%"+query+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		m := &Memory{}
		if err := rows.Scan(&m.ID, &m.UserID, &m.Title, &m.Content, &m.Type, &m.Tags, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		memories = append(memories, m)
	}
	return memories, nil
}

func (db *SQLiteDB) UpdateMemory(memory *Memory) error {
	_, err := db.db.Exec(
		`UPDATE memories SET title = ?, content = ?, type = ?, tags = ?, updated_at = ? WHERE id = ?`,
		memory.Title, memory.Content, memory.Type, memory.Tags, memory.UpdatedAt, memory.ID,
	)
	return err
}

func (db *SQLiteDB) DeleteMemory(id string) error {
	_, err := db.db.Exec(`DELETE FROM memories WHERE id = ?`, id)
	return err
}

// ========== 渠道操作 ==========

func (db *SQLiteDB) CreateChannel(channel *Channel) error {
	_, err := db.db.Exec(
		`INSERT INTO channels (id, user_id, type, name, config, status, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		channel.ID, channel.UserID, channel.Type, channel.Name, channel.Config,
		channel.Status, channel.CreatedAt, channel.UpdatedAt,
	)
	return err
}

func (db *SQLiteDB) GetChannel(id string) (*Channel, error) {
	c := &Channel{}
	err := db.db.QueryRow(
		`SELECT id, user_id, type, name, config, status, created_at, updated_at 
		 FROM channels WHERE id = ?`, id,
	).Scan(&c.ID, &c.UserID, &c.Type, &c.Name, &c.Config, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (db *SQLiteDB) ListChannels(userID string) ([]*Channel, error) {
	rows, err := db.db.Query(
		`SELECT id, user_id, type, name, config, status, created_at, updated_at 
		 FROM channels WHERE user_id = ? ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		c := &Channel{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.Type, &c.Name, &c.Config, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, c)
	}
	return channels, nil
}

func (db *SQLiteDB) UpdateChannel(channel *Channel) error {
	_, err := db.db.Exec(
		`UPDATE channels SET name = ?, config = ?, status = ?, updated_at = ? WHERE id = ?`,
		channel.Name, channel.Config, channel.Status, channel.UpdatedAt, channel.ID,
	)
	return err
}

func (db *SQLiteDB) DeleteChannel(id string) error {
	_, err := db.db.Exec(`DELETE FROM channels WHERE id = ?`, id)
	return err
}

// ========== 配置操作 ==========

func (db *SQLiteDB) GetConfig(userID, key string) (string, error) {
	var value string
	err := db.db.QueryRow(
		`SELECT value FROM configs WHERE user_id = ? AND key = ?`, userID, key,
	).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (db *SQLiteDB) SetConfig(userID, key, value string) error {
	_, err := db.db.Exec(
		`INSERT INTO configs (user_id, key, value) VALUES (?, ?, ?)
		 ON CONFLICT(user_id, key) DO UPDATE SET value = ?`,
		userID, key, value, value,
	)
	return err
}

// ========== 辅助方法 ==========

// GetTags 获取标签列表
func (m *Memory) GetTags() []string {
	if m.Tags == "" {
		return []string{}
	}
	var tags []string
	json.Unmarshal([]byte(m.Tags), &tags)
	return tags
}

// SetTags 设置标签列表
func (m *Memory) SetTags(tags []string) {
	data, _ := json.Marshal(tags)
	m.Tags = string(data)
}

// GetMetadata 获取元数据
func (m *Message) GetMetadata() map[string]interface{} {
	if m.Metadata == "" {
		return map[string]interface{}{}
	}
	var meta map[string]interface{}
	json.Unmarshal([]byte(m.Metadata), &meta)
	return meta
}

// SetMetadata 设置元数据
func (m *Message) SetMetadata(meta map[string]interface{}) {
	data, _ := json.Marshal(meta)
	m.Metadata = string(data)
}

// GetChannelConfig 获取渠道配置
func (c *Channel) GetChannelConfig() map[string]interface{} {
	if c.Config == "" {
		return map[string]interface{}{}
	}
	var config map[string]interface{}
	json.Unmarshal([]byte(c.Config), &config)
	return config
}

// SetChannelConfig 设置渠道配置
func (c *Channel) SetChannelConfig(config map[string]interface{}) {
	data, _ := json.Marshal(config)
	c.Config = string(data)
}
