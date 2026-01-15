import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/memory_provider.dart';

class MemoryPage extends ConsumerStatefulWidget {
  const MemoryPage({super.key});

  @override
  ConsumerState<MemoryPage> createState() => _MemoryPageState();
}

class _MemoryPageState extends ConsumerState<MemoryPage> {
  final _searchController = TextEditingController();

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(memoryProvider);
    final filteredMemories = ref.watch(filteredMemoriesProvider);
    final typeFilter = ref.watch(memoryTypeFilterProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('记忆'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.read(memoryProvider.notifier).refresh(),
          ),
        ],
      ),
      body: Column(
        children: [
          // Search Bar
          Padding(
            padding: const EdgeInsets.all(16),
            child: TextField(
              controller: _searchController,
              decoration: InputDecoration(
                hintText: '搜索记忆...',
                prefixIcon: const Icon(Icons.search),
                suffixIcon: _searchController.text.isNotEmpty
                    ? IconButton(
                        icon: const Icon(Icons.clear),
                        onPressed: () {
                          _searchController.clear();
                          ref.read(memoryProvider.notifier).search('');
                        },
                      )
                    : null,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
              onChanged: (value) {
                ref.read(memoryProvider.notifier).search(value);
              },
            ),
          ),
          
          // Type Filter Chips
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Row(
              children: [
                FilterChip(
                  label: const Text('全部'),
                  selected: typeFilter == null,
                  onSelected: (_) {
                    ref.read(memoryTypeFilterProvider.notifier).state = null;
                  },
                ),
                const SizedBox(width: 8),
                ...MemoryType.values.map((type) {
                  return Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: FilterChip(
                      label: Text(_getTypeName(type)),
                      selected: typeFilter == type,
                      onSelected: (_) {
                        ref.read(memoryTypeFilterProvider.notifier).state = type;
                      },
                    ),
                  );
                }),
              ],
            ),
          ),
          
          const SizedBox(height: 8),
          
          // Memory List
          Expanded(
            child: state.isLoading
                ? const Center(child: CircularProgressIndicator())
                : filteredMemories.isEmpty
                    ? _buildEmptyState()
                    : ListView.builder(
                        padding: const EdgeInsets.all(16),
                        itemCount: filteredMemories.length,
                        itemBuilder: (context, index) {
                          final memory = filteredMemories[index];
                          return _MemoryCard(
                            memory: memory,
                            onDelete: () => _deleteMemory(memory.id),
                          );
                        },
                      ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showAddDialog(context),
        child: const Icon(Icons.add),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.psychology_outlined,
            size: 80,
            color: Colors.grey[400],
          ),
          const SizedBox(height: 16),
          Text(
            '还没有记忆',
            style: TextStyle(
              fontSize: 18,
              color: Colors.grey[600],
            ),
          ),
          const SizedBox(height: 8),
          Text(
            '点击 + 添加第一条记忆',
            style: TextStyle(
              color: Colors.grey[500],
            ),
          ),
        ],
      ),
    );
  }

  String _getTypeName(MemoryType type) {
    switch (type) {
      case MemoryType.fact:
        return '事实';
      case MemoryType.preference:
        return '偏好';
      case MemoryType.interest:
        return '兴趣';
      case MemoryType.work:
        return '工作';
      case MemoryType.personal:
        return '个人';
    }
  }

  Future<void> _deleteMemory(String id) async {
    await ref.read(memoryProvider.notifier).deleteEntry(id);
  }

  Future<void> _showAddDialog(BuildContext context) async {
    final contentController = TextEditingController();
    MemoryType selectedType = MemoryType.fact;

    await showDialog(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) {
          return AlertDialog(
            title: const Text('添加记忆'),
            content: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextField(
                  controller: contentController,
                  decoration: const InputDecoration(
                    labelText: '内容',
                    hintText: '输入记忆内容...',
                  ),
                  maxLines: 3,
                ),
                const SizedBox(height: 16),
                DropdownButtonFormField<MemoryType>(
                  value: selectedType,
                  decoration: const InputDecoration(
                    labelText: '类型',
                  ),
                  items: MemoryType.values.map((type) {
                    return DropdownMenuItem(
                      value: type,
                      child: Text(_getTypeName(type)),
                    );
                  }).toList(),
                  onChanged: (value) {
                    if (value != null) {
                      setState(() => selectedType = value);
                    }
                  },
                ),
              ],
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(context),
                child: const Text('取消'),
              ),
              ElevatedButton(
                onPressed: () {
                  if (contentController.text.isNotEmpty) {
                    final entry = MemoryEntry(
                      id: DateTime.now().millisecondsSinceEpoch.toString(),
                      content: contentController.text,
                      type: selectedType,
                      createdAt: DateTime.now(),
                    );
                    ref.read(memoryProvider.notifier).addEntry(entry);
                    Navigator.pop(context);
                  }
                },
                child: const Text('添加'),
              ),
            ],
          );
        },
      ),
    );
  }
}

