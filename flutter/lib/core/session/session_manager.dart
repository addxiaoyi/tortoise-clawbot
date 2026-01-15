import 'dart:async';
import '../ai/ai_engine.dart';
import '../storage/storage_service.dart';

/// 会话管理器 - 管理对话历史
class SessionManager {
  static SessionManager? _instance;
  static SessionManager get instance => _instance ??= SessionManager._();
  SessionManager._();

  final StorageService _storage = StorageService.instance;
  final _sessionsController = StreamController<List<Session>>.broadcast();
  
  final Map<String, Session> _sessions = {};
  String? _currentSessionId;

  Stream<List<Session>> get sessionsStream => _sessionsController.stream;
  
  List<Session> get sessions => _sessions.values.toList()
    ..sort((a, b) => b.updatedAt.compareTo(a.updatedAt));
  
  Session? get currentSession => 
    _currentSessionId != null ? _sessions[_currentSessionId] : null;

  /// 初始化
  Future<void> initialize() async {
    await _loadSessions();
  }

  /// 创建新会话
  Session createSession({String? title, String model = 'gpt-4'}) {
    final now = DateTime.now();
    final id = now.millisecondsSinceEpoch.toString();
    
    final session = Session(
      id: id,
      title: title ?? '新对话 ${_sessions.length + 1}',
      createdAt: now,
      updatedAt: now,
      model: model,
      messages: [],
    );
    
    _sessions[id] = session;
    _currentSessionId = id;
    _notify();
    _saveSession(session);
    
    return session;
  }

  /// 选择会话
  void selectSession(String id) {
    if (_sessions.containsKey(id)) {
      _currentSessionId = id;
    }
  }

  /// 添加消息
  void addMessage(ChatMessage message) {
    if (currentSession == null) {
      createSession();
    }
    
    final session = currentSession!;
    final updatedMessages = [...session.messages, message];
    
    final updatedSession = session.copyWith(
      messages: updatedMessages,
      updatedAt: DateTime.now(),
    );
    
    _sessions[session.id] = updatedSession;
    _currentSessionId = session.id;
    _notify();
    _saveSession(updatedSession);
  }

  /// 删除会话
  Future<void> deleteSession(String id) async {
    _sessions.remove(id);
    if (_currentSessionId == id) {
      _currentSessionId = _sessions.isNotEmpty ? _sessions.keys.first : null;
    }
    await _storage.deleteSession(id);
    _notify();
  }

  /// 清空当前会话
  void clearCurrentSession() {
    if (currentSession == null) return;
    
    final updated = currentSession!.copyWith(
      messages: [],
      updatedAt: DateTime.now(),
    );
    _sessions[currentSession!.id] = updated;
    _saveSession(updated);
    _notify();
  }

  /// 加载会话
  Future<void> _loadSessions() async {
    final allSessions = _storage.getAllSessions();
    for (final data in allSessions) {
      try {
        final session = Session.fromJson(data);
        _sessions[session.id] = session;
      } catch (_) {}
    }
    _notify();
  }

  /// 保存会话
  void _saveSession(Session session) {
    _storage.saveSession(session.id, session.toJson());
  }

  void _notify() {
    _sessionsController.add(sessions);
  }

  void dispose() {
    _sessionsController.close();
  }
}

/// 会话数据模型
class Session {
  final String id;
  final String title;
  final DateTime createdAt;
  final DateTime updatedAt;
  final List<ChatMessage> messages;
  final String model;

  Session({
    required this.id,
    required this.title,
    required this.createdAt,
    required this.updatedAt,
    this.messages = const [],
    this.model = 'gpt-4',
  });

  Session copyWith({
    String? title,
    DateTime? updatedAt,
    List<ChatMessage>? messages,
    String? model,
  }) {
    return Session(
      id: id,
      title: title ?? this.title,
      createdAt: createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
      messages: messages ?? this.messages,
      model: model ?? this.model,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'createdAt': createdAt.toIso8601String(),
      'updatedAt': updatedAt.toIso8601String(),
      'model': model,
      'messages': messages.map((m) => {'role': m.role, 'content': m.content}).toList(),
    };
  }

  factory Session.fromJson(Map<String, dynamic> json) {
    return Session(
      id: (json['id'] ?? '').toString(),
      title: (json['title'] ?? '对话').toString(),
      createdAt: DateTime.tryParse((json['createdAt'] ?? '').toString()) ?? DateTime.now(),
      updatedAt: DateTime.tryParse((json['updatedAt'] ?? '').toString()) ?? DateTime.now(),
      model: (json['model'] ?? 'gpt-4').toString(),
      messages: (json['messages'] as List? ?? [])
          .map((m) => ChatMessage(
            role: (m is Map ? (m['role'] ?? 'user') : 'user').toString(),
            content: (m is Map ? (m['content'] ?? '') : '').toString(),
          ))
          .toList(),
    );
  }
}
