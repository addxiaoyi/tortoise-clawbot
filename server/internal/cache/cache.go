package cache

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"
)

// ============ LRU Cache LRU 缓存 ============

// Cache LRU 缓存接口
type Cache interface {
	// 设置值
	Set(key string, value interface{}) error
	
	// 获取值
	Get(key string) (interface{}, bool)
	
	// 删除
	Delete(key string) error
	
	// 清除
	Clear() error
	
	// 获取大小
	Size() int
	
	// 获取最大容量
	Capacity() int
}

// LRUCache LRU 缓存实现
type LRUCache struct {
	mu       sync.RWMutex
	capacity int
	ttl      time.Duration
	items    map[string]*list.Element
	eviction *list.List
}

// entry 缓存条目
type entry struct {
	key       string
	value     interface{}
	expiresAt time.Time
}

// NewLRUCache 创建 LRU 缓存
func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[string]*list.Element),
		eviction: list.New(),
	}
}

// Set 设置值
func (c *LRUCache) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否已存在
	if elem, exists := c.items[key]; exists {
		// 更新值
		ent := elem.Value.(*entry)
		ent.value = value
		if c.ttl > 0 {
			ent.expiresAt = time.Now().Add(c.ttl)
		} else {
			ent.expiresAt = time.Time{}
		}
		// 移到列表头部
		c.eviction.MoveToFront(elem)
		return nil
	}

	// 检查容量
	if c.eviction.Len() >= c.capacity {
		// 驱逐最旧的
		c.evictOldest()
	}

	// 创建新条目
	ent := &entry{
		key:   key,
		value: value,
	}
	if c.ttl > 0 {
		ent.expiresAt = time.Now().Add(c.ttl)
	}

	elem := c.eviction.PushFront(ent)
	c.items[key] = elem

	return nil
}

// Get 获取值
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		return nil, false
	}

	ent := elem.Value.(*entry)

	// 检查过期
	if c.ttl > 0 && time.Now().After(ent.expiresAt) {
		c.eviction.Remove(elem)
		delete(c.items, key)
		return nil, false
	}

	// 移到列表头部
	c.eviction.MoveToFront(elem)

	return ent.value, true
}

// Delete 删除
func (c *LRUCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		return nil
	}

	c.eviction.Remove(elem)
	delete(c.items, key)

	return nil
}

// Clear 清除
func (c *LRUCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.eviction.Init()

	return nil
}

// Size 获取大小
func (c *LRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.eviction.Len()
}

// Capacity 获取容量
func (c *LRUCache) Capacity() int {
	return c.capacity
}

// evictOldest 驱逐最旧的
func (c *LRUCache) evictOldest() {
	elem := c.eviction.Back()
	if elem != nil {
		ent := elem.Value.(*entry)
		delete(c.items, ent.key)
		c.eviction.Remove(elem)
	}
}

// ============ Thread-Safe Map Cache ============

// MapCache 线程安全的 Map 缓存
type MapCache struct {
	mu    sync.RWMutex
	data  map[string]interface{}
	ttl   time.Duration
	expire map[string]time.Time
}

// NewMapCache 创建 Map 缓存
func NewMapCache(ttl time.Duration) *MapCache {
	return &MapCache{
		data:    make(map[string]interface{}),
		ttl:     ttl,
		expire:  make(map[string]time.Time),
	}
}

// Set 设置值
func (c *MapCache) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = value
	if c.ttl > 0 {
		c.expire[key] = time.Now().Add(c.ttl)
	}

	return nil
}

// Get 获取值
func (c *MapCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// 检查过期
	if c.ttl > 0 {
		if exp, ok := c.expire[key]; ok && time.Now().After(exp) {
			return nil, false
		}
	}

	return value, true
}

// Delete 删除
func (c *MapCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	delete(c.expire, key)

	return nil
}

// Clear 清除
func (c *MapCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]interface{})
	c.expire = make(map[string]time.Time)

	return nil
}

