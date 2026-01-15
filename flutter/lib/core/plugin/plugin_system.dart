// 插件系统 - 遵循 OpenClaw 插件规范

/// 插件元数据
class PluginMetadata {
  final String id;
  final String name;
  final String version;
  final String description;
  final String author;
  final String? icon;
  final List<String> dependencies;
  final Map<String, dynamic> config;

  PluginMetadata({
    required this.id,
    required this.name,
    required this.version,
    required this.description,
    required this.author,
    this.icon,
    this.dependencies = const [],
    this.config = const {},
  });

  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'version': version,
    'description': description,
    'author': author,
    'icon': icon,
    'dependencies': dependencies,
    'config': config,
  };

  factory PluginMetadata.fromJson(Map<String, dynamic> json) => PluginMetadata(
    id: json['id'] as String,
    name: json['name'] as String,
    version: json['version'] as String,
    description: json['description'] as String,
    author: json['author'] as String,
    icon: json['icon'] as String?,
    dependencies: (json['dependencies'] as List<dynamic>?)?.cast<String>() ?? [],
    config: Map<String, dynamic>.from(json['config'] as Map<dynamic, dynamic>? ?? {}),
  );
}

/// 插件状态
enum PluginState {
  uninstalled,
  installing,
  installed,
  enabled,
  disabled,
  error,
}

/// 插件信息
class Plugin {
  final PluginMetadata metadata;
  final PluginState state;
  final String? error;
  final DateTime? installedAt;
  final DateTime? lastUpdated;

  Plugin({
    required this.metadata,
    this.state = PluginState.uninstalled,
    this.error,
    this.installedAt,
    this.lastUpdated,
  });

  Plugin copyWith({
    PluginMetadata? metadata,
    PluginState? state,
    String? error,
    DateTime? installedAt,
    DateTime? lastUpdated,
  }) {
    return Plugin(
      metadata: metadata ?? this.metadata,
      state: state ?? this.state,
      error: error ?? this.error,
      installedAt: installedAt ?? this.installedAt,
      lastUpdated: lastUpdated ?? this.lastUpdated,
    );
  }

  Map<String, dynamic> toJson() => {
    'metadata': metadata.toJson(),
    'state': state.name,
    'error': error,
    'installedAt': installedAt?.toIso8601String(),
    'lastUpdated': lastUpdated?.toIso8601String(),
  };
}

/// 插件基类
abstract class TortoisePlugin {
  PluginMetadata get metadata;
  
  Future<void> onLoad() async {}
  Future<void> onUnload() async {}
  Future<void> onEnable() async {}
  Future<void> onDisable() async {}
  Map<String, dynamic> getConfig() => {};
  Future<void> setConfig(Map<String, dynamic> config) async {}
}

/// 插件管理器
class PluginManager {
  static PluginManager? _instance;
  static PluginManager get instance => _instance ??= PluginManager._();
  PluginManager._();

  final Map<String, TortoisePlugin> _plugins = {};
  final Map<String, Plugin> _pluginInfo = {};
  
  List<Plugin> get plugins => _pluginInfo.values.toList();
  
  Future<void> loadPlugin(TortoisePlugin plugin) async {
    final id = plugin.metadata.id;
    _plugins[id] = plugin;
    _pluginInfo[id] = Plugin(
      metadata: plugin.metadata,
      state: PluginState.installed,
      installedAt: DateTime.now(),
    );
    await plugin.onLoad();
  }
  
  Future<void> enablePlugin(String id) async {
    final plugin = _plugins[id];
    if (plugin == null) return;
    await plugin.onEnable();
    _pluginInfo[id] = _pluginInfo[id]!.copyWith(state: PluginState.enabled);
  }
  
  Future<void> disablePlugin(String id) async {
    final plugin = _plugins[id];
    if (plugin == null) return;
    await plugin.onDisable();
    _pluginInfo[id] = _pluginInfo[id]!.copyWith(state: PluginState.disabled);
  }
  
  Future<void> unloadPlugin(String id) async {
    final plugin = _plugins[id];
    if (plugin == null) return;
    await plugin.onUnload();
    _plugins.remove(id);
    _pluginInfo.remove(id);
  }
  
  Plugin? getPlugin(String id) => _pluginInfo[id];
  
  bool isEnabled(String id) => _pluginInfo[id]?.state == PluginState.enabled;
}
