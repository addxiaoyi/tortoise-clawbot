import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/config/api_config.dart';

class ChannelsPage extends ConsumerStatefulWidget {
  const ChannelsPage({super.key});

  @override
  ConsumerState<ChannelsPage> createState() => _ChannelsPageState();
}

class _ChannelsPageState extends ConsumerState<ChannelsPage> {
  final Map<String, bool> _enabledChannels = {
    Channels.telegram: false,
    Channels.discord: false,
    Channels.slack: false,
    Channels.whatsapp: false,
    Channels.matrix: false,
    Channels.email: false,
    Channels.web: true,
  };

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('消息渠道'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text(
            '已启用的渠道',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: _enabledChannels.entries
                .where((e) => e.value)
                .map((e) => Chip(
                      avatar: Icon(_getChannelIcon(e.key), size: 18),
                      label: Text(Channels.names[e.key] ?? e.key),
                      backgroundColor: Theme.of(context).colorScheme.primary.withOpacity(0.1),
                    ))
                .toList(),
          ),
          const SizedBox(height: 24),
          Text(
            '所有渠道',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          Card(
            child: Column(
              children: Channels.all.map((channel) {
                return _ChannelTile(
                  channel: channel,
                  isEnabled: _enabledChannels[channel] ?? false,
                  onToggle: (value) {
                    setState(() {
                      _enabledChannels[channel] = value ?? false;
                    });
                  },
                  onConfigure: () => _showChannelConfig(context, channel),
                );
              }).toList(),
            ),
          ),
        ],
      ),
    );
  }

  IconData _getChannelIcon(String channel) {
    switch (channel) {
      case Channels.telegram:
        return Icons.send;
      case Channels.discord:
        return Icons.headset;
      case Channels.slack:
        return Icons.tag;
      case Channels.whatsapp:
        return Icons.chat;
      case Channels.matrix:
        return Icons.grid_view;
      case Channels.email:
        return Icons.email;
      case Channels.sms:
        return Icons.sms;
      case Channels.web:
        return Icons.language;
      default:
        return Icons.cable;
    }
  }

  void _showChannelConfig(BuildContext context, String channel) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) => _ChannelConfigSheet(channel: channel),
    );
  }
}

class _ChannelTile extends StatelessWidget {
  final String channel;
  final bool isEnabled;
  final Function(bool?) onToggle;
  final VoidCallback onConfigure;

  const _ChannelTile({
    required this.channel,
    required this.isEnabled,
    required this.onToggle,
    required this.onConfigure,
  });

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Icon(_getChannelIcon(channel)),
      title: Text(Channels.names[channel] ?? channel),
      subtitle: Text(isEnabled ? '已连接' : '未连接'),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Switch(
            value: isEnabled,
            onChanged: onToggle,
          ),
          IconButton(
            icon: const Icon(Icons.settings),
            onPressed: onConfigure,
          ),
        ],
      ),
    );
  }

  IconData _getChannelIcon(String channel) {
    switch (channel) {
      case 'telegram':
        return Icons.send;
      case 'discord':
        return Icons.headset;
      case 'slack':
        return Icons.tag;
      case 'whatsapp':
        return Icons.chat;
      case 'matrix':
        return Icons.grid_view;
      case 'email':
        return Icons.email;
      case 'sms':
        return Icons.sms;
      case 'web':
        return Icons.language;
      default:
        return Icons.cable;
    }
  }
}

class _ChannelConfigSheet extends StatefulWidget {
  final String channel;

  const _ChannelConfigSheet({required this.channel});

  @override
  State<_ChannelConfigSheet> createState() => _ChannelConfigSheetState();
}

class _ChannelConfigSheetState extends State<_ChannelConfigSheet> {
  final _tokenController = TextEditingController();
  final _webhookController = TextEditingController();

  @override
  void dispose() {
    _tokenController.dispose();
    _webhookController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.only(
        bottom: MediaQuery.of(context).viewInsets.bottom,
      ),
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  _getChannelIcon(widget.channel),
                  size: 32,
                  color: Theme.of(context).colorScheme.primary,
                ),
                const SizedBox(width: 12),
                Text(
                  '${Channels.names[widget.channel] ?? widget.channel} 配置',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
              ],
            ),
            const SizedBox(height: 24),
            if (_needsBotToken(widget.channel)) ...[
              TextField(
                controller: _tokenController,
                decoration: const InputDecoration(
                  labelText: 'Bot Token',
                  hintText: '输入 Bot Token',
                  prefixIcon: Icon(Icons.key),
                ),
                obscureText: true,
              ),
              const SizedBox(height: 16),
            ],
            if (_needsWebhook(widget.channel)) ...[
              TextField(
                controller: _webhookController,
                decoration: const InputDecoration(
                  labelText: 'Webhook URL',
                  hintText: '输入 Webhook URL',
                  prefixIcon: Icon(Icons.link),
                ),
              ),
              const SizedBox(height: 16),
            ],
            if (_needsOtherConfig(widget.channel)) ...[
              _buildChannelSpecificFields(),
              const SizedBox(height: 16),
            ],
            const SizedBox(height: 24),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: const Text('取消'),
                ),
                const SizedBox(width: 12),
                ElevatedButton(
                  onPressed: () {
                    // 保存配置
                    Navigator.pop(context);
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('配置已保存')),
                    );
                  },
                  child: const Text('保存'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  bool _needsBotToken(String channel) {
    return ['telegram', 'discord', 'slack', 'whatsapp'].contains(channel);
  }

  bool _needsWebhook(String channel) {
    return ['slack', 'discord'].contains(channel);
  }

  bool _needsOtherConfig(String channel) {
    return ['email', 'sms'].contains(channel);
  }

  Widget _buildChannelSpecificFields() {
    return Column(
      children: [
        TextField(
          decoration: const InputDecoration(
            labelText: 'SMTP 服务器',
            hintText: 'smtp.example.com',
            prefixIcon: Icon(Icons.dns),
          ),
        ),
        const SizedBox(height: 12),
        TextField(
          decoration: const InputDecoration(
            labelText: '端口',
            hintText: '587',
            prefixIcon: Icon(Icons.numbers),
          ),
          keyboardType: TextInputType.number,
        ),
      ],
    );
  }

  IconData _getChannelIcon(String channel) {
    switch (channel) {
      case 'telegram':
        return Icons.send;
      case 'discord':
        return Icons.headset;
      case 'slack':
        return Icons.tag;
      case 'whatsapp':
        return Icons.chat;
      case 'matrix':
        return Icons.grid_view;
      case 'email':
        return Icons.email;
      case 'sms':
        return Icons.sms;
      case 'web':
        return Icons.language;
      default:
        return Icons.cable;
    }
  }
}
