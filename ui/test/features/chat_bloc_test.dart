import 'package:flutter_test/flutter_test.dart';
import 'package:tortoise/features/chat/bloc/chat_bloc.dart';

void main() {
  group('ChatBloc', () {
    late ChatBloc chatBloc;

    setUp(() {
      chatBloc = ChatBloc();
    });

    tearDown(() {
      chatBloc.close();
    });

    test('initial state is empty', () {
      expect(chatBloc.state.messages, isEmpty);
      expect(chatBloc.state.isLoading, isFalse);
      expect(chatBloc.state.isTyping, isFalse);
      expect(chatBloc.state.error, isNull);
    });

    test('sendMessage adds user message', () async {
      await chatBloc.sendMessage('Hello');

      expect(chatBloc.state.messages.length, 2); // user + assistant
      expect(chatBloc.state.messages.first.role, MessageRole.user);
      expect(chatBloc.state.messages.first.content, 'Hello');
    });

    test('sendMessage with empty content does nothing', () async {
      await chatBloc.sendMessage('');

      expect(chatBloc.state.messages, isEmpty);
    });

    test('sendMessage sets isLoading while processing', () async {
      final future = chatBloc.sendMessage('Test');

      // The state should be loading during processing
      expect(chatBloc.state.isLoading, isTrue);

      await future;

      expect(chatBloc.state.isLoading, isFalse);
    });

    test('clearChat removes all messages', () async {
      await chatBloc.sendMessage('Hello');
      chatBloc.clearChat();

      expect(chatBloc.state.messages, isEmpty);
    });

    test('deleteMessage removes specific message', () async {
      await chatBloc.sendMessage('Hello');
      await chatBloc.sendMessage('World');

      expect(chatBloc.state.messages.length, 4); // 2 user + 2 assistant

      final firstUserId = chatBloc.state.messages.first.id;
      chatBloc.deleteMessage(firstUserId);

      expect(chatBloc.state.messages.length, 3);
    });

    test('retry removes message and resends', () async {
      await chatBloc.sendMessage('Hello');
      
      final userMessageId = chatBloc.state.messages.first.id;
      await chatBloc.retry(userMessageId);

      // Should have removed the original messages and sent again
      expect(chatBloc.state.messages.any((m) => m.id == userMessageId), isFalse);
    });
  });

  group('Message', () {
    test('copyWith creates new instance with updated values', () {
      final original = Message(
        id: '1',
        role: MessageRole.user,
        content: 'Hello',
        timestamp: DateTime(2024, 1, 1),
      );

      final copied = original.copyWith(content: 'World');

      expect(copied.id, '1');
      expect(copied.content, 'World');
      expect(copied.role, MessageRole.user);
    });

    test('equality works correctly', () {
      final msg1 = Message(
        id: '1',
        role: MessageRole.user,
        content: 'Hello',
        timestamp: DateTime(2024, 1, 1),
      );

      final msg2 = Message(
        id: '1',
        role: MessageRole.user,
        content: 'Hello',
        timestamp: DateTime(2024, 1, 1),
      );

      expect(msg1, equals(msg2));
    });
  });

  group('ChatState', () {
    test('copyWith preserves values', () {
      final original = const ChatState(
        messages: [],
        isLoading: false,
        isTyping: false,
        error: null,
      );

      final copied = original.copyWith(isLoading: true);

      expect(copied.isLoading, isTrue);
      expect(copied.messages, isEmpty);
    });
  });
}
