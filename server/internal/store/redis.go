package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore Redis 存储适配器
type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

// RedisConfig Redis 配置
type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

// NewRedisStore 创建 Redis 存储
func NewRedisStore(config *RedisConfig) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.URL,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx := context.Background()

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStore{
		client: client,
		ctx:    ctx,
	}, nil
}

// Close 关闭连接
func (s *RedisStore) Close() error {
	return s.client.Close()
}

// ============ Session 持久化 ============

const sessionPrefix = "session:"
const sessionTTL = 30 * 24 * time.Hour // 30 天

// SaveSession 保存会话
func (s *RedisStore) SaveSession(session *Session) error {
	key := sessionPrefix + session.ID
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return s.client.Set(s.ctx, key, data, sessionTTL).Err()
}

// GetSession 获取会话
func (s *RedisStore) GetSession(id string) (*Session, error) {
	key := sessionPrefix + id
	data, err := s.client.Get(s.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteSession 删除会话
func (s *RedisStore) DeleteSession(id string) error {
	key := sessionPrefix + id
	return s.client.Del(s.ctx, key).Err()
}

// GetAllSessions 获取所有会话
func (s *RedisStore) GetAllSessions() ([]*Session, error) {
	keys, err := s.client.Keys(s.ctx, sessionPrefix+"*").Result()
	if err != nil {
		return nil, err
	}

	sessions := make([]*Session, 0, len(keys))
	for _, key := range keys {
		data, err := s.client.Get(s.ctx, key).Bytes()
		if err != nil {
			continue
		}
		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}
		sessions = append(sessions, &session)
	}
	return sessions, nil
}

// ============ Message 持久化 ============

const messagePrefix = "messages:"
const messageTTL = 7 * 24 * time.Hour // 7 天

// SaveMessage 保存消息
func (s *RedisStore) SaveMessage(message *Message) error {
	key := messagePrefix + message.SessionID
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	// 使用有序集合，score 为时间戳
	return s.client.ZAdd(s.ctx, key, redis.Z{
		Score:  float64(message.CreatedAt.Unix()),
		Member: data,
	}).Err()
}

// GetMessages 获取会话消息
func (s *RedisStore) GetMessages(sessionID string, limit int) ([]*Message, error) {
	key := messagePrefix + sessionID
	
	// 按时间倒序获取
	results, err := s.client.ZRevRange(s.ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	messages := make([]*Message, 0, len(results))
	for _, data := range results {
		var msg Message
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			continue
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}

// DeleteMessages 删除会话所有消息
func (s *RedisStore) DeleteMessages(sessionID string) error {
	key := messagePrefix + sessionID
	return s.client.Del(s.ctx, key).Err()
}

// ============ Memory 持久化 ============

const memoryPrefix = "memory:"
const memoryTTL = 365 * 24 * time.Hour // 1 年

// SaveMemory 保存记忆
func (s *RedisStore) SaveMemory(memory *Memory) error {
	key := memoryPrefix + memory.ID
	data, err := json.Marshal(memory)
	if err != nil {
		return err
	}
	return s.client.Set(s.ctx, key, data, memoryTTL).Err()
}

// GetMemory 获取记忆
func (s *RedisStore) GetMemory(id string) (*Memory, error) {
	key := memoryPrefix + id
	data, err := s.client.Get(s.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var memory Memory
	if err := json.Unmarshal(data, &memory); err != nil {
		return nil, err
	}
	return &memory, nil
}

// DeleteMemory 删除记忆
func (s *RedisStore) DeleteMemory(id string) error {
	key := memoryPrefix + id
	return s.client.Del(s.ctx, key).Err()
}

// GetAllMemories 获取所有记忆
func (s *RedisStore) GetAllMemories() ([]*Memory, error) {
	keys, err := s.client.Keys(s.ctx, memoryPrefix+"*").Result()
	if err != nil {
		return nil, err
	}

	memorys := make([]*Memory, 0, len(keys))
	for _, key := range keys {
		data, err := s.client.Get(s.ctx, key).Bytes()
		if err != nil {
			continue
		}
		var memory Memory
		if err := json.Unmarshal(data, &memory); err != nil {
			continue
		}
		memorys = append(memorys, &memory)
	}
	return memorys, nil
}

// ============ Config 持久化 ============

const configKey = "config:app"

// SaveConfig 保存配置
func (s *RedisStore) SaveConfig(config *Config) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	// 配置永不过期
	return s.client.Set(s.ctx, configKey, data, 0).Err()
}

// GetConfig 获取配置
func (s *RedisStore) GetConfig() (*Config, error) {
	data, err := s.client.Get(s.ctx, configKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// ============ 通用操作 ============

// Ping 测试连接
func (s *RedisStore) Ping() error {
	return s.client.Ping(s.ctx).Err()
}

// FlushDB 清空数据库
func (s *RedisStore) FlushDB() error {
	return s.client.FlushDB(s.ctx).Err()
}