// Size 获取大小
func (c *MapCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// Keys 获取所有键
func (c *MapCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}

// ============ Cache Manager 缓存管理器 ============

// Manager 缓存管理器
type Manager struct {
	mu      sync.RWMutex
	caches  map[string]Cache
	defaults map[string]int // 默认容量
}

// NewManager 创建缓存管理器
func NewManager() *Manager {
	return &Manager{
		caches:  make(map[string]Cache),
		defaults: map[string]int{
			"session":    1000,
			"message":    5000,
			"memory":     2000,
			"config":     100,
			"ai_response": 500,
			"embedding":  1000,
		},
	}
}

// Register 注册缓存
func (m *Manager) Register(name string, cache Cache) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.caches[name] = cache
}

// Get 获取缓存
func (m *Manager) Get(name string) Cache {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.caches[name]
}

// Create 创建缓存
func (m *Manager) Create(name string, capacity int, ttl time.Duration) Cache {
	m.mu.Lock()
	defer m.mu.Unlock()

	if capacity <= 0 {
		capacity = m.defaults[name]
		if capacity <= 0 {
			capacity = 100
		}
	}

	cache := NewLRUCache(capacity, ttl)
	m.caches[name] = cache

	return cache
}

// GetOrCreate 获取或创建缓存
func (m *Manager) GetOrCreate(name string, capacity int, ttl time.Duration) Cache {
	m.mu.RLock()
	cache := m.caches[name]
	m.mu.RUnlock()

	if cache != nil {
		return cache
	}

	return m.Create(name, capacity, ttl)
}

// ClearAll 清除所有缓存
func (m *Manager) ClearAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cache := range m.caches {
		cache.Clear()
	}
}

// Stats 获取统计
func (m *Manager) Stats() map[string]CacheStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]CacheStats)
	for name, cache := range m.caches {
		stats[name] = CacheStats{
			Size:     cache.Size(),
			Capacity: cache.Capacity(),
		}
	}
	return stats
}

// CacheStats 缓存统计
type CacheStats struct {
	Size     int
	Capacity int
	Hits     int64
	Misses   int64
}

// ============ Cached Client 缓存包装器 ============

// CachedClient 缓存包装的客户端
type CachedClient struct {
	cache *LRUCache
	client interface{}
}

// NewCachedClient 创建缓存客户端
func NewCachedClient(client interface{}, capacity int, ttl time.Duration) *CachedClient {
	return &CachedClient{
		cache:  NewLRUCache(capacity, ttl),
		client: client,
	}
}

// CacheKey 生成缓存键
func CacheKey(parts ...string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += ":"
		}
		result += part
	}
	return result
}

// ============ 辅助函数 ============

// WithCache 带缓存的函数执行
func WithCache[K comparable](cache Cache, key string, ttl time.Duration, fn func() (K, error)) (K, error) {
	// 尝试从缓存获取
	if val, ok := cache.Get(key); ok {
		if v, ok := val.(K); ok {
			return v, nil
		}
	}

	// 执行函数
	result, err := fn()
	if err != nil {
		return result, err
	}

	// 存入缓存
	cache.Set(key, result)

	return result, nil
}

// InvalidateCachePattern 使缓存失效 (支持前缀匹配)
func InvalidateCachePattern(cache Cache, pattern string) error {
	// 这需要知道缓存的内部结构
	// 简化实现：只清除所有
	return cache.Clear()
}

// ============ Context 缓存 (请求级缓存) ============

type contextKey string

const (
	cacheContextKey contextKey = "cache"
)

// WithContext 添加缓存到 Context
func WithContext(ctx context.Context, cache *LRUCache) context.Context {
	return context.WithValue(ctx, cacheContextKey, cache)
}

// FromContext 从 Context 获取缓存
func FromContext(ctx context.Context) *LRUCache {
	if cache, ok := ctx.Value(cacheContextKey).(*LRUCache); ok {
		return cache
	}
	return nil
}

// CachedContextCall 带上下文缓存的调用
func CachedContextCall[K comparable](ctx context.Context, key string, fn func() (K, error)) (K, error) {
	cache := FromContext(ctx)
	if cache == nil {
		return fn()
	}

	if val, ok := cache.Get(key); ok {
		if v, ok := val.(K); ok {
			return v, nil
		}
	}

	result, err := fn()
	if err != nil {
		return result, err
	}

	cache.Set(key, result)
	return result, nil
}
