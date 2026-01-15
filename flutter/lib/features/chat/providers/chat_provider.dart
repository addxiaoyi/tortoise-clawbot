import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';
import '../models/chat_models.dart' as models;
import '../../../core/ai/ai_engine.dart';

const _uuid = Uuid();

/// 会话列表Notifier
class SessionsNotifier extends StateNotifier<List<models.ChatSession>> {
  SessionsNotifier() : super([
    models.ChatSession(
      id: 'default',
      title: '默认会话',
      messages: [],
      createdAt: DateTime.now(),
      updatedAt: DateTime.now(),
    ),
  ]);
  
  String createSession({String? title}) {
    final session = models.ChatSession(
      id: _uuid.v4(),
      title: title ?? '新会话',
      messages: [],
      createdAt: DateTime.now(),
      updatedAt: DateTime.now(),
    );
    state = [...state, session];
    return session.id;
  }
  
  void updateSession(models.ChatSession session) {
    state = state.map((s) => s.id == session.id ? session : s).toList();
  }
  
  void deleteSession(String sessionId) {
    if (state.length > 1) {
      state = state.where((s) => s.id != sessionId).toList();
    }
  }
  
  void updateTitle(String sessionId, String title) {
    state = state.map((s) {
      if (s.id == sessionId) {
        return models.ChatSession(
          id: s.id,
          title: title,
          messages: s.messages,
          createdAt: s.createdAt,
          updatedAt: DateTime.now(),
        );
      }
      return s;
    }).toList();
  }
}

/// 当前会话ID
final activeSessionIdProvider = StateProvider<String?>((ref) => 'default');

/// 会话列表
final sessionsProvider = StateNotifierProvider<SessionsNotifier, List<models.ChatSession>>((ref) {
  return SessionsNotifier();
});

/// 消息列表Notifier
class MessagesNotifier extends StateNotifier<Map<String, List<models.ChatMessage>>> {
  MessagesNotifier() : super({});
  
  void addMessage({
    required String sessionId,
    required String content,
    required String role,
    String? model,
  }) {
    final message = models.ChatMessage(
      id: _uuid.v4(),
      sessionId: sessionId,
      role: role,
      content: content,
      createdAt: DateTime.now(),
      model: model,
    );
    
    final current = state[sessionId] ?? [];
    state = {
      ...state,
      sessionId: [...current, message],
    };
  }
  
  List<models.ChatMessage> getMessages(String sessionId) {
    return state[sessionId] ?? [];
  }
  
  void clearSession(String sessionId) {
    state = {...state, sessionId: []};
  }
}

/// 聊天状态
enum ChatStatus { idle, loading, error }

/// 聊天状态Provider
final chatStatusProvider = StateProvider<ChatStatus>((ref) => ChatStatus.idle);
