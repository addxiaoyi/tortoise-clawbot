import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/marketplace_provider.dart';

// Marketplace categories
enum MarketplaceCategory {
  all,
  channels,
  skills,
  themes,
  integrations,
}

// Provider for selected category
final selectedCategoryProvider = StateProvider<MarketplaceCategory>((ref) => MarketplaceCategory.all);

// Search query provider
final marketplaceSearchProvider = StateProvider<String>((ref) => '');

// Filtered plugins provider
final filteredPluginsProvider = Provider<List<PluginItem>>((ref) {
  final plugins = ref.watch(marketplacePluginsProvider);
  final category = ref.watch(selectedCategoryProvider);
  final search = ref.watch(marketplaceSearchProvider).toLowerCase();
  
  return plugins.where((plugin) {
    // Category filter
    if (category != MarketplaceCategory.all && plugin.category != category.name) {
      return false;
    }
    
    // Search filter
    if (search.isNotEmpty) {
      return plugin.name.toLowerCase().contains(search) ||
             plugin.description.toLowerCase().contains(search);
    }
    
    return true;
  }).toList();
});

class MarketplacePage extends ConsumerStatefulWidget {
  const MarketplacePage({super.key});

  @override
  ConsumerState<MarketplacePage> createState() => _MarketplacePageState();
}

class _MarketplacePageState extends ConsumerState<MarketplacePage> {
  final _searchController = TextEditingController();

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final plugins = ref.watch(filteredPluginsProvider);
    final selectedCategory = ref.watch(selectedCategoryProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('插件市场'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.refresh(marketplacePluginsProvider),
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
                hintText: '搜索插件...',
                prefixIcon: const Icon(Icons.search),
                suffixIcon: _searchController.text.isNotEmpty
                    ? IconButton(
                        icon: const Icon(Icons.clear),
                        onPressed: () {
                          _searchController.clear();
                          ref.read(marketplaceSearchProvider.notifier).state = '';
                        },
                      )
                    : null,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                filled: true,
              ),
              onChanged: (value) {
                ref.read(marketplaceSearchProvider.notifier).state = value;
              },
            ),
          ),
          
          // Category Chips
          SizedBox(
            height: 50,
            child: ListView(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.symmetric(horizontal: 16),
              children: MarketplaceCategory.values.map((category) {
                final isSelected = category == selectedCategory;
                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: FilterChip(
                    label: Text(_getCategoryLabel(category)),
                    selected: isSelected,
                    onSelected: (selected) {
                      ref.read(selectedCategoryProvider.notifier).state = category;
                    },
                    avatar: Icon(
                      _getCategoryIcon(category),
                      size: 18,
                    ),
                  ),
                );
              }).toList(),
            ),
          ),
          
          const SizedBox(height: 8),
          
          // Plugin Grid
          Expanded(
            child: plugins.isEmpty
                ? _buildEmptyState()
                : GridView.builder(
                    padding: const EdgeInsets.all(16),
                    gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                      crossAxisCount: 2,
                      childAspectRatio: 0.85,
                      crossAxisSpacing: 12,
                      mainAxisSpacing: 12,
                    ),
                    itemCount: plugins.length,
                    itemBuilder: (context, index) {
                      return _PluginCard(plugin: plugins[index]);
                    },
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.extension_off,
            size: 64,
            color: Colors.grey[400],
          ),
          const SizedBox(height: 16),
          Text(
            '没有找到相关插件',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
              color: Colors.grey[600],
            ),
          ),
          const SizedBox(height: 8),
          Text(
            '尝试其他搜索词或分类',
            style: TextStyle(color: Colors.grey[500]),
          ),
        ],
      ),
    );
  }

  String _getCategoryLabel(MarketplaceCategory category) {
    switch (category) {
      case MarketplaceCategory.all:
        return '全部';
      case MarketplaceCategory.channels:
        return '渠道';
      case MarketplaceCategory.skills:
        return '技能';
      case MarketplaceCategory.themes:
        return '主题';
      case MarketplaceCategory.integrations:
        return '集成';
    }
  }

  IconData _getCategoryIcon(MarketplaceCategory category) {
    switch (category) {
      case MarketplaceCategory.all:
        return Icons.apps;
      case MarketplaceCategory.channels:
        return Icons.cable;
      case MarketplaceCategory.skills:
        return Icons.extension;
      case MarketplaceCategory.themes:
        return Icons.palette;
      case MarketplaceCategory.integrations:
        return Icons.integration_instructions;
    }
  }
}

class _PluginCard extends ConsumerWidget {
  final PluginItem plugin;

  const _PluginCard({required this.plugin});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final isInstalled = ref.watch(installedPluginsProvider).contains(plugin.id);

    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () => _showPluginDetails(context),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Plugin Icon
            Container(
              height: 80,
              color: _getCategoryColor(plugin.category),
              child: Center(
                child: Icon(
                  _getPluginIcon(plugin.category),
                  size: 40,
                  color: Colors.white,
                ),
              ),
            ),
            
