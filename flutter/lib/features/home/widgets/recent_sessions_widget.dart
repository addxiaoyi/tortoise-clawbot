import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// 最近会话列表组件
class RecentSessionsWidget extends StatelessWidget {
  final List<RecentSession> sessions;

  const RecentSessionsWidget({
    super.key,
    this.sessions = const [],
  });

  @override
  Widget build(BuildContext context) {
    final displaySessions = sessions.isEmpty
        ? _defaultSessions
        : sessions;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  '💬 最近会话',
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                ),
                TextButton(
                  onPressed: () => context.go('/sessions'),
                  child: const Text('查看全部'),
                ),
              ],
            ),
            const SizedBox(height: 12),
            if (displaySessions.isEmpty)
              const Center(
                child: Padding(
                  padding: EdgeInsets.all(20),
                  child: Text(
                    '暂无会话记录',
                    style: TextStyle(color: Colors.grey),
                  ),
                ),
              )
            else
              ...displaySessions.take(5).map(
                    (session) => _SessionTile(
                      session: session,
                      onTap: () => context.go('/chat/${session.id}'),
                    ),
                  ),
          ],
        ),
      ),
    );
  }

  static final List<RecentSession> _defaultSessions = [
    RecentSession(
      id: '1',
      title: '关于AI助手开发的讨论',
      lastMessage: '今天天气怎么样？',
      updatedAt: '5 分钟前',
      messageCount: 24,
    ),
    RecentSession(
      id: '2',
      title: 'Flutter 项目代码审查',
      lastMessage: '这个组件需要优化...',
      updatedAt: '30 分钟前',
      messageCount: 18,
    ),
    RecentSession(
      id: '3',
      title: 'Rust 异步编程学习',
      lastMessage: 'Tokio 的工作原理是什么？',
      updatedAt: '1 小时前',
      messageCount: 42,
    ),
  ];
}

class RecentSession {
  final String id;
  final String title;
  final String lastMessage;
  final String updatedAt;
  final int messageCount;

  const RecentSession({
    required this.id,
    required this.title,
    required this.lastMessage,
    required this.updatedAt,
    required this.messageCount,
  });
}

class _SessionTile extends StatelessWidget {
  final RecentSession session;
  final VoidCallback onTap;

  const _SessionTile({
    required this.session,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return ListTile(
      contentPadding: EdgeInsets.zero,
      leading: CircleAvatar(
        backgroundColor: Theme.of(context).colorScheme.primaryContainer,
        child: Icon(
          Icons.chat_bubble_outline,
          color: Theme.of(context).colorScheme.primary,
        ),
      ),
      title: Text(
        session.title,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
      ),
      subtitle: Text(
        session.lastMessage,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: TextStyle(
          color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
        ),
      ),
      trailing: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          Text(
            session.updatedAt,
            style: const TextStyle(fontSize: 12, color: Colors.grey),
          ),
          const SizedBox(height: 4),
          Text(
            '${session.messageCount} 条消息',
            style: TextStyle(
              fontSize: 11,
              color: Theme.of(context).colorScheme.onSurface.withOpacity(0.5),
            ),
          ),
        ],
      ),
      onTap: onTap,
    );
  }
}
