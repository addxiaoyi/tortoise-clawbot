import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class PluginsPage extends ConsumerStatefulWidget {
  const PluginsPage({super.key});

  @override
  ConsumerState<PluginsPage> createState() => _PluginsPageState();
}

class _PluginsPageState extends ConsumerState<PluginsPage> {
  final List<_Plugin> _plugins = [
    _Plugin(
      id: 'web-search',
      name: '网络搜索',
      description: '使用搜索引擎获取实时信息',
      version: '1.0.0',
      author: 'Tortoise Team',
      icon: Icons.search,
      isEnabled: true,
      category: 'tools',
    ),
    _Plugin(
      id: 'code-interpreter',
      name: '代码解释器',
      description: '执行代码并返回结果',
      version: '1.0.0',
      author: 'Tortoise Team',
      icon: Icons.code,
      isEnabled: true,
      category: 'tools',
    ),
    _Plugin(
      id: 'file-manager',
      name: '文件管理',
      description: '读取和写入文件',
      version: '1.0.0',
      author: 'Tortoise Team',
      icon: Icons.folder,
      isEnabled: false,
      category: 'tools',
    ),
    _Plugin(
      id: 'github',
      name: 'GitHub 集成',
      description: '与 GitHub 仓库交互',
      version: '1.0.0',
      author: 'Community',
      icon: Icons.code_off,
      isEnabled: false,
      category: 'integration',
    ),
    _Plugin(
      id: 'notion',
      name: 'Notion 集成',
      description: '与 Notion 笔记同步',
      version: '1.0.0',
      author: 'Community',
      icon: Icons.note,
      isEnabled: false,
      category: 'integration',
    ),
  ];

  String _selectedCategory = 'all';

  @override
  Widget build(BuildContext context) {
    final filteredPlugins = _selectedCategory == 'all'
        ? _plugins
        : _plugins.where((p) => p.category == _selectedCategory).toList();

    return Scaffold(
      appBar: AppBar(
        title: const Text('插件'),
        actions: [
          IconButton(
            icon: const Icon(Icons.store),
            onPressed: () => _showPluginStore(context),
            tooltip: '插件商店',
          ),
        ],
      ),
      body: Column(
        children: [
          _buildCategoryFilter(),
          Expanded(
            child: _buildPluginGrid(filteredPlugins),
          ),
        ],
      ),
    );
  }

  Widget _buildCategoryFilter() {
    final categories = ['all', 'tools', 'integration', 'ai', 'custom'];
    final categoryNames = {
      'all': '全部',
      'tools': '工具',
      'integration': '集成',
      'ai': 'AI',
      'custom': '自定义',
    };

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: categories.map((category) {
            final isSelected = _selectedCategory == category;
            return Padding(
              padding: const EdgeInsets.only(right: 8),
              child: FilterChip(
                label: Text(categoryNames[category] ?? category),
                selected: isSelected,
                onSelected: (selected) {
                  setState(() => _selectedCategory = category);
                },
              ),
            );
          }).toList(),
        ),
      ),
    );
  }

  Widget _buildPluginGrid(List<_Plugin> plugins) {
    if (plugins.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.extension_off,
              size: 64,
              color: Theme.of(context).colorScheme.primary.withOpacity(0.5),
            ),
            const SizedBox(height: 16),
            Text(
              '没有找到插件',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            ElevatedButton.icon(
              onPressed: () => _showPluginStore(context),
              icon: const Icon(Icons.store),
              label: const Text('浏览插件商店'),
            ),
          ],
        ),
      );
    }

    return GridView.builder(
      padding: const EdgeInsets.all(16),
      gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(
        maxCrossAxisExtent: 300,
        mainAxisSpacing: 16,
        crossAxisSpacing: 16,
        childAspectRatio: 1.2,
      ),
      itemCount: plugins.length,
      itemBuilder: (context, index) {
        final plugin = plugins[index];
        return _PluginCard(
          plugin: plugin,
          onToggle: () {
            setState(() {
              plugin.isEnabled = !plugin.isEnabled;
            });
          },
          onConfigure: () => _showPluginConfig(context, plugin),
        );
      },
    );
  }

  void _showPluginStore(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.9,
        minChildSize: 0.5,
        maxChildSize: 0.95,
        expand: false,
        builder: (context, scrollController) {
          return _PluginStoreSheet(scrollController: scrollController);
        },
      ),
    );
  }

  void _showPluginConfig(BuildContext context, _Plugin plugin) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Row(
          children: [
            Icon(plugin.icon, color: Theme.of(context).colorScheme.primary),
            const SizedBox(width: 12),
            Text(plugin.name),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _InfoRow(label: '版本', value: plugin.version),
            _InfoRow(label: '作者', value: plugin.author),
            _InfoRow(label: '分类', value: plugin.category),
            const SizedBox(height: 16),
            Text(
              plugin.description,
              style: Theme.of(context).textTheme.bodyMedium,
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('关闭'),
          ),
        ],
      ),
    );
  }
}

