import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/chat_provider.dart';

class SessionsPage extends ConsumerWidget {
  const SessionsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sessions = ref.watch(sessionsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('会话列表'),
        actions: [
          IconButton(
            icon: const Icon(Icons.add),
            onPressed: () {
              ref.read(sessionsProvider.notifier).createSession();
            },
            tooltip: '新建会话',
          ),
        ],
      ),
      body: sessions.isEmpty
          ? _buildEmptyState(context)
          : ListView.builder(
              itemCount: sessions.length,
              itemBuilder: (context, index) {
                final session = sessions[index];
                return _buildSessionTile(context, ref, session);
              },
            ),
    );
  }

  Widget _buildEmptyState(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.chat_bubble_outline,
            size: 80,
            color: Theme.of(context).colorScheme.primary.withOpacity(0.3),
          ),
          const SizedBox(height: 16),
          Text(
            '还没有会话',
            style: Theme.of(context).textTheme.titleLarge?.copyWith(
                  color: Theme.of(context).colorScheme.onSurface.withOpacity(0.5),
                ),
          ),
        ],
      ),
    );
  }

  Widget _buildSessionTile(BuildContext context, WidgetRef ref, session) {
    return ListTile(
      leading: CircleAvatar(
        backgroundColor: Theme.of(context).colorScheme.primary.withOpacity(0.1),
        child: Icon(Icons.chat, color: Theme.of(context).colorScheme.primary),
      ),
      title: Text(session.title),
      subtitle: const Text(' 条消息'),
      trailing: PopupMenuButton<String>(
        itemBuilder: (context) => [
          const PopupMenuItem(value: 'rename', child: Text('重命名')),
          const PopupMenuItem(value: 'delete', child: Text('删除', style: TextStyle(color: Colors.red))),
        ],
        onSelected: (value) {
          if (value == 'delete') {
            ref.read(sessionsProvider.notifier).deleteSession(session.id);
          }
        },
      ),
      onTap: () {
        ref.read(currentSessionProvider.notifier).state = session.id;
      },
    );
  }
}

final currentSessionProvider = StateProvider<String?>((ref) => null);