            // Plugin Info
            Expanded(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      plugin.name,
                      style: Theme.of(context).textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 4),
                    Expanded(
                      child: Text(
                        plugin.description,
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Colors.grey[600],
                        ),
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        Icon(Icons.star, size: 14, color: Colors.amber[700]),
                        const SizedBox(width: 4),
                        Text(
                          plugin.rating.toString(),
                          style: Theme.of(context).textTheme.bodySmall,
                        ),
                        const Spacer(),
                        Text(
                          'v${plugin.version}',
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: Colors.grey[500],
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            
            // Action Button
            Container(
              color: Theme.of(context).colorScheme.surfaceContainerHighest,
              child: TextButton(
                onPressed: isInstalled
                    ? () => _uninstallPlugin(context, ref)
                    : () => _installPlugin(context, ref),
                child: Text(
                  isInstalled ? '已安装' : '安装',
                  style: TextStyle(
                    color: isInstalled
                        ? Colors.grey[600]
                        : Theme.of(context).colorScheme.primary,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _installPlugin(BuildContext context, WidgetRef ref) async {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('正在安装 ${plugin.name}...'),
        duration: const Duration(seconds: 1),
      ),
    );
    
    await ref.read(marketplacePluginsProvider.notifier).installPlugin(plugin.id);
    
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('${plugin.name} 安装成功'),
          backgroundColor: Colors.green,
        ),
      );
    }
  }

  Future<void> _uninstallPlugin(BuildContext context, WidgetRef ref) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('确认卸载'),
        content: Text('确定要卸载 ${plugin.name} 吗？'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('卸载'),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      await ref.read(marketplacePluginsProvider.notifier).uninstallPlugin(plugin.id);
      
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('${plugin.name} 已卸载'),
          ),
        );
      }
    }
  }

  void _showPluginDetails(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.6,
        minChildSize: 0.3,
        maxChildSize: 0.9,
        expand: false,
        builder: (context, scrollController) {
          return SingleChildScrollView(
            controller: scrollController,
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Handle
                  Center(
                    child: Container(
                      width: 40,
                      height: 4,
                      decoration: BoxDecoration(
                        color: Colors.grey[300],
                        borderRadius: BorderRadius.circular(2),
                      ),
                    ),
                  ),
                  const SizedBox(height: 24),
                  
                  // Header
                  Row(
                    children: [
                      Container(
                        width: 64,
                        height: 64,
                        decoration: BoxDecoration(
                          color: _getCategoryColor(plugin.category),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Icon(
                          _getPluginIcon(plugin.category),
                          size: 32,
                          color: Colors.white,
                        ),
                      ),
                      const SizedBox(width: 16),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              plugin.name,
                              style: Theme.of(context).textTheme.titleLarge,
                            ),
                            Text(
                              'v${plugin.version} by ${plugin.author}',
                              style: TextStyle(color: Colors.grey[600]),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 24),
                  
                  // Stats
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceAround,
                    children: [
                      _buildStat(Icons.star, '${plugin.rating}', '评分'),
                      _buildStat(Icons.download, '${plugin.downloads}', '下载'),
                      _buildStat(Icons.code, '${plugin.version}', '版本'),
                    ],
                  ),
                  const SizedBox(height: 24),
                  
                  // Description
                  Text(
                    '描述',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 8),
                  Text(plugin.description),
                  const SizedBox(height: 24),
                  
                  // Features
                  Text(
                    '功能特性',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 8),
                  ...plugin.features.map((feature) => Padding(
                    padding: const EdgeInsets.only(bottom: 4),
                    child: Row(
                      children: [
                        const Icon(Icons.check_circle, size: 16, color: Colors.green),
                        const SizedBox(width: 8),
                        Text(feature),
                      ],
                    ),
                  )),
                  const SizedBox(height: 24),
                  
                  // Install Button
                  Consumer(
                    builder: (context, ref, child) {
                      final isInstalled = ref.watch(installedPluginsProvider).contains(plugin.id);
                      return SizedBox(
                        width: double.infinity,
                        child: ElevatedButton(
                          onPressed: () {
                            Navigator.pop(context);
                            if (isInstalled) {
                              _uninstallPlugin(context, ref);
                            } else {
                              _installPlugin(context, ref);
                            }
                          },
                          child: Text(isInstalled ? '卸载插件' : '安装插件'),
                        ),
                      );
                    },
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildStat(IconData icon, String value, String label) {
    return Column(
      children: [
        Icon(icon, color: Colors.grey[600]),
        const SizedBox(height: 4),
        Text(
          value,
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
        Text(
          label,
          style: TextStyle(color: Colors.grey[600], fontSize: 12),
        ),
      ],
    );
  }

  Color _getCategoryColor(String category) {
    switch (category) {
      case 'channels':
        return Colors.blue;
      case 'skills':
        return Colors.green;
      case 'themes':
        return Colors.purple;
      case 'integrations':
        return Colors.orange;
      default:
        return Colors.grey;
    }
  }

  IconData _getPluginIcon(String category) {
    switch (category) {
      case 'channels':
        return Icons.cable;
      case 'skills':
        return Icons.extension;
      case 'themes':
        return Icons.palette;
      case 'integrations':
        return Icons.integration_instructions;
      default:
        return Icons.extension;
    }
  }
}
