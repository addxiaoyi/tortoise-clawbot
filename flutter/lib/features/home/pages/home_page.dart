import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../chat/providers/chat_provider.dart';
import '../../chat/models/chat_models.dart';

class HomePage extends ConsumerWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sessions = ref.watch(sessionsProvider);
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Tortoise'),
        actions: [
          IconButton(
            icon: const Icon(Icons.search),
            onPressed: () {
              // TODO: Global search
            },
          ),
          IconButton(
            icon: const Icon(Icons.notifications_outlined),
            onPressed: () {
              // TODO: Notifications
            },
          ),
        ],
      ),
      body: CustomScrollView(
        slivers: [
          // Quick Actions
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '快速开始',
                    style: theme.textTheme.titleLarge,
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: _QuickActionCard(
                          icon: Icons.chat,
                          title: '新对话',
                          color: Colors.blue,
                          onTap: () {
                            ref.read(sessionsProvider.notifier).createSession();
                            context.go('/chat');
                          },
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: _QuickActionCard(
                          icon: Icons.mic,
                          title: '语音助手',
                          color: Colors.purple,
                          onTap: () {
                            context.go('/voice');
                          },
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: _QuickActionCard(
                          icon: Icons.cable,
                          title: '渠道管理',
                          color: Colors.green,
                          onTap: () {
                            context.go('/channels');
                          },
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: _QuickActionCard(
                          icon: Icons.extension,
                          title: '插件市场',
                          color: Colors.orange,
                          onTap: () {
                            context.go('/marketplace');
                          },
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          // Recent Sessions Header
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    '最近的对话',
                    style: theme.textTheme.titleLarge,
                  ),
                  TextButton(
                    onPressed: () {
                      context.go('/chat');
                    },
                    child: const Text('查看全部'),
                  ),
                ],
              ),
            ),
          ),
          // Recent Sessions List
          sessions.isEmpty
              ? SliverToBoxAdapter(
                  child: Padding(
                    padding: const EdgeInsets.all(32),
                    child: Column(
                      children: [
                        Icon(
                          Icons.chat_bubble_outline,
                          size: 64,
                          color: theme.colorScheme.outline,
                        ),
                        const SizedBox(height: 16),
                        Text(
                          '还没有对话',
                          style: theme.textTheme.titleMedium?.copyWith(
                            color: theme.colorScheme.outline,
                          ),
                        ),
                        const SizedBox(height: 8),
                        ElevatedButton.icon(
                          onPressed: () {
                            ref.read(sessionsProvider.notifier).createSession();
                            context.go('/chat');
                          },
                          icon: const Icon(Icons.add),
                          label: const Text('开始对话'),
                        ),
                      ],
                    ),
                  ),
                )
              : SliverList(
                  delegate: SliverChildBuilderDelegate(
                    (context, index) {
                      if (index >= 5) return null;
                      final session = sessions[index];
                      return _SessionTile(session: session);
                    },
                    childCount: sessions.length > 5 ? 5 : sessions.length,
                  ),
                ),
          // Stats Section
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '统计信息',
                    style: theme.textTheme.titleLarge,
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: _StatCard(
                          title: '总对话数',
                          value: '${sessions.length}',
                          icon: Icons.chat_bubble,
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: _StatCard(
                          title: '活跃渠道',
                          value: '8',
                          icon: Icons.cable,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _QuickActionCard extends StatelessWidget {
  final IconData icon;
  final String title;
  final Color color;
  final VoidCallback onTap;

  const _QuickActionCard({
    required this.icon,
    required this.title,
    required this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            children: [
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: color.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(icon, color: color, size: 28),
              ),
              const SizedBox(height: 8),
              Text(
                title,
                style: Theme.of(context).textTheme.titleSmall,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _StatCard extends StatelessWidget {
  final String title;
  final String value;
  final IconData icon;

  const _StatCard({
    required this.title,
    required this.value,
    required this.icon,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            Icon(icon, color: theme.colorScheme.primary),
            const SizedBox(width: 12),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  value,
                  style: theme.textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                Text(
                  title,
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: theme.colorScheme.outline,
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

class _SessionTile extends ConsumerWidget {
  final ChatSession session;

  const _SessionTile({required this.session});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final messageCount = session.messages.length;
    final theme = Theme.of(context);

    return ListTile(
      leading: CircleAvatar(
        child: Text(session.title.isNotEmpty ? session.title[0] : '?'),
      ),
      title: Text(session.title.isEmpty ? '新对话' : session.title),
      subtitle: Text(
        '$messageCount 条消息 • ${_formatDate(session.updatedAt)}',
      ),
      trailing: PopupMenuButton<String>(
        onSelected: (value) {
          switch (value) {
            case 'delete':
              ref.read(sessionsProvider.notifier).deleteSession(session.id);
              break;
            case 'rename':
              _showRenameDialog(context, ref);
              break;
          }
        },
        itemBuilder: (context) => [
          const PopupMenuItem(
            value: 'rename',
            child: Row(
              children: [
                Icon(Icons.edit),
                SizedBox(width: 8),
                Text('重命名'),
              ],
            ),
          ),
          const PopupMenuItem(
            value: 'delete',
            child: Row(
              children: [
                Icon(Icons.delete),
                SizedBox(width: 8),
                Text('删除'),
              ],
            ),
          ),
        ],
      ),
      onTap: () {
        ref.read(activeSessionIdProvider.notifier).state = session.id;
        context.go('/chat');
      },
    );
  }

  void _showRenameDialog(BuildContext context, WidgetRef ref) {
    final controller = TextEditingController(text: session.title);
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('重命名对话'),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(
            labelText: '对话名称',
          ),
          autofocus: true,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          ElevatedButton(
            onPressed: () {
              if (controller.text.isNotEmpty) {
                ref.read(sessionsProvider.notifier).updateTitle(
                  session.id,
                  controller.text,
                );
              }
              Navigator.pop(context);
            },
            child: const Text('确定'),
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final diff = now.difference(date);

    if (diff.inDays == 0) {
      if (diff.inHours == 0) {
        return '${diff.inMinutes} 分钟前';
      }
      return '${diff.inHours} 小时前';
    } else if (diff.inDays == 1) {
      return '昨天';
    } else if (diff.inDays < 7) {
      return '${diff.inDays} 天前';
    } else {
      return '${date.month}/${date.day}';
    }
  }
}
