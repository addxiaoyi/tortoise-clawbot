import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// 快捷操作按钮组件
class QuickActionsWidget extends StatelessWidget {
  const QuickActionsWidget({super.key});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              '⚡ 快捷操作',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 16),
            Wrap(
              spacing: 12,
              runSpacing: 12,
              children: [
                _ActionButton(
                  icon: Icons.add_comment,
                  label: '新会话',
                  color: Colors.teal,
                  onTap: () => context.go('/chat'),
                ),
                _ActionButton(
                  icon: Icons.cable,
                  label: '渠道管理',
                  color: Colors.blue,
                  onTap: () => context.go('/channels'),
                ),
                _ActionButton(
                  icon: Icons.extension,
                  label: '插件中心',
                  color: Colors.purple,
                  onTap: () => context.go('/plugins'),
                ),
                _ActionButton(
                  icon: Icons.psychology,
                  label: '记忆库',
                  color: Colors.orange,
                  onTap: () => context.go('/memory'),
                ),
                _ActionButton(
                  icon: Icons.settings,
                  label: '设置',
                  color: Colors.grey,
                  onTap: () => context.go('/settings'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _ActionButton extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color color;
  final VoidCallback onTap;

  const _ActionButton({
    required this.icon,
    required this.label,
    required this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Material(
      color: color.withOpacity(0.1),
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(icon, color: color, size: 20),
              const SizedBox(width: 8),
              Text(
                label,
                style: TextStyle(
                  color: color,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
