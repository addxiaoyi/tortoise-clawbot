// Chat BLoC

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:equatable/equatable.dart';

/// Message Role
enum MessageRole { user, assistant, system }

/// Message Status
enum MessageStatus { sending, sent, error }

/// Message Model
class Message extends Equatable {
  final String id;
  final MessageRole role;
  final String content;
  final DateTime timestamp;
  final MessageStatus status;
  final bool isStreaming;
  final String? agentName;
  final String? agentRole;

  const Message({
    required this.id,
    required this.role,
    required this.content,
    required this.timestamp,
    this.status = MessageStatus.sent,
    this.isStreaming = false,
    this.agentName,
    this.agentRole,
  });

  Message copyWith({
    String? id,
    MessageRole? role,
    String? content,
    DateTime? timestamp,
    MessageStatus? status,
    bool? isStreaming,
    String? agentName,
    String? agentRole,
  }) {
    return Message(
      id: id ?? this.id,
      role: role ?? this.role,
      content: content ?? this.content,
      timestamp: timestamp ?? this.timestamp,
      status: status ?? this.status,
      isStreaming: isStreaming ?? this.isStreaming,
      agentName: agentName ?? this.agentName,
      agentRole: agentRole ?? this.agentRole,
    );
  }

  @override
  List<Object?> get props => [id, role, content, timestamp, status, isStreaming, agentName, agentRole];
}

/// Chat State
class ChatState extends Equatable {
  final List<Message> messages;
  final bool isLoading;
  final bool isTyping;
  final String? error;

  const ChatState({
    this.messages = const [],
    this.isLoading = false,
    this.isTyping = false,
    this.error,
  });

  ChatState copyWith({
    List<Message>? messages,
    bool? isLoading,
    bool? isTyping,
    String? error,
  }) {
    return ChatState(
      messages: messages ?? this.messages,
      isLoading: isLoading ?? this.isLoading,
      isTyping: isTyping ?? this.isTyping,
      error: error,
    );
  }

  @override
  List<Object?> get props => [messages, isLoading, isTyping, error];
}

/// Chat BLoC
class ChatBloc extends StateNotifier<ChatState> {
  ChatBloc() : super(const ChatState());

  Future<void> sendMessage(String content) async {
    if (content.trim().isEmpty) return;

    final userMessage = Message(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      role: MessageRole.user,
      content: content,
      timestamp: DateTime.now(),
      status: MessageStatus.sent,
    );

    state = state.copyWith(
      messages: [...state.messages, userMessage],
      isLoading: true,
    );

    // Simulate API call
    await Future.delayed(const Duration(seconds: 1));

    final assistantMessage = Message(
      id: (DateTime.now().millisecondsSinceEpoch + 1).toString(),
      role: MessageRole.assistant,
      content: 'This is a placeholder response. In production, this would be the actual AI response.',
      timestamp: DateTime.now(),
      status: MessageStatus.sent,
      agentName: 'Tortoise',
    );

    state = state.copyWith(
      messages: [...state.messages, assistantMessage],
      isLoading: false,
      isTyping: false,
    );
  }

  Future<void> retry(String messageId) async {
    final index = state.messages.indexWhere((m) => m.id == messageId);
    if (index == -1) return;

    final message = state.messages[index];
    state = state.copyWith(
      messages: state.messages.sublist(0, index),
    );
    await sendMessage(message.content);
  }

  void deleteMessage(String messageId) {
    state = state.copyWith(
      messages: state.messages.where((m) => m.id != messageId).toList(),
    );
  }

  void clearChat() {
    state = const ChatState();
  }
}

/// Provider
final chatBlocProvider = StateNotifierProvider<ChatBloc, ChatState>((ref) {
  return ChatBloc();
});
