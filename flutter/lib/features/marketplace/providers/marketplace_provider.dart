import 'package:flutter_riverpod/flutter_riverpod.dart';

// Plugin item model
class PluginItem {
  final String id;
  final String name;
  final String description;
  final String author;
  final String version;
  final String category;
  final double rating;
  final int downloads;
  final List<String> features;
  final String? iconUrl;
  final bool isInstalled;

  const PluginItem({
    required this.id,
    required this.name,
    required this.description,
    required this.author,
    required this.version,
    required this.category,
    this.rating = 4.5,
    this.downloads = 1000,
    this.features = const [],
    this.iconUrl,
    this.isInstalled = false,
  });

  PluginItem copyWith({
    String? id,
    String? name,
    String? description,
    String? author,
    String? version,
    String? category,
    double? rating,
    int? downloads,
    List<String>? features,
    String? iconUrl,
    bool? isInstalled,
  }) {
    return PluginItem(
      id: id ?? this.id,
      name: name ?? this.name,
      description: description ?? this.description,
      author: author ?? this.author,
      version: version ?? this.version,
      category: category ?? this.category,
      rating: rating ?? this.rating,
      downloads: downloads ?? this.downloads,
      features: features ?? this.features,
      iconUrl: iconUrl ?? this.iconUrl,
      isInstalled: isInstalled ?? this.isInstalled,
    );
  }
}

// Marketplace plugins state
class MarketplacePluginsNotifier extends StateNotifier<List<PluginItem>> {
  MarketplacePluginsNotifier() : super(_samplePlugins);

  static const _samplePlugins = [
    // Channels
    PluginItem(
      id: 'matrix',
      name: 'Matrix 渠道',
      description: '支持 Matrix 协议的端对端加密通讯',
      author: 'Tortoise Team',
      version: '1.0.0',
      category: 'channels',
      rating: 4.8,
      downloads: 2500,
      features: ['E2E 加密', '房间管理', '同步历史'],
    ),
    PluginItem(
      id: 'slack',
      name: 'Slack 集成',
      description: '与企业级 Slack 工作区无缝集成',
      author: 'Tortoise Team',
      version: '1.2.0',
      category: 'channels',
      rating: 4.6,
      downloads: 3200,
      features: ['频道管理', '斜杠命令', '附件支持'],
    ),
    PluginItem(
      id: 'email',
      name: 'Email 渠道',
      description: '通过 SMTP/IMAP 收发邮件',
      author: 'Tortoise Team',
      version: '1.1.0',
      category: 'channels',
      rating: 4.5,
      downloads: 1800,
      features: ['SMTP 发送', 'IMAP 接收', '附件处理'],
    ),
    PluginItem(
      id: 'whatsapp',
      name: 'WhatsApp 渠道',
      description: '使用 Baileys 连接 WhatsApp',
      author: 'Community',
      version: '0.9.0',
      category: 'channels',
      rating: 4.2,
      downloads: 1500,
      features: ['多设备支持', '媒体发送', '状态管理'],
    ),
    
    // Skills
    PluginItem(
      id: 'code-review',
      name: '代码审查',
      description: '自动审查代码质量和安全问题',
      author: 'Tortoise Team',
      version: '2.0.0',
      category: 'skills',
      rating: 4.9,
      downloads: 4500,
      features: ['PR 审查', '安全扫描', '风格检查'],
    ),
    PluginItem(
      id: 'planning',
      name: '任务规划',
      description: '智能分解任务并生成执行计划',
      author: 'Tortoise Team',
      version: '1.5.0',
      category: 'skills',
      rating: 4.7,
      downloads: 2800,
      features: ['WBS 分解', '时间估算', '依赖分析'],
    ),
    PluginItem(
      id: 'github',
      name: 'GitHub 助手',
      description: 'GitHub 操作自动化和 PR 管理',
      author: 'Tortoise Team',
      version: '1.8.0',
      category: 'skills',
      rating: 4.8,
      downloads: 5100,
      features: ['PR 创建', 'Issue 管理', 'CI 集成'],
    ),
    PluginItem(
      id: 'notion',
      name: 'Notion 集成',
      description: '与 Notion 数据库同步和操作',
      author: 'Community',
      version: '1.0.0',
      category: 'skills',
      rating: 4.4,
      downloads: 1200,
      features: ['页面读取', '数据库操作', '块编辑'],
    ),
    
    // Integrations
    PluginItem(
      id: 'tailscale',
      name: 'Tailscale 网络',
      description: '通过 Tailscale 实现零信任网络',
      author: 'Tortoise Team',
      version: '1.0.0',
      category: 'integrations',
      rating: 4.6,
      downloads: 2000,
      features: ['设备发现', 'VPN 连接', 'ACL 管理'],
    ),
    PluginItem(
      id: 'office',
      name: 'Office 365',
      description: '与 Microsoft 365 套件集成',
      author: 'Community',
      version: '0.8.0',
      category: 'integrations',
      rating: 4.3,
      downloads: 980,
      features: ['邮件集成', '日历同步', '文件访问'],
    ),
    
    // Themes
    PluginItem(
      id: 'dark-pro',
      name: '专业深色主题',
      description: '精心设计的深色配色方案',
      author: 'Community',
      version: '1.0.0',
      category: 'themes',
      rating: 4.7,
      downloads: 3500,
      features: ['护眼配色', '多级灰度', '高对比度'],
    ),
  ];

  Future<void> installPlugin(String id) async {
    // Simulate installation
    await Future.delayed(const Duration(seconds: 1));
    
    state = state.map((plugin) {
      if (plugin.id == id) {
        return plugin.copyWith(isInstalled: true);
      }
      return plugin;
    }).toList();
  }

  Future<void> uninstallPlugin(String id) async {
    // Simulate uninstallation
    await Future.delayed(const Duration(milliseconds: 500));
    
    state = state.map((plugin) {
      if (plugin.id == id) {
        return plugin.copyWith(isInstalled: false);
      }
      return plugin;
    }).toList();
  }

  Future<void> refreshPlugins() async {
    // In production, fetch from API
    await Future.delayed(const Duration(seconds: 1));
    // State already contains sample data
  }
}

// Main marketplace provider
final marketplacePluginsProvider =
    StateNotifierProvider<MarketplacePluginsNotifier, List<PluginItem>>((ref) {
  return MarketplacePluginsNotifier();
});

// Installed plugins list provider
final installedPluginsProvider = Provider<List<String>>((ref) {
  final plugins = ref.watch(marketplacePluginsProvider);
  return plugins.where((p) => p.isInstalled).map((p) => p.id).toList();
});

// Plugin search provider
final pluginSearchProvider = StateProvider<String>((ref) => '');
