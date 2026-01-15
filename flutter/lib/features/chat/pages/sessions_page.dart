import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class SessionsPage extends ConsumerWidget {
  const SessionsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('会话列表'),
        actions: [
          IconButton(
            icon: const Icon(Icons.search),
            onPressed: () {
              // 搜索功能
            },
          ),
        ],
      ),
      body: ListView.builder(
        itemCount: 20,
        itemBuilder: (context, index) {
          return ListTile(
            leading: CircleAvatar(
              backgroundColor: Theme.of(context).colorScheme.primary.withOpacity(0.1),
              child: Icon(
                Icons.chat,
                color: Theme.of(context).colorScheme.primary,
              ),
            ),
            title: Text('会话 ${index + 1}'),
            subtitle: Text(
              '最后活动: ${_getTimeAgo(index)} • 消息数: ${(index + 1) * 3}',
            ),
            trailing: PopupMenuButton(
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
                  value: 'export',
                  child: Row(
                    children: [
                      Icon(Icons.download),
                      SizedBox(width: 8),
                      Text('导出'),
                    ],
                  ),
                ),
                const PopupMenuItem(
                  value: 'delete',
                  child: Row(
                    children: [
                      Icon(Icons.delete, color: Colors.red),
                      SizedBox(width: 8),
                      Text('删除', style: TextStyle(color: Colors.red)),
                    ],
                  ),
                ),
              ],
              onSelected: (value) {
                switch (value) {
                  case 'rename':
                    // 重命名逻辑
                    break;
                  case 'export':
                    // 导出逻辑
                    break;
                  case 'delete':
                    // 删除逻辑
                    break;
                }
              },
            ),
            onTap: () {
              // 打开会话
            },
          );
        },
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () {
          // 创建新会话
        },
        icon: const Icon(Icons.add),
        label: const Text('新建会话'),
      ),
    );
  }

  String _getTimeAgo(int index) {
    final times = ['刚刚', '5 分钟前', '15 分钟前', '30 分钟前', '1 小时前', '2 小时前', '昨天', '3 天前'];
    return times[index % times.length];
  }
}
