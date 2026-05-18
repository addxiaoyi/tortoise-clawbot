import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class ChannelsPage extends ConsumerStatefulWidget {
  const ChannelsPage({super.key});

  @override
  ConsumerState<ChannelsPage> createState() => _ChannelsPageState();
}

class _ChannelsPageState extends ConsumerState<ChannelsPage> {
  final Map<String, bool> _channels = {
    'Telegram': false,
    'Discord': false,
    'Slack': false,
    'WhatsApp': false,
    'Signal': false,
    'Matrix': false,
    'Email': false,
    'SMS': false,
  };

  final Map<String, IconData> _icons = {
    'Telegram': Icons.send,
    'Discord': Icons.games,
    'Slack': Icons.work,
    'WhatsApp': Icons.chat,
    'Signal': Icons.security,
    'Matrix': Icons.grid_on,
    'Email': Icons.email,
    'SMS': Icons.sms,
  };

  final Map<String, Color> _colors = {
    'Telegram': const Color(0xFF0088CC),
    'Discord': const Color(0xFF5865F2),
    'Slack': const Color(0xFF4A154B),
    'WhatsApp': const Color(0xFF25D366),
    'Signal': const Color(0xFF3A76F0),
    'Matrix': const Color(0xFF0DBD8B),
    'Email': const Color(0xFFEA4335),
    'SMS': const Color(0xFF6B7280),
  };

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text(
          'Channels',
          style: TextStyle(fontWeight: FontWeight.w600),
        ),
        elevation: 0,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Connected Channels',
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w500,
                color: Color(0xFF64748B),
              ),
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: _channels.entries
                  .where((e) => e.value)
                  .map((e) => Chip(
                        avatar: Icon(_icons[e.key], size: 18, color: _colors[e.key]),
                        label: Text(e.key),
                        backgroundColor: _colors[e.key]!.withValues(alpha: 0.1),
                      ))
                  .toList(),
            ),
            const SizedBox(height: 32),
            const Text(
              'Available Channels',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.w600,
                color: Color(0xFF1E293B),
              ),
            ),
            const SizedBox(height: 16),
            GridView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 2,
                crossAxisSpacing: 16,
                mainAxisSpacing: 16,
                childAspectRatio: 1.2,
              ),
              itemCount: _channels.length,
              itemBuilder: (context, index) {
                final entry = _channels.entries.elementAt(index);
                return _ChannelCard(
                  name: entry.key,
                  icon: _icons[entry.key]!,
                  color: _colors[entry.key]!,
                  isConnected: entry.value,
                  onTap: () => _showChannelConfig(entry.key),
                );
              },
            ),
          ],
        ),
      ),
    );
  }

  void _showChannelConfig(String channel) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) => _ChannelConfigSheet(
        channel: channel,
        icon: _icons[channel]!,
        color: _colors[channel]!,
      ),
    );
  }
}

class _ChannelCard extends StatelessWidget {
  final String name;
  final IconData icon;
  final Color color;
  final bool isConnected;
  final VoidCallback onTap;

  const _ChannelCard({
    required this.name,
    required this.icon,
    required this.color,
    required this.isConnected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: isConnected ? color.withValues(alpha: 0.3) : Colors.grey.shade200,
          ),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.04),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: color.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(icon, color: color, size: 28),
            ),
            const SizedBox(height: 12),
            Text(
              name,
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: Color(0xFF1E293B),
              ),
            ),
            const SizedBox(height: 4),
            Row(
              children: [
                Container(
                  width: 8,
                  height: 8,
                  decoration: BoxDecoration(
                    color: isConnected ? const Color(0xFF22C55E) : const Color(0xFF94A3B8),
                    shape: BoxShape.circle,
                  ),
                ),
                const SizedBox(width: 6),
                Text(
                  isConnected ? 'Connected' : 'Not connected',
                  style: TextStyle(
                    fontSize: 12,
                    color: isConnected ? const Color(0xFF22C55E) : const Color(0xFF94A3B8),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _ChannelConfigSheet extends StatelessWidget {
  final String channel;
  final IconData icon;
  final Color color;

  const _ChannelConfigSheet({
    required this.channel,
    required this.icon,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: color.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(icon, color: color, size: 24),
              ),
              const SizedBox(width: 16),
              Text(
                channel,
                style: const TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
          const SizedBox(height: 24),
          TextField(
            decoration: InputDecoration(
              labelText: 'API Token',
              hintText: 'Enter your $channel API token',
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            obscureText: true,
          ),
          const SizedBox(height: 16),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: () => Navigator.pop(context),
              style: ElevatedButton.styleFrom(
                backgroundColor: color,
                padding: const EdgeInsets.symmetric(vertical: 16),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
              child: const Text(
                'Connect',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
              ),
            ),
          ),
          const SizedBox(height: 16),
        ],
      ),
    );
  }
}
