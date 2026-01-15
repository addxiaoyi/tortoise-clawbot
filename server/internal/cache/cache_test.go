package cache

import (
	"testing"
	"time"
)

// ============ LRUCache Tests ============

func TestLRUCache_SetAndGet(t *testing.T) {
	cache := NewLRUCache(100, 0)
	
	cache.Set("key1", "value1")
	
	val, ok := cache.Get("key1")
	if !ok {
		t.Error("Get should return true for existing key")
	}
	
	if val != "value1" {
		t.Errorf("Expected 'value1', got '%v'", val)
	}
}

func TestLRUCache_Update(t *testing.T) {
	cache := NewLRUCache(100, 0)
	
	cache.Set("key1", "value1")
	cache.Set("key1", "value2")
	
	val, _ := cache.Get("key1")
	if val != "value2" {
		t.Errorf("Expected 'value2', got '%v'", val)
	}
	
	// 大小应该不变
	if cache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cache.Size())
	}
}

func TestLRUCache_Delete(t *testing.T) {
	cache := NewLRUCache(100, 0)
	
	cache.Set("key1", "value1")
	cache.Delete("key1")
	
	_, ok := cache.Get("key1")
	if ok {
		t.Error("Get should return false after delete")
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRUCache(3, 0)
	
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	
	// 触发驱逐
	cache.Set("key4", "value4")
	
	// key1 应该被驱逐
	_, ok := cache.Get("key1")
	if ok {
		t.Error("key1 should have been evicted")
	}
	
	// 其他 key 应该还在
	_, ok = cache.Get("key2")
	if !ok {
		t.Error("key2 should still exist")
	}
	
	_, ok = cache.Get("key3")
	if !ok {
		t.Error("key3 should still exist")
	}
	
	_, ok = cache.Get("key4")
	if !ok {
		t.Error("key4 should exist")
	}
}

func TestLRUCache_TTL(t *testing.T) {
	cache := NewLRUCache(100, 100*time.Millisecond)
	
	cache.Set("key1", "value1")
	
	// 立即获取应该成功
	_, ok := cache.Get("key1")
	if !ok {
		t.Error("Get should succeed immediately after set")
	}
	
	// 等待过期
	time.Sleep(150 * time.Millisecond)
	
	// 应该过期
	_, ok = cache.Get("key1")
	if ok {
		t.Error("Get should return false after TTL expires")
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(100, 0)
	
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	
	cache.Clear()
	
	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}
}

func TestLRUCache_LRUOrder(t *testing.T) {
	cache := NewLRUCache(3, 0)
	
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	
	// 访问 key1 使其变为最新
	cache.Get("key1")
	
	// 添加新key，触发驱逐
	cache.Set("key4", "value4")
	
	// key2 应该被驱逐 (最久未访问)
	_, ok := cache.Get("key2")
	if ok {
		t.Error("key2 should have been evicted (LRU)")
	}
}

// ============ MapCache Tests ============

func TestMapCache_SetAndGet(t *testing.T) {
	cache := NewMapCache(0)
	
	cache.Set("key1", "value1")
	
	val, ok := cache.Get("key1")
	if !ok {
		t.Error("Get should return true")
	}
	
	if val != "value1" {
		t.Errorf("Expected 'value1', got '%v'", val)
	}
}

func TestMapCache_Delete(t *testing.T) {
	cache := NewMapCache(0)
	
	cache.Set("key1", "value1")
	cache.Delete("key1")
	
	_, ok := cache.Get("key1")
	if ok {
		t.Error("Get should return false after delete")
	}
}

func TestMapCache_Keys(t *testing.T) {
	cache := NewMapCache(0)
	
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	
	keys := cache.Keys()
	
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}
}

func TestMapCache_Clear(t *testing.T) {
	cache := NewMapCache(0)
	
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	
	cache.Clear()
	
	if cache.Size() != 0 {
		t.Errorf("Expected size 0, got %d", cache.Size())
	}
}

// ============ Cache Manager Tests ============

func TestCacheManager_Register(t *testing.T) {
	manager := NewManager()
	cache := NewLRUCache(100, 0)
	
	manager.Register("test", cache)
	
	retrieved := manager.Get("test")
	if retrieved == nil {
		t.Error("Get should return registered cache")
	}
}

func TestCacheManager_Create(t *testing.T) {
	manager := NewManager()
	
	cache := manager.Create("session", 50, 0)
	
	if cache == nil {
		t.Fatal("Create should return a cache")
	}
	
	if cache.Capacity() != 50 {
		t.Errorf("Expected capacity 50, got %d", cache.Capacity())
	}
}

func TestCacheManager_GetOrCreate(t *testing.T) {
	manager := NewManager()
	
	// 第一次创建
	cache1 := manager.GetOrCreate("session", 100, 0)
	
	// 第二次应该返回同一个
	cache2 := manager.GetOrCreate("session", 200, 0)
	
	if cache1 != cache2 {
		t.Error("GetOrCreate should return the same cache instance")
	}
}

func TestCacheManager_ClearAll(t *testing.T) {
	manager := NewManager()
	
	manager.Create("cache1", 100, 0).Set("key1", "value1")
	manager.Create("cache2", 100, 0).Set("key2", "value2")
	
	manager.ClearAll()
	
	if manager.Get("cache1").Size() != 0 {
		t.Error("cache1 should be empty after ClearAll")
	}
	
	if manager.Get("cache2").Size() != 0 {
		t.Error("cache2 should be empty after ClearAll")
	}
}

func TestCacheManager_Stats(t *testing.T) {
	manager := NewManager()
	
	manager.Create("session", 100, 0)
	manager.Get("session").Set("key1", "value1")
	manager.Get("session").Set("key2", "value2")
	
	stats := manager.Stats()
	
	if stats["session"].Size != 2 {
		t.Errorf("Expected size 2, got %d", stats["session"].Size)
	}
	
	if stats["session"].Capacity != 100 {
		t.Errorf("Expected capacity 100, got %d", stats["session"].Capacity)
	}
}

// ============ Benchmark Tests ============

func BenchmarkLRUCache_Set(b *testing.B) {
	cache := NewLRUCache(10000, 0)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(string(rune(i)), "value")
	}
}

func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(10000, 0)
	
	// 填充缓存
	for i := 0; i < 1000; i++ {
		cache.Set(string(rune(i)), "value")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key500")
	}
}

func BenchmarkMapCache_Set(b *testing.B) {
	cache := NewMapCache(0)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(string(rune(i)), "value")
	}
}

func BenchmarkMapCache_Get(b *testing.B) {
	cache := NewMapCache(0)
	
	// 填充缓存
	for i := 0; i < 1000; i++ {
		cache.Set(string(rune(i)), "value")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key500")
	}
}
