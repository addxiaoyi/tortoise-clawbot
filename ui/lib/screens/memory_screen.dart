import 'package:flutter/material.dart';

class MemoryScreen extends StatelessWidget {
  const MemoryScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Memory'),
        actions: [
          IconButton(
            icon: const Icon(Icons.search),
            onPressed: () {},
          ),
        ],
      ),
      body: ListView(
        children: [
          // Memory Overview
          Padding(
            padding: const EdgeInsets.all(16),
            child: _MemoryOverviewCard(),
          ),
          
          // Memory Types
          const Padding(
            padding: EdgeInsets.symmetric(horizontal: 16),
            child: Text(
              'Memory Types',
              style: TextStyle(
                fontWeight: FontWeight.bold,
                fontSize: 16,
              ),
            ),
          ),
          const SizedBox(height: 8),
          _MemoryTypeTile(
            title: 'Short Term',
            description: 'Recent conversations and temporary data',
            count: 45,
            icon: Icons.short_text,
            color: Colors.blue,
          ),
          _MemoryTypeTile(
            title: 'Medium Term',
            description: 'Important information and learned patterns',
            count: 128,
            icon: Icons.memory,
            color: Colors.orange,
          ),
          _MemoryTypeTile(
            title: 'Long Term',
            description: 'Persistent knowledge and facts',
            count: 1024,
            icon: Icons.storage,
            color: Colors.green,
          ),
          
          const Divider(),
          
          // Recent Memories
          const Padding(
            padding: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: Text(
              'Recent Memories',
              style: TextStyle(
                fontWeight: FontWeight.bold,
                fontSize: 16,
              ),
            ),
          ),
          _MemoryItemTile(
            title: 'User preferences',
            preview: 'Dark mode enabled, language set to English...',
            timestamp: '2 hours ago',
          ),
          _MemoryItemTile(
            title: 'Project details',
            preview: 'Working on Tortoise AI Agent Framework...',
            timestamp: '5 hours ago',
          ),
          _MemoryItemTile(
            title: 'Code patterns',
            preview: 'React components use functional components...',
            timestamp: '1 day ago',
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () {
          _showAddMemoryDialog(context);
        },
        child: const Icon(Icons.add),
      ),
    );
  }

  void _showAddMemoryDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Add Memory'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              decoration: const InputDecoration(
                labelText: 'Content',
                border: OutlineInputBorder(),
              ),
              maxLines: 3,
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                const Text('Importance:'),
                Expanded(
                  child: Slider(
                    value: 0.5,
                    onChanged: (value) {},
                  ),
                ),
              ],
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Add'),
          ),
        ],
      ),
    );
  }
}

class _MemoryOverviewCard extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  'Memory Usage',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 18,
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.refresh),
                  onPressed: () {},
                ),
              ],
            ),
            const SizedBox(height: 16),
            ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: LinearProgressIndicator(
                value: 0.67,
                minHeight: 12,
                backgroundColor: Theme.of(context).colorScheme.surfaceContainerHighest,
              ),
            ),
            const SizedBox(height: 8),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  '67% used',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
                Text(
                  '67 / 100 GB',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _MemoryTypeTile extends StatelessWidget {
  final String title;
  final String description;
  final int count;
  final IconData icon;
  final Color color;

  const _MemoryTypeTile({
    required this.title,
    required this.description,
    required this.count,
    required this.icon,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: color.withOpacity(0.1),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Icon(icon, color: color),
      ),
      title: Text(title),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(description),
          Text(
            '$count items',
            style: Theme.of(context).textTheme.bodySmall,
          ),
        ],
      ),
      trailing: const Icon(Icons.chevron_right),
      onTap: () {},
    );
  }
}

class _MemoryItemTile extends StatelessWidget {
  final String title;
  final String preview;
  final String timestamp;

  const _MemoryItemTile({
    required this.title,
    required this.preview,
    required this.timestamp,
  });

  @override
  Widget build(BuildContext context) {
    return ListTile(
      title: Text(title),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            preview,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
          Text(
            timestamp,
            style: Theme.of(context).textTheme.bodySmall,
          ),
        ],
      ),
      trailing: PopupMenuButton(
        itemBuilder: (context) => [
          const PopupMenuItem(
            value: 'edit',
            child: Row(
              children: [
                Icon(Icons.edit),
                SizedBox(width: 8),
                Text('Edit'),
              ],
            ),
          ),
          const PopupMenuItem(
            value: 'delete',
            child: Row(
              children: [
                Icon(Icons.delete),
                SizedBox(width: 8),
                Text('Delete'),
              ],
            ),
          ),
        ],
        onSelected: (value) {
          // Handle menu actions
        },
      ),
      onTap: () {},
    );
  }
}
