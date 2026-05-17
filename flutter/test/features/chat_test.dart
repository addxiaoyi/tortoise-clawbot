import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:tortoise/features/chat/providers/chat_provider.dart';

void main() {
  group('Chat Provider Tests', () {
    test('初始状态应该包含默认会话', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final sessions = container.read(sessionsProvider);
      expect(sessions, isNotEmpty);
      expect(sessions.first.title, '默认会话');
    });

    test('创建新会话应该添加到列表', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final notifier = container.read(sessionsProvider.notifier);
      final sessionId = notifier.createSession(title: '测试会话');

      final sessions = container.read(sessionsProvider);
      expect(sessions.any((s) => s.id == sessionId), isTrue);
    });

    test('删除会话应该从列表中移除', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final notifier = container.read(sessionsProvider.notifier);
      final sessionId = notifier.createSession(title: '待删除会话');
      
      // 初始应该有2个会话
      expect(container.read(sessionsProvider).length, 2);

      notifier.deleteSession(sessionId);
      
      // 删除后应该只有1个会话
      expect(container.read(sessionsProvider).length, 1);
    });

    test('更新会话标题应该生效', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final notifier = container.read(sessionsProvider.notifier);
      final sessionId = notifier.createSession(title: '原标题');
      notifier.updateTitle(sessionId, '新标题');

      final sessions = container.read(sessionsProvider);
      final session = sessions.firstWhere((s) => s.id == sessionId);
      expect(session.title, '新标题');
    });
  });
}
