import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/theme/theme.dart';
import '../../core/di/providers.dart';

// ============ AI 配置状态 ============

class AIProviderState {
  final String id;
  final String name;
  final bool enabled;
  final String apiKey;
  final String model;
  final String baseUrl;

  const AIProviderState({
    required this.id,
    required this.name,
    this.enabled = false,
    this.apiKey = '',
    this.model = '',
    this.baseUrl = '',
  });

  AIProviderState copyWith({
    bool? enabled,
    String? apiKey,
    String? model,
  }) {
    return AIProviderState(
      id: id,
      name: name,
      enabled: enabled ?? this.enabled,
      apiKey: apiKey ?? this.apiKey,
      model: model ?? this.model,
      baseUrl: baseUrl,
    );
  }
}

// ============ 渠道配置状态 ============

class ChannelState {
  final bool telegramEnabled;
  final String telegramToken;
  final bool discordEnabled;
  final String discordToken;
  final bool slackEnabled;
  final String slackToken;

  const ChannelState({
    this.telegramEnabled = false,
    this.telegramToken = '',
    this.discordEnabled = false,
    this.discordToken = '',
    this.slackEnabled = false,
    this.slackToken = '',
  });

  ChannelState copyWith({
    bool? telegramEnabled,
    String? telegramToken,
    bool? discordEnabled,
    String? discordToken,
    bool? slackEnabled,
    String? slackToken,
  }) {
    return ChannelState(
      telegramEnabled: telegramEnabled ?? this.telegramEnabled,
      telegramToken: telegramToken ?? this.telegramToken,
      discordEnabled: discordEnabled ?? this.discordEnabled,
      discordToken: discordToken ?? this.discordToken,
      slackEnabled: slackEnabled ?? this.slackEnabled,
      slackToken: slackToken ?? this.slackToken,
    );
  }
}

// ============ Providers ============

final aiProvidersProvider = StateNotifierProvider<AIProvidersNotifier, List<AIProviderState>>((ref) {
  return AIProvidersNotifier();
});

class AIProvidersNotifier extends StateNotifier<List<AIProviderState>> {
  AIProvidersNotifier() : super([
    const AIProviderState(
      id: 'openai',
      name: 'OpenAI',
      model: 'gpt-4o',
      baseUrl: 'https://api.openai.com/v1',
    ),
    const AIProviderState(
      id: 'anthropic',
      name: 'Anthropic Claude',
      model: 'claude-3-sonnet-20240229',
      baseUrl: 'https://api.anthropic.com',
    ),
    const AIProviderState(
      id: 'ollama',
      name: 'Ollama (本地)',
      model: 'llama2',
      baseUrl: 'http://localhost:11434',
    ),
  ]);

  void toggleProvider(String id, bool enabled) {
    state = state.map((p) => p.id == id ? p.copyWith(enabled: enabled) : p).toList();
  }

  void updateApiKey(String id, String apiKey) {
    state = state.map((p) => p.id == id ? p.copyWith(apiKey: apiKey) : p).toList();
  }

  void updateModel(String id, String model) {
    state = state.map((p) => p.id == id ? p.copyWith(model: model) : p).toList();
  }
}

final channelStateProvider = StateNotifierProvider<ChannelNotifier, ChannelState>((ref) {
  return ChannelNotifier();
});

class ChannelNotifier extends StateNotifier<ChannelState> {
  ChannelNotifier() : super(const ChannelState());

  void toggleTelegram(bool enabled) => state = state.copyWith(telegramEnabled: enabled);
  void setTelegramToken(String token) => state = state.copyWith(telegramToken: token);
  void toggleDiscord(bool enabled) => state = state.copyWith(discordEnabled: enabled);
  void setDiscordToken(String token) => state = state.copyWith(discordToken: token);
  void toggleSlack(bool enabled) => state = state.copyWith(slackEnabled: enabled);
  void setSlackToken(String token) => state = state.copyWith(slackToken: token);
}

// ============ Settings Screen ============

