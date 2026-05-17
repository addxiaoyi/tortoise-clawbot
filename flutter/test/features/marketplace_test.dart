import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:tortoise/features/marketplace/providers/marketplace_provider.dart';

void main() {
  group('Marketplace Provider Tests', () {
    test('初始状态应该包含示例插件', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final plugins = container.read(marketplacePluginsProvider);
      expect(plugins, isNotEmpty);
    });

    test('插件应该按类别分类', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final plugins = container.read(marketplacePluginsProvider);
      
      final channels = plugins.where((p) => p.category == 'channels').toList();
      final skills = plugins.where((p) => p.category == 'skills').toList();
      
      expect(channels, isNotEmpty);
      expect(skills, isNotEmpty);
    });

    test('安装插件应该更新状态', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final notifier = container.read(marketplacePluginsProvider.notifier);
      final plugins = container.read(marketplacePluginsProvider);
      final pluginToInstall = plugins.first;

      await notifier.installPlugin(pluginToInstall.id);

      final updatedPlugins = container.read(marketplacePluginsProvider);
      final installedPlugin = updatedPlugins.firstWhere((p) => p.id == pluginToInstall.id);
      expect(installedPlugin.isInstalled, isTrue);
    });

    test('卸载插件应该更新状态', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final notifier = container.read(marketplacePluginsProvider.notifier);
      final plugins = container.read(marketplacePluginsProvider);
      final pluginToUninstall = plugins.first;

      // 先安装
      await notifier.installPlugin(pluginToUninstall.id);
      
      // 再卸载
      await notifier.uninstallPlugin(pluginToUninstall.id);

      final updatedPlugins = container.read(marketplacePluginsProvider);
      final uninstalledPlugin = updatedPlugins.firstWhere((p) => p.id == pluginToUninstall.id);
      expect(uninstalledPlugin.isInstalled, isFalse);
    });
  });
}
