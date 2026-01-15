import 'package:flutter/material.dart';

class ChannelsScreen extends StatelessWidget {
  const ChannelsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Channels'),
        actions: [
          IconButton(
            icon: const Icon(Icons.add),
            onPressed: () {},
          ),
        ],
      ),
      body: ListView(
        children: [
          _ChannelTile(
            icon: Icons.discord,
            name: 'Discord',
            status: 'Connected',
            color: const Color(0xFF5865F2),
          ),
          _ChannelTile(
            icon: Icons.telegram,
            name: 'Telegram',
            status: 'Connected',
            color: const Color(0xFF0088CC),
          ),
          _ChannelTile(
            icon: Icons.whatsapp,
            name: 'WhatsApp',
            status: 'Disconnected',
            color: const Color(0xFF25D366),
          ),
          _ChannelTile(
            icon: Icons.slack,
            name: 'Slack',
            status: 'Disconnected',
            color: const Color(0xFF4A154B),
          ),
          _ChannelTile(
            icon: Icons.email,
            name: 'Email',
            status: 'Disconnected',
            color: const Color(0xFFEA4335),
          ),
        ],
      ),
    );
  }
}

class _ChannelTile extends StatelessWidget {
  final IconData icon;
  final String name;
  final String status;
  final Color color;

  const _ChannelTile({
    required this.icon,
    required this.name,
    required this.status,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    final isConnected = status == 'Connected';

    return ListTile(
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: color.withOpacity(0.1),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Icon(icon, color: color),
      ),
      title: Text(name),
      subtitle: Text(status),
      trailing: Switch(
        value: isConnected,
        onChanged: (value) {},
      ),
      onTap: () {},
    );
  }
}