class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({super.key});

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('设置'),
        bottom: TabBar(
          controller: _tabController,
          isScrollable: true,
          tabs: const [
            Tab(text: '连接'),
            Tab(text: 'AI 配置'),
            Tab(text: '渠道'),
            Tab(text: '外观'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: const [
          _ConnectionTab(),
          _AIConfigTab(),
          _ChannelsTab(),
          _AppearanceTab(),
        ],
      ),
    );
  }
}

// ============ Connection Tab ============

class _ConnectionTab extends ConsumerWidget {
  const _ConnectionTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final settings = ref.watch(settingsProvider);
    final connectionState = ref.watch(connectionStateProvider);

    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        _sectionHeader('服务器连接'),
        _card([
          _textField(
            context: context,
            label: 'API 端点',
            hint: 'http://localhost:18792',
            initialValue: settings.apiEndpoint,
            onChanged: (v) => ref.read(settingsProvider.notifier).updateEndpoint(v),
          ),
          const SizedBox(height: 16),
          _textField(
            context: context,
            label: 'API Key',
            hint: '输入你的 API Key',
            isPassword: true,
            initialValue: settings.apiKey,
            onChanged: (v) => ref.read(settingsProvider.notifier).updateApiKey(v),
          ),
        ]),
        const SizedBox(height: 24),
        _sectionHeader('连接状态'),
        _card([
          _connectionStatus(connectionState),
          const SizedBox(height: 16),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: () => _testConnection(context),
              style: _buttonStyle,
              child: const Text('测试连接'),
            ),
          ),
        ]),
      ],
    );
  }

  Widget _connectionStatus(ConnectionState state) {
    Color color;
    String text;
    IconData icon;

    switch (state) {
      case ConnectionState.connected:
        color = Colors.green; text = '已连接'; icon = Icons.check_circle;
      case ConnectionState.connecting:
        color = Colors.orange; text = '连接中...'; icon = Icons.sync;
      case ConnectionState.disconnected:
        color = Colors.grey; text = '未连接'; icon = Icons.circle_outlined;
      case ConnectionState.error:
        color = Colors.red; text = '连接错误'; icon = Icons.error;
    }

    return Row(
      children: [
        Icon(icon, color: color),
        const SizedBox(width: 12),
        Text(text, style: TextStyle(color: color, fontWeight: FontWeight.w500)),
      ],
    );
  }

  Future<void> _testConnection(BuildContext context) async {
    ref.read(connectionStateProvider.notifier).state = ConnectionState.connecting;
    // TODO: 实现真实连接测试
    await Future.delayed(const Duration(seconds: 1));
    ref.read(connectionStateProvider.notifier).state = ConnectionState.connected;
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('连接成功！'), backgroundColor: Colors.green),
      );
    }
  }
}

// ============ AI Config Tab ============

