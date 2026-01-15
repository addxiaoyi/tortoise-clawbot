import 'dart:async';
import '../storage/storage_service.dart';

/// 缓存服务
class CacheService {
  static CacheService? _instance;
  static CacheService get instance => _instance ??= CacheService._();
  CacheService._();

  final StorageService _storage = StorageService.instance;
  final Map<String, _CacheEntry> _memoryCache = {};
  Timer? _cleanupTimer;

  /// 初始化
  void initialize() {
    _startCleanupTimer();
  }

  /// 启动清理定时器
  void _startCleanupTimer() {
    _cleanupTimer?.cancel();
    _cleanupTimer = Timer.periodic(const Duration(minutes: 5), (_) {
      _cleanupExpired();
    });
  }

  /// 设置缓存
  Future<void> set(String key, dynamic value, {Duration? expiresIn}) async {
    final entry = _CacheEntry(
      value: value,
      expiresAt: expiresIn != null ? DateTime.now().add(expiresIn) : null,
    );
    _memoryCache[key] = entry;
    await _storage.saveConfig('cache_$key', value);
  }

  /// 获取缓存
  T? get<T>(String key) {
    final entry = _memoryCache[key];
    if (entry == null) return null;
    
    if (entry.expiresAt != null && entry.expiresAt!.isBefore(DateTime.now())) {
      _memoryCache.remove(key);
      return null;
    }
    
    return entry.value as T?;
  }

  /// 检查是否存在
  bool contains(String key) {
    final entry = _memoryCache[key];
    if (entry == null) return false;
    
    if (entry.expiresAt != null && entry.expiresAt!.isBefore(DateTime.now())) {
      _memoryCache.remove(key);
      return false;
    }
    
    return true;
  }

  /// 删除缓存
  Future<void> remove(String key) async {
    _memoryCache.remove(key);
    // 注意：Hive 不支持直接删除键，需要清空整个 box 或使用其他方式
  }

  /// 清空所有缓存
  Future<void> clear() async {
    _memoryCache.clear();
  }

  /// 清理过期缓存
  void _cleanupExpired() {
    final now = DateTime.now();
    _memoryCache.removeWhere((key, entry) {
      return entry.expiresAt != null && entry.expiresAt!.isBefore(now);
    });
  }

  /// 获取缓存统计
  CacheStats getStats() {
    return CacheStats(
      count: _memoryCache.length,
      keys: _memoryCache.keys.toList(),
    );
  }

  /// 销毁
  void dispose() {
    _cleanupTimer?.cancel();
    _memoryCache.clear();
  }
}

/// 缓存条目
class _CacheEntry {
  final dynamic value;
  final DateTime? expiresAt;

  _CacheEntry({required this.value, this.expiresAt});
}

/// 缓存统计
class CacheStats {
  final int count;
  final List<String> keys;

  CacheStats({required this.count, required this.keys});
}