class _MemoryCard extends StatelessWidget {
  final MemoryEntry memory;
  final VoidCallback onDelete;

  const _MemoryCard({
    required this.memory,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 4,
                  ),
                  decoration: BoxDecoration(
                    color: _getTypeColor(memory.type).withOpacity(0.1),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    _getTypeName(memory.type),
                    style: TextStyle(
                      color: _getTypeColor(memory.type),
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                const Spacer(),
                if (memory.accessCount > 0)
                  Row(
                    children: [
                      const Icon(Icons.visibility, size: 16, color: Colors.grey),
                      const SizedBox(width: 4),
                      Text(
                        '${memory.accessCount}',
                        style: const TextStyle(color: Colors.grey),
                      ),
                    ],
                  ),
                IconButton(
                  icon: const Icon(Icons.delete_outline, color: Colors.red),
                  onPressed: onDelete,
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              memory.content,
              style: const TextStyle(fontSize: 16),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Icon(Icons.access_time, size: 14, color: Colors.grey[500]),
                const SizedBox(width: 4),
                Text(
                  _formatDate(memory.createdAt),
                  style: TextStyle(
                    color: Colors.grey[500],
                    fontSize: 12,
                  ),
                ),
                const Spacer(),
                // Importance indicator
                Row(
                  children: [
                    Text(
                      '重要性: ',
                      style: TextStyle(
                        color: Colors.grey[500],
                        fontSize: 12,
                      ),
                    ),
                    ...List.generate(5, (index) {
                      return Icon(
                        index < (memory.importance * 5).round()
                            ? Icons.star
                            : Icons.star_border,
                        size: 14,
                        color: Colors.amber,
                      );
                    }),
                  ],
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Color _getTypeColor(MemoryType type) {
    switch (type) {
      case MemoryType.fact:
        return Colors.blue;
      case MemoryType.preference:
        return Colors.purple;
      case MemoryType.interest:
        return Colors.orange;
      case MemoryType.work:
        return Colors.green;
      case MemoryType.personal:
        return Colors.pink;
    }
  }

  String _getTypeName(MemoryType type) {
    switch (type) {
      case MemoryType.fact:
        return '事实';
      case MemoryType.preference:
        return '偏好';
      case MemoryType.interest:
        return '兴趣';
      case MemoryType.work:
        return '工作';
      case MemoryType.personal:
        return '个人';
    }
  }

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final diff = now.difference(date);

    if (diff.inDays > 30) {
      return '${date.month}/${date.day}/${date.year}';
    } else if (diff.inDays > 0) {
      return '${diff.inDays} 天前';
    } else if (diff.inHours > 0) {
      return '${diff.inHours} 小时前';
    } else if (diff.inMinutes > 0) {
      return '${diff.inMinutes} 分钟前';
    } else {
      return '刚刚';
    }
  }
}