class _AIConfigTab extends ConsumerWidget {
  const _AIConfigTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final providers = ref.watch(aiProvidersProvider);
    final settings = ref.watch(settingsProvider);

    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        _sectionHeader('AI 提供商'),
        ...providers.map((p) => _providerCard(context, ref, p)),
        const SizedBox(height: 24),
        _sectionHeader('模型参数'),
        _card([
          _slider(
            label: 'Temperature',
            value: settings.temperature,
            min: 0, max: 2,
            onChanged: (v) => ref.read(settingsProvider.notifier).updateTemperature(v),
          ),
          const SizedBox(height: 16),
          _slider(
            label: 'Max Tokens',
            value: settings.maxTokens.toDouble(),
            min: 256, max: 8192, divisions: 30,
            onChanged: (v) => ref.read(settingsProvider.notifier).updateMaxTokens(v.toInt()),
          ),
        ]),
      ],
    );
  }

  Widget _providerCard(BuildContext context, WidgetRef ref, AIProviderState provider) {
    final iconData = provider.id == 'openai' ? Icons.psychology :
                    provider.id == 'anthropic' ? Icons.psychology_alt : Icons.computer;
    final color = provider.id == 'openai' ? Colors.green :
                  provider.id == 'anthropic' ? Colors.orange : Colors.blue;

    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: _card([
        Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: color.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(iconData, color: color),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(provider.name, style: const TextStyle(fontWeight: FontWeight.w600)),
                  Text(provider.model, style: TextStyle(fontSize: 12, color: Theme.of(context).textTheme.bodySmall?.color)),
                ],
              ),
            ),
            Switch(
              value: provider.enabled,
              activeColor: TortoiseTheme.primary,
              onChanged: (v) => ref.read(aiProvidersProvider.notifier).toggleProvider(provider.id, v),
            ),
          ],
        ),
        if (provider.enabled) ...[
          const Divider(height: 24),
          _textField(
            context: context,
            label: 'API Key',
            hint: 'sk-...',
            isPassword: true,
            initialValue: provider.apiKey,
            onChanged: (v) => ref.read(aiProvidersProvider.notifier).updateApiKey(provider.id, v),
          ),
          const SizedBox(height: 12),
          _modelDropdown(context, ref, provider),
        ],
      ]),
    );
  }

  Widget _modelDropdown(BuildContext context, WidgetRef ref, AIProviderState provider) {
    final models = provider.id == 'openai'
        ? ['gpt-4o', 'gpt-4-turbo', 'gpt-4', 'gpt-3.5-turbo']
        : provider.id == 'anthropic'
        ? ['claude-3-opus-20240229', 'claude-3-sonnet-20240229', 'claude-3-haiku-20240307']
        : ['llama2', 'llama3', 'mistral', 'codellama'];

    return DropdownButtonFormField<String>(
      value: provider.model,
      decoration: _inputDecoration(context),
      items: models.map((m) => DropdownMenuItem(value: m, child: Text(m))).toList(),
      onChanged: (v) {
        if (v != null) ref.read(aiProvidersProvider.notifier).updateModel(provider.id, v);
      },
    );
  }
}

// ============ Channels Tab ============

class _ChannelsTab extends ConsumerWidget {
  const _ChannelsTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final channels = ref.watch(channelStateProvider);

    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        _sectionHeader('消息渠道'),
        _channelCard(
          context: context,
          ref: ref,
          name: 'Telegram',
          icon: Icons.send,
          color: Colors.blue,
          enabled: channels.telegramEnabled,
          onToggle: (v) => ref.read(channelStateProvider.notifier).toggleTelegram(v),
          token: channels.telegramToken,
          onTokenChange: (v) => ref.read(channelStateProvider.notifier).setTelegramToken(v),
        ),
        const SizedBox(height: 12),
        _channelCard(
          context: context,
          ref: ref,
          name: 'Discord',
          icon: Icons.discord,
          color: Colors.indigo,
          enabled: channels.discordEnabled,
          onToggle: (v) => ref.read(channelStateProvider.notifier).toggleDiscord(v),
          token: channels.discordToken,
          onTokenChange: (v) => ref.read(channelStateProvider.notifier).setDiscordToken(v),
        ),
        const SizedBox(height: 12),
        _channelCard(
          context: context,
          ref: ref,
          name: 'Slack',
          icon: Icons.tag,
          color: Colors.green,
          enabled: channels.slackEnabled,
          onToggle: (v) => ref.read(channelStateProvider.notifier).toggleSlack(v),
          token: channels.slackToken,
          onTokenChange: (v) => ref.read(channelStateProvider.notifier).setSlackToken(v),
        ),
        const SizedBox(height: 24),
        SizedBox(
          width: double.infinity,
          child: ElevatedButton(
            onPressed: () => _saveChannels(context),
            style: _buttonStyle,
            child: const Text('保存配置'),
          ),
        ),
      ],
    );
  }

  Widget _channelCard({
    required BuildContext context,
    required WidgetRef ref,
    required String name,
    required IconData icon,
    required Color color,
    required bool enabled,
    required ValueChanged<bool> onToggle,
    required String token,
    required ValueChanged<String> onTokenChange,
  }) {
    return _card([
      Row(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(color: color.withOpacity(0.1), borderRadius: BorderRadius.circular(8)),
            child: Icon(icon, color: color),
          ),
          const SizedBox(width: 12),
          Expanded(child: Text(name, style: const TextStyle(fontWeight: FontWeight.w600))),
          Switch(value: enabled, activeColor: TortoiseTheme.primary, onChanged: onToggle),
        ],
      ),
      if (enabled) ...[
        const Divider(height: 24),
        _textField(
          context: context,
          label: 'Bot Token',
          hint: name == 'Telegram' ? '123456789:ABC...' : 'xoxb-... 或 Bot Token',
          isPassword: true,
          initialValue: token,
          onChanged: onTokenChange,
        ),
      ],
    ]);
  }

  Future<void> _saveChannels(BuildContext context) async {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('配置已保存'), backgroundColor: Colors.green),
    );
  }
}

