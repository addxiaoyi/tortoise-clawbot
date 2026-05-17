import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:tortoise/features/settings/providers/settings_provider.dart';

void main() {
  group('Settings Provider Tests', () {
    test('初始状态应该使用默认设置', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final state = container.read(settingsProvider);
      expect(state.settings.themeMode, 'system');
      expect(state.settings.language, 'zh-CN');
    });

    test('更新主题模式应该生效', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final notifier = container.read(settingsProvider.notifier);
      notifier.updateThemeMode('dark');

      final state = container.read(settingsProvider);
      expect(state.settings.themeMode, 'dark');
    });

    test('更新语言应该生效', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final notifier = container.read(settingsProvider.notifier);
      notifier.updateLanguage('en-US');

      final state = container.read(settingsProvider);
      expect(state.settings.language, 'en-US');
    });

    test('更新通知设置应该生效', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final notifier = container.read(settingsProvider.notifier);
      notifier.updateNotifications(
        pushEnabled: false,
        soundEnabled: false,
      );

      final state = container.read(settingsProvider);
      expect(state.settings.pushEnabled, false);
      expect(state.settings.soundEnabled, false);
    });

    test('重置设置应该恢复默认值', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final notifier = container.read(settingsProvider.notifier);
      
      // 修改设置
      notifier.updateThemeMode('dark');
      notifier.updateLanguage('en-US');
      
      // 重置
      notifier.resetSettings();

      final state = container.read(settingsProvider);
      expect(state.settings.themeMode, 'system');
      expect(state.settings.language, 'zh-CN');
    });
  });
}
