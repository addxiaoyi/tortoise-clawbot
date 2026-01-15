import '../core/ai/ai_engine.dart';

/// 对话会话
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
}

/// 会话管理器
class SessionProvider {
  static SessionProvider? _instance;
  static SessionProvider get instance => _instance ??= SessionProvider._();

  SessionProvider._();

  final List<Session> _sessions = [];
  Session? _currentSession;

  List<Session> get sessions => List.unmodifiable(_sessions);
  Session? get currentSession => _currentSession;

  Session createSession({String? title, String model = 'gpt-4'}) {
    final now = DateTime.now();
    final session = Session(
      id: now.millisecondsSinceEpoch.toString(),
      title: title ?? '新对话 ${_sessions.length + 1}',
      createdAt: now,
      updatedAt: now,
      model: model,
    );
    _sessions.insert(0, session);
    _currentSession = session;
    return session;
  }

  void selectSession(Session session) {
    _currentSession = session;
  }

  void addMessage(ChatMessage message) {
    if (_currentSession == null) {
      createSession();
    }
    final updatedMessages = [..._currentSession!.messages, message];
    _currentSession = _currentSession!.copyWith(
      messages: updatedMessages,
      updatedAt: DateTime.now(),
    );
    
    final index = _sessions.indexWhere((s) => s.id == _currentSession!.id);
    if (index >= 0) {
      _sessions[index] = _currentSession!;
    }
  }

  void deleteSession(String id) {
    _sessions.removeWhere((s) => s.id == id);
    if (_currentSession?.id == id) {
      _currentSession = _sessions.isNotEmpty ? _sessions.first : null;
    }
  }

  void clearCurrentSession() {
    if (_currentSession != null) {
      _currentSession = _currentSession!.copyWith(messages: [], updatedAt: DateTime.now());
      final index = _sessions.indexWhere((s) => s.id == _currentSession!.id);
      if (index >= 0) {
        _sessions[index] = _currentSession!;
      }
    }
  }
}