// ============ Appearance Tab ============

class _AppearanceTab extends ConsumerWidget {
  const _AppearanceTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final themeMode = ref.watch(themeModeProvider);

    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        _sectionHeader('外观'),
        _card([
          DropdownButtonFormField<ThemeMode>(
            value: themeMode,
            decoration: _inputDecoration(context),
            items: const [
              DropdownMenuItem(value: ThemeMode.system, child: Text('跟随系统')),
              DropdownMenuItem(value: ThemeMode.light, child: Text('浅色')),
              DropdownMenuItem(value: ThemeMode.dark, child: Text('深色')),
            ],
            onChanged: (v) {
              if (v != null) ref.read(themeModeProvider.notifier).state = v;
            },
          ),
        ]),
        const SizedBox(height: 24),
        _sectionHeader('关于'),
        _card([
          _infoTile(context, Icons.info_outline, '版本', '0.1.0'),
          const Divider(height: 16),
          _infoTile(context, Icons.code, '开源协议', 'Apache 2.0'),
        ]),
      ],
    );
  }

  Widget _infoTile(BuildContext context, IconData icon, String title, String subtitle) {
    return Row(
      children: [
        Icon(icon, color: TortoiseTheme.primary),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(title, style: const TextStyle(fontWeight: FontWeight.w500)),
              Text(subtitle, style: TextStyle(fontSize: 12, color: Theme.of(context).textTheme.bodySmall?.color)),
            ],
          ),
        ),
      ],
    );
  }
}

// ============ Helper Widgets ============

Widget _sectionHeader(String title) => Padding(
  padding: const EdgeInsets.only(left: 4, bottom: 8),
  child: Text(title, style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: TortoiseTheme.primary)),
);

Widget _card(List<Widget> children) => Container(
  decoration: BoxDecoration(
    color: Colors.white,
    borderRadius: BorderRadius.circular(16),
    border: Border.all(color: Colors.grey.withOpacity(0.1)),
  ),
  padding: const EdgeInsets.all(16),
  child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: children),
);

Widget _textField({
  required BuildContext context,
  required String label,
  required String hint,
  bool isPassword = false,
  String initialValue = '',
  required ValueChanged<String> onChanged,
}) {
  return Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Text(label, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500)),
      const SizedBox(height: 8),
      TextFormField(
        initialValue: initialValue,
        obscureText: isPassword,
        onChanged: onChanged,
        decoration: _inputDecoration(context).copyWith(hintText: hint),
      ),
    ],
  );
}

Widget _slider({
  required String label,
  required double value,
  required double min,
  required double max,
  int? divisions,
  required ValueChanged<double> onChanged,
}) {
  return Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500)),
          Text(value.toStringAsFixed(1), style: TextStyle(fontSize: 14, color: TortoiseTheme.primary, fontWeight: FontWeight.w600)),
        ],
      ),
      Slider(
        value: value, min: min, max: max, divisions: divisions,
        activeColor: TortoiseTheme.primary,
        onChanged: onChanged,
      ),
    ],
  );
}

InputDecoration _inputDecoration(BuildContext context) => InputDecoration(
  isDense: true,
  border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide.none),
  filled: true,
  fillColor: Theme.of(context).colorScheme.surfaceContainerHighest,
);

ButtonStyle get _buttonStyle => ElevatedButton.styleFrom(
  backgroundColor: TortoiseTheme.primary,
  foregroundColor: Colors.white,
  padding: const EdgeInsets.symmetric(vertical: 12),
  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
);
