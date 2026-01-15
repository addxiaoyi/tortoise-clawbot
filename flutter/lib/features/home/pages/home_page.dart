import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../chat/providers/chat_provider.dart';
import '../../chat/models/chat_models.dart';

class HomePage extends ConsumerWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sessions = ref.watch(sessionsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Tortoise'),
      ),
      body: ListView.builder(
        itemCount: sessions.length,
        itemBuilder: (context, index) {
          final session = sessions[index];
          return _SessionTile(session: session);
        },
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () {
          ref.read(sessionsProvider.notifier).createSession();
        },
        child: const Icon(Icons.add),
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
    
    return ListTile(
      title: Text(session.title),
      subtitle: Text('$messageCount 条消息'),
      trailing: Text(
        _formatDate(session.updatedAt),
        style: Theme.of(context).textTheme.bodySmall,
      ),
      onTap: () {
        ref.read(activeSessionIdProvider.notifier).state = session.id;
        Navigator.of(context).pushNamed('/chat');
      },
    );
  }

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final diff = now.difference(date);
    
    if (diff.inDays == 0) {
      return '今天';
    } else if (diff.inDays == 1) {
      return '昨天';
    } else if (diff.inDays < 7) {
      return '${diff.inDays}天前';
    } else {
      return '${date.month}/${date.day}';
    }
  }
}