class _Plugin {
  final String id;
  final String name;
  final String description;
  final String version;
  final String author;
  final IconData icon;
  bool isEnabled;
  final String category;

  _Plugin({
    required this.id,
    required this.name,
    required this.description,
    required this.version,
    required this.author,
    required this.icon,
    required this.isEnabled,
    required this.category,
  });
}

class _PluginCard extends StatelessWidget {
  final _Plugin plugin;
  final VoidCallback onToggle;
  final VoidCallback onConfigure;

  const _PluginCard({
    required this.plugin,
    required this.onToggle,
    required this.onConfigure,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      child: InkWell(
        onTap: onConfigure,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: Theme.of(context).colorScheme.primary.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Icon(
                      plugin.icon,
                      color: Theme.of(context).colorScheme.primary,
                    ),
                  ),
                  const Spacer(),
                  Switch(
                    value: plugin.isEnabled,
                    onChanged: (_) => onToggle(),
                  ),
                ],
              ),
              const Spacer(),
              Text(
                plugin.name,
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              const SizedBox(height: 4),
              Text(
                plugin.description,
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
                ),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
              const SizedBox(height: 8),
              Text(
                'v${plugin.version}',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Theme.of(context).colorScheme.primary,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _PluginStoreSheet extends StatelessWidget {
  final ScrollController scrollController;

  const _PluginStoreSheet({required this.scrollController});

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Theme.of(context).scaffoldBackgroundColor,
        borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
      ),
      child: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                const Text(
                  '插件商店',
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const Spacer(),
                IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: () => Navigator.pop(context),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          Expanded(
            child: ListView(
              controller: scrollController,
              padding: const EdgeInsets.all(16),
              children: [
                _buildFeaturedSection(context),
                const SizedBox(height: 24),
                _buildCategorySection(context, '工具', Icons.build),
                const SizedBox(height: 24),
                _buildCategorySection(context, '集成', Icons.integration_instructions),
                const SizedBox(height: 24),
                _buildCategorySection(context, 'AI', Icons.psychology),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFeaturedSection(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          '精选插件',
          style: Theme.of(context).textTheme.titleLarge?.copyWith(
            fontWeight: FontWeight.bold,
          ),
        ),
        const SizedBox(height: 12),
        SizedBox(
          height: 150,
          child: ListView(
            scrollDirection: Axis.horizontal,
            children: [
              _FeaturedCard(
                icon: Icons.image,
                name: 'DALL-E 图像生成',
                description: '使用 AI 生成图像',
                onInstall: () {},
              ),
              _FeaturedCard(
                icon: Icons.translate,
                name: '翻译助手',
                description: '多语言翻译支持',
                onInstall: () {},
              ),
              _FeaturedCard(
                icon: Icons.calculate,
                name: '数学计算',
                description: '高级数学计算',
                onInstall: () {},
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildCategorySection(BuildContext context, String title, IconData icon) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(icon, size: 20),
            const SizedBox(width: 8),
            Text(
              title,
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.bold,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        ...List.generate(3, (index) {
          return ListTile(
            leading: CircleAvatar(
              child: Icon(icon),
            ),
            title: Text('插件 ${index + 1}'),
            subtitle: const Text('插件描述...'),
            trailing: OutlinedButton(
              onPressed: () {},
              child: const Text('安装'),
            ),
          );
        }),
      ],
    );
  }
}

class _FeaturedCard extends StatelessWidget {
  final IconData icon;
  final String name;
  final String description;
  final VoidCallback onInstall;

  const _FeaturedCard({
    required this.icon,
    required this.name,
    required this.description,
    required this.onInstall,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 200,
      margin: const EdgeInsets.only(right: 12),
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Icon(icon, size: 32, color: Theme.of(context).colorScheme.primary),
              const SizedBox(height: 12),
              Text(
                name,
                style: Theme.of(context).textTheme.titleSmall?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              Text(
                description,
                style: Theme.of(context).textTheme.bodySmall,
                maxLines: 2,
              ),
              const Spacer(),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: onInstall,
                  child: const Text('安装'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _InfoRow extends StatelessWidget {
  final String label;
  final String value;

  const _InfoRow({required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        children: [
          Text(
            '$label: ',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          Text(value),
        ],
      ),
    );
  }
}
