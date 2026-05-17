import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/settings_provider.dart';

class SettingsPage extends ConsumerWidget {
  const SettingsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final settings = ref.watch(settingsProvider);
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('设置'),
      ),
      body: ListView(
        children: [
          // AI Settings Section
          _SectionHeader(title: 'AI 设置'),
          ListTile(
            leading: const Icon(Icons.smart_toy),
            title: const Text('AI 提供商'),
            subtitle: Text(_getProviderName(settings.aiProvider)),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => _showProviderSelector(context, ref),
          ),
          ListTile(
            leading: const Icon(Icons.model_training),
            title: const Text('模型'),
            subtitle: Text(settings.aiModel),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => _showModelSelector(context, ref),
          ),
          ListTile(
            leading: const Icon(Icons.key),
            title: const Text('API 密钥'),
            subtitle: Text(
              settings.apiKey.isEmpty ? '未设置' : '••••••••${settings.apiKey.substring(settings.apiKey.length > 8 ? settings.apiKey.length - 8 : 0)}',
            ),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => _showApiKeyDialog(context, ref),
          ),
          ListTile(
            leading: const Icon(Icons.link),
            title: const Text('API 端点'),
            subtitle: Text(settings.apiEndpoint),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => _showEndpointDialog(context, ref),
          ),
          const Divider(),

          // Voice Settings Section
          _SectionHeader(title: '语音设置'),
          SwitchListTile(
            secondary: const Icon(Icons.mic),
            title: const Text('语音唤醒'),
            subtitle: const Text('使用唤醒词激活助手'),
            value: settings.voiceEnabled,
            onChanged: (value) {
              ref.read(settingsProvider.notifier).updateVoiceEnabled(value);
            },
          ),
          ListTile(
            leading: const Icon(Icons.record_voice_over),
            title: const Text('唤醒词'),
            subtitle: Text(settings.wakeWord),
            trailing: const Icon(Icons.chevron_right),
            enabled: settings.voiceEnabled,
            onTap: settings.voiceEnabled
                ? () => _showWakeWordDialog(context, ref)
                : null,
          ),
          ListTile(
            leading: const Icon(Icons.sensitivity),
            title: const Text('灵敏度'),
            subtitle: Slider(
              value: settings.wakeSensitivity,
              min: 0.1,
              max: 1.0,
              divisions: 9,
              label: '${(settings.wakeSensitivity * 100).round()}%',
              onChanged: settings.voiceEnabled
                  ? (value) {
                      ref.read(settingsProvider.notifier).updateWakeSensitivity(value);
                    }
                  : null,
            ),
          ),
          const Divider(),

          // Appearance Section
          _SectionHeader(title: '外观'),
          ListTile(
            leading: const Icon(Icons.palette),
            title: const Text('主题模式'),
            subtitle: Text(_getThemeModeName(settings.themeMode)),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => _showThemeSelector(context, ref),
          ),
          const Divider(),

          // Notifications Section
          _SectionHeader(title: '通知'),
          SwitchListTile(
            secondary: const Icon(Icons.notifications),
            title: const Text('启用通知'),
            subtitle: const Text('接收新消息通知'),
            value: settings.notificationsEnabled,
            onChanged: (value) {
              ref.read(settingsProvider.notifier).updateNotificationsEnabled(value);
            },
          ),
          const Divider(),

          // Language Section
          _SectionHeader(title: '语言'),
          ListTile(
            leading: const Icon(Icons.language),
            title: const Text('界面语言'),
            subtitle: Text(_getLanguageName(settings.language)),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => _showLanguageSelector(context, ref),
          ),
          const Divider(),

          // About Section
          _SectionHeader(title: '关于'),
          ListTile(
            leading: const Icon(Icons.info),
            title: const Text('版本'),
            subtitle: const Text('1.0.0'),
          ),
          ListTile(
            leading: const Icon(Icons.code),
            title: const Text('开源协议'),
            subtitle: const Text('Apache 2.0'),
            trailing: const Icon(Icons.open_in_new),
            onTap: () {
              // TODO: Open license
            },
          ),
          const SizedBox(height: 32),
        ],
      ),
    );
  }

  String _getProviderName(String provider) {
    switch (provider) {
      case 'openai':
        return 'OpenAI';
      case 'anthropic':
        return 'Anthropic';
      case 'google':
        return 'Google';
      default:
        return provider;
    }
  }

  String _getThemeModeName(ThemeMode mode) {
    switch (mode) {
      case ThemeMode.system:
        return '跟随系统';
      case ThemeMode.light:
        return '浅色模式';
      case ThemeMode.dark:
        return '深色模式';
    }
  }

  String _getLanguageName(String code) {
    switch (code) {
      case 'zh-CN':
        return '简体中文';
      case 'zh-TW':
        return '繁體中文';
      case 'en':
        return 'English';
      case 'ja':
        return '日本語';
      default:
        return code;
    }
  }

  void _showProviderSelector(BuildContext context, WidgetRef ref) {
    showModalBottomSheet(
      context: context,
      builder: (context) => Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          ListTile(
            leading: const Icon(Icons.smart_toy),
            title: const Text('OpenAI'),
            onTap: () {
              ref.read(settingsProvider.notifier).updateAiProvider('openai');
              Navigator.pop(context);
            },
          ),
          ListTile(
            leading: const Icon(Icons.smart_toy),
            title: const Text('Anthropic'),
            onTap: () {
              ref.read(settingsProvider.notifier).updateAiProvider('anthropic');
              Navigator.pop(context);
            },
          ),
          ListTile(
            leading: const Icon(Icons.smart_toy),
            title: const Text('Google'),
            onTap: () {
              ref.read(settingsProvider.notifier).updateAiProvider('google');
              Navigator.pop(context);
            },
          ),
          const SizedBox(height: 16),
        ],
      ),
    );
  }

  void _showModelSelector(BuildContext context, WidgetRef ref) {
    final settings = ref.read(settingsProvider);
    List<String> models;

    switch (settings.aiProvider) {
      case 'openai':
        models = ['gpt-4o', 'gpt-4-turbo', 'gpt-4', 'gpt-3.5-turbo'];
        break;
      case 'anthropic':
        models = ['claude-3-5-sonnet', 'claude-3-opus', 'claude-3-sonnet', 'claude-3-haiku'];
        break;
      case 'google':
        models = ['gemini-2.0-flash', 'gemini-1.5-pro', 'gemini-1.5-flash'];
        break;
      default:
        models = [];
    }

    showModalBottomSheet(
      context: context,
      builder: (context) => Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          ...models.map((model) => ListTile(
            title: Text(model),
            trailing: settings.aiModel == model ? const Icon(Icons.check) : null,
            onTap: () {
              ref.read(settingsProvider.notifier).updateAiModel(model);
              Navigator.pop(context);
            },
          )),
          const SizedBox(height: 16),
        ],
      ),
    );
  }

  void _showApiKeyDialog(BuildContext context, WidgetRef ref) {
    final controller = TextEditingController(text: ref.read(settingsProvider).apiKey);
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('API 密钥'),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(
            labelText: '输入 API 密钥',
            hintText: 'sk-...',
          ),
          obscureText: true,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          ElevatedButton(
            onPressed: () {
              ref.read(settingsProvider.notifier).updateApiKey(controller.text);
              Navigator.pop(context);
            },
            child: const Text('保存'),
          ),
        ],
      ),
    );
  }

  void _showEndpointDialog(BuildContext context, WidgetRef ref) {
    final controller = TextEditingController(text: ref.read(settingsProvider).apiEndpoint);
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('API 端点'),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(
            labelText: '输入 API 端点',
            hintText: 'https://api.example.com/v1',
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          ElevatedButton(
            onPressed: () {
              ref.read(settingsProvider.notifier).updateApiEndpoint(controller.text);
              Navigator.pop(context);
            },
            child: const Text('保存'),
          ),
        ],
      ),
    );
  }

  void _showWakeWordDialog(BuildContext context, WidgetRef ref) {
    final controller = TextEditingController(text: ref.read(settingsProvider).wakeWord);
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('唤醒词'),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(
            labelText: '唤醒词',
            hintText: 'Hey Tortoise',
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          ElevatedButton(
            onPressed: () {
              ref.read(settingsProvider.notifier).updateWakeWord(controller.text);
              Navigator.pop(context);
            },
            child: const Text('保存'),
          ),
        ],
      ),
    );
  }

  void _showThemeSelector(BuildContext context, WidgetRef ref) {
    showModalBottomSheet(
      context: context,
      builder: (context) => Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          ListTile(
            leading: const Icon(Icons.brightness_auto),
            title: const Text('跟随系统'),
            onTap: () {
              ref.read(settingsProvider.notifier).updateThemeMode(ThemeMode.system);
              Navigator.pop(context);
            },
          ),
          ListTile(
            leading: const Icon(Icons.light_mode),
            title: const Text('浅色模式'),
            onTap: () {
              ref.read(settingsProvider.notifier).updateThemeMode(ThemeMode.light);
              Navigator.pop(context);
            },
          ),
          ListTile(
            leading: const Icon(Icons.dark_mode),
            title: const Text('深色模式'),
            onTap: () {
              ref.read(settingsProvider.notifier).updateThemeMode(ThemeMode.dark);
              Navigator.pop(context);
            },
          ),
          const SizedBox(height: 16),
        ],
      ),
    );
  }

  void _showLanguageSelector(BuildContext context, WidgetRef ref) {
    showModalBottomSheet(
      context: context,
      builder: (context) => Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          ListTile(
            title: const Text('简体中文'),
            onTap: () {
              ref.read(settingsProvider.notifier).updateLanguage('zh-CN');
              Navigator.pop(context);
            },
          ),
          ListTile(
            title: const Text('繁體中文'),
            onTap: () {
              ref.read(settingsProvider.notifier).updateLanguage('zh-TW');
              Navigator.pop(context);
            },
          ),
          ListTile(
            title: const Text('English'),
            onTap: () {
              ref.read(settingsProvider.notifier).updateLanguage('en');
              Navigator.pop(context);
            },
          ),
          ListTile(
            title: const Text('日本語'),
            onTap: () {
              ref.read(settingsProvider.notifier).updateLanguage('ja');
              Navigator.pop(context);
            },
          ),
          const SizedBox(height: 16),
        ],
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  final String title;

  const _SectionHeader({required this.title});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
      child: Text(
        title,
        style: Theme.of(context).textTheme.titleSmall?.copyWith(
          color: Theme.of(context).colorScheme.primary,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }
}
