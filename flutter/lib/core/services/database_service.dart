import 'package:hive_flutter/hive_flutter.dart';

/// 数据库服务 - 基于 Hive 的本地存储
class DatabaseService {
  static DatabaseService? _instance;
  
  // Box 名称
  static const String sessionsBoxName = 'sessions';
  static const String messagesBoxName = 'messages';
  static const String memoriesBoxName = 'memories';
  static const String configBoxName = 'config';
  static const String cacheBoxName = 'cache';
  
  late Box<Map> _sessionsBox;
  late Box<Map> _messagesBox;
  late Box<Map> _memoriesBox;
  late Box<Map> _configBox;
  late Box<Map> _cacheBox;
  
  DatabaseService._();
  
  static DatabaseService get instance {
    _instance ??= DatabaseService._();
    return _instance!;
  }
  
  /// 初始化
  Future<void> init() async {
    _sessionsBox = await Hive.openBox<Map>(sessionsBoxName);
    _messagesBox = await Hive.openBox<Map>(messagesBoxName);
    _memoriesBox = await Hive.openBox<Map>(memoriesBoxName);
    _configBox = await Hive.openBox<Map>(configBoxName);
    _cacheBox = await Hive.openBox<Map>(cacheBoxName);
  }
  
  // ========== 会话操作 ==========
  
  /// 保存会话
  Future<void> saveSession(Session session) async {
    await _sessionsBox.put(session.id, session.toJson());
  }
  
  /// 获取会话
  Session? getSession(String id) {
    final data = _sessionsBox.get(id);
    return data != null ? Session.fromJson(Map<String, dynamic>.from(data)) : null;
  }
  
  /// 获取所有会话
  List<Session> getAllSessions() {
    return _sessionsBox.values
        .map((data) => Session.fromJson(Map<String, dynamic>.from(data)))
        .toList()
      ..sort((a, b) => b.updatedAt.compareTo(a.updatedAt));
  }
  
  /// 删除会话
  Future<void> deleteSession(String id) async {
    await _sessionsBox.delete(id);
  }
  
  // ========== 消息操作 ==========
  
  /// 保存消息
  Future<void> saveMessage(Message message) async {
    await _messagesBox.put(message.id, message.toJson());
  }
  
  /// 获取会话的消息
  List<Message> getSessionMessages(String sessionId) {
    return _messagesBox.values
        .where((m) => m['session_id'] == sessionId)
        .map((data) => Message.fromJson(Map<String, dynamic>.from(data)))
        .toList()
      ..sort((a, b) => a.createdAt.compareTo(b.createdAt));
  }
  
  /// 删除消息
  Future<void> deleteMessage(String id) async {
    await _messagesBox.delete(id);
  }
  
  // ========== 记忆操作 ==========
  
  /// 保存记忆
  Future<void> saveMemory(Memory memory) async {
    await _memoriesBox.put(memory.id, memory.toJson());
  }
  
  /// 获取记忆
  Memory? getMemory(String id) {
    final data = _memoriesBox.get(id);
    return data != null ? Memory.fromJson(Map<String, dynamic>.from(data)) : null;
  }
  
  /// 获取所有记忆
  List<Memory> getAllMemories() {
    return _memoriesBox.values
        .map((data) => Memory.fromJson(Map<String, dynamic>.from(data)))
        .toList();
  }
  
  /// 删除记忆
  Future<void> deleteMemory(String id) async {
    await _memoriesBox.delete(id);
  }
  
  // ========== 配置操作 ==========
  
  /// 保存配置
  Future<void> saveConfig(String key, dynamic value) async {
    await _configBox.put(key, value is Map ? value : {'value': value});
  }
  
  /// 获取配置
  dynamic getConfig(String key) {
    return _configBox.get(key)?['value'];
  }
  
  // ========== 缓存操作 ==========
  
  /// 保存缓存
  Future<void> saveCache(String key, dynamic value) async {
    await _cacheBox.put(key, value is Map ? value : {'value': value});
  }
  
  /// 获取缓存
  dynamic getCache(String key) {
    return _cacheBox.get(key)?['value'];
  }
  
  /// 清除缓存
  Future<void> clearCache() async {
    await _cacheBox.clear();
  }
  
  /// 清除所有数据
  Future<void> clearAll() async {
    await _sessionsBox.clear();
    await _messagesBox.clear();
    await _memoriesBox.clear();
    await _configBox.clear();
    await _cacheBox.clear();
  }
}

/// 会话模型
class Session {
  final String id;
  final String title;
  final DateTime createdAt;
  final DateTime updatedAt;

  Session({
    required this.id,
    required this.title,
    required this.createdAt,
    required this.updatedAt,
  });

  Map<String, dynamic> toJson() => {
    'id': id,
    'title': title,
    'created_at': createdAt.millisecondsSinceEpoch,
    'updated_at': updatedAt.millisecondsSinceEpoch,
  };

  factory Session.fromJson(Map<String, dynamic> json) => Session(
    id: json['id'] as String? ?? '',
    title: json['title'] as String? ?? '',
    createdAt: DateTime.fromMillisecondsSinceEpoch(json['created_at'] as int? ?? 0),
    updatedAt: DateTime.fromMillisecondsSinceEpoch(json['updated_at'] as int? ?? 0),
  );
}

/// 消息模型
class Message {
  final String id;
  final String sessionId;
  final String role;
  final String content;
  final DateTime createdAt;

  Message({
    required this.id,
    required this.sessionId,
    required this.role,
    required this.content,
    required this.createdAt,
  });

  Map<String, dynamic> toJson() => {
    'id': id,
    'session_id': sessionId,
    'role': role,
    'content': content,
    'created_at': createdAt.millisecondsSinceEpoch,
  };

  factory Message.fromJson(Map<String, dynamic> json) => Message(
    id: json['id'] as String? ?? '',
    sessionId: json['session_id'] as String? ?? '',
    role: json['role'] as String? ?? 'user',
    content: json['content'] as String? ?? '',
    createdAt: DateTime.fromMillisecondsSinceEpoch(json['created_at'] as int? ?? 0),
  );
}

/// 记忆模型
class Memory {
  final String id;
  final String title;
  final String content;
  final String type;
  final List<String> tags;
  final DateTime createdAt;
  final DateTime updatedAt;

  Memory({
    required this.id,
    required this.title,
    required this.content,
    this.type = 'general',
    this.tags = const [],
    required this.createdAt,
    required this.updatedAt,
  });

  Map<String, dynamic> toJson() => {
    'id': id,
    'title': title,
    'content': content,
    'type': type,
    'tags': tags,
    'created_at': createdAt.millisecondsSinceEpoch,
    'updated_at': updatedAt.millisecondsSinceEpoch,
  };

  factory Memory.fromJson(Map<String, dynamic> json) => Memory(
    id: json['id'] as String? ?? '',
    title: json['title'] as String? ?? '',
    content: json['content'] as String? ?? '',
    type: json['type'] as String? ?? 'general',
    tags: (json['tags'] as List?)?.cast<String>() ?? [],
    createdAt: DateTime.fromMillisecondsSinceEpoch(json['created_at'] as int? ?? 0),
    updatedAt: DateTime.fromMillisecondsSinceEpoch(json['updated_at'] as int? ?? 0),
  );
}
