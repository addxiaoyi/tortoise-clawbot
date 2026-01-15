import 'package:flutter/material.dart';

/// 快速统计组件
class QuickStatsWidget extends StatelessWidget {
  final int sessionsCount;
  final int messagesCount;
  final int channelsConnected;
  final int pluginsEnabled;

  const QuickStatsWidget({
    super.key,
    this.sessionsCount = 0,
    this.messagesCount = 0,
    this.channelsConnected = 0,
    this.pluginsEnabled = 0,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              '📊 统计概览',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 16),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _StatItem(
                  icon: Icons.chat_bubble_outline,
                  label: '会话',
                  value: sessionsCount.toString(),
                  color: Colors.blue,
                ),
                _StatItem(
                  icon: Icons.message_outlined,
                  label: '消息',
                  value: messagesCount.toString(),
                  color: Colors.green,
                ),
                _StatItem(
                  icon: Icons.cable_outlined,
                  label: '渠道',
                  value: channelsConnected.toString(),
                  color: Colors.orange,
                ),
                _StatItem(
                  icon: Icons.extension_outlined,
                  label: '插件',
                  value: pluginsEnabled.toString(),
                  color: Colors.purple,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _StatItem extends StatelessWidget {
  final IconData icon;
  final String label;
  final String value;
  final Color color;

  const _StatItem({
    required this.icon,
    required this.label,
    required this.value,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: color.withOpacity(0.1),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Icon(icon, color: color, size: 24),
        ),
        const SizedBox(height: 8),
        Text(
          value,
          style: Theme.of(context).textTheme.titleLarge?.copyWith(
                fontWeight: FontWeight.bold,
              ),
        ),
        Text(
          label,
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
              ),
        ),
      ],
    );
  }
}
