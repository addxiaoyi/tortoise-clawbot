import 'dart:convert';

/// Neocortex Memory - 神经皮层记忆系统
/// 模拟人类大脑皮层的分层记忆架构
class NeocortexMemory {
  /// 记忆层级
  static const String layerWorking = 'working';      // 工作记忆 (几秒-几分钟)
  static const String layerShortTerm = 'short_term'; // 短时记忆 (几分钟-几天)
  static const String layerLongTerm = 'long_term';   // 长时记忆 (几天-永久)
  static const String layerSemantic = 'semantic';    // 语义记忆 (概念、事实)
  static const String layerEpisodic = 'episodic';    //情景记忆 (经历、事件)

  /// 记忆项
  final String id;
  final String content;
  final String layer;
  final double importance; // 0.0 - 1.0
  final int accessCount;
  final DateTime createdAt;
  final DateTime lastAccess;
  final List<String> tags;
  final Map<String, dynamic> metadata;

  NeocortexMemory({
    required this.id,
    required this.content,
    required this.layer,
    this.importance = 0.5,
    this.accessCount = 0,
    required this.createdAt,
    required this.lastAccess,
    this.tags = const [],
    this.metadata = const {},
  });

  /// 重要性阈值 - 决定是否升级到长时记忆
  static const double importanceThreshold = 0.7;

  /// 访问次数阈值 - 决定是否升级
  static const int accessThreshold = 5;

  /// 是否应该升级到更高层级
  bool shouldPromote() {
    if (layer == layerLongTerm) return false;
    return importance >= importanceThreshold || accessCount >= accessThreshold;
  }

  /// 是否应该降级
  bool shouldDemote() {
    if (layer == layerWorking) return false;
    final age = DateTime.now().difference(lastAccess).inDays;
    if (age > 30 && accessCount < 2) return true;
    return false;
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'content': content,
    'layer': layer,
    'importance': importance,
    'accessCount': accessCount,
    'createdAt': createdAt.toIso8601String(),
    'lastAccess': lastAccess.toIso8601String(),
    'tags': tags,
    'metadata': metadata,
  };

  factory NeocortexMemory.fromJson(Map<String, dynamic> json) => NeocortexMemory(
    id: json['id'],
    content: json['content'],
    layer: json['layer'],
    importance: (json['importance'] as num).toDouble(),
    accessCount: json['accessCount'],
    createdAt: DateTime.parse(json['createdAt']),
    lastAccess: DateTime.parse(json['lastAccess']),
    tags: List<String>.from(json['tags'] ?? []),
    metadata: Map<String, dynamic>.from(json['metadata'] ?? {}),
  );
}

/// 记忆管理器
class NeocortexManager {
  final Map<String, NeocortexMemory> _memories = {};

  /// 添加记忆
  String addMemory({
    required String content,
    required String layer,
    double importance = 0.5,
    List<String> tags = const [],
    Map<String, dynamic> metadata = const {},
  }) {
    final id = DateTime.now().millisecondsSinceEpoch.toString();
    final now = DateTime.now();

    _memories[id] = NeocortexMemory(
      id: id,
      content: content,
      layer: layer,
      importance: importance,
      createdAt: now,
      lastAccess: now,
      tags: tags,
      metadata: metadata,
    );

    return id;
  }

  /// 访问记忆
  NeocortexMemory? accessMemory(String id) {
    final memory = _memories[id];
    if (memory == null) return null;

    // 更新访问信息
    _memories[id] = NeocortexMemory(
      id: memory.id,
      content: memory.content,
      layer: memory.layer,
      importance: memory.importance,
      accessCount: memory.accessCount + 1,
      createdAt: memory.createdAt,
      lastAccess: DateTime.now(),
      tags: memory.tags,
      metadata: memory.metadata,
    );

    return _memories[id];
  }

  /// 获取指定层级的记忆
  List<NeocortexMemory> getByLayer(String layer) {
    return _memories.values
        .where((m) => m.layer == layer)
        .toList()
      ..sort((a, b) => b.importance.compareTo(a.importance));
  }

  /// 搜索记忆
  List<NeocortexMemory> search(String query) {
    final lowerQuery = query.toLowerCase();
    return _memories.values
        .where((m) =>
            m.content.toLowerCase().contains(lowerQuery) ||
            m.tags.any((t) => t.toLowerCase().contains(lowerQuery)))
        .toList()
      ..sort((a, b) => b.importance.compareTo(a.importance));
  }

  /// 获取所有记忆
  List<NeocortexMemory> getAll() {
    return _memories.values.toList()
      ..sort((a, b) => b.lastAccess.compareTo(a.lastAccess));
  }

  /// 记忆 consolidation (整合)
  /// 将短期记忆整合到长期记忆
  Future<void> consolidate() async {
    final toPromote = <String>[];
    final toDemote = <String>[];

    for (final memory in _memories.values) {
      if (memory.shouldPromote()) {
        toPromote.add(memory.id);
      } else if (memory.shouldDemote()) {
        toDemote.add(memory.id);
      }
    }

    // 升级记忆
    for (final id in toPromote) {
      final memory = _memories[id];
      if (memory == null) continue;

      String newLayer;
      switch (memory.layer) {
        case NeocortexMemory.layerWorking:
          newLayer = NeocortexMemory.layerShortTerm;
          break;
        case NeocortexMemory.layerShortTerm:
          newLayer = NeocortexMemory.layerLongTerm;
          break;
        default:
          newLayer = memory.layer;
      }

      _memories[id] = NeocortexMemory(
        id: memory.id,
        content: memory.content,
        layer: newLayer,
        importance: memory.importance * 1.1, // 提升重要性
        accessCount: memory.accessCount,
        createdAt: memory.createdAt,
        lastAccess: memory.lastAccess,
        tags: memory.tags,
        metadata: memory.metadata,
      );
    }

    // 降级记忆
    for (final id in toDemote) {
      final memory = _memories[id];
      if (memory == null) continue;

      String newLayer;
      switch (memory.layer) {
        case NeocortexMemory.layerLongTerm:
          newLayer = NeocortexMemory.layerShortTerm;
          break;
        case NeocortexMemory.layerShortTerm:
          newLayer = NeocortexMemory.layerWorking;
          break;
        default:
          newLayer = memory.layer;
      }

      _memories[id] = NeocortexMemory(
        id: memory.id,
        content: memory.content,
        layer: newLayer,
        importance: memory.importance * 0.9,
        accessCount: memory.accessCount,
        createdAt: memory.createdAt,
        lastAccess: memory.lastAccess,
        tags: memory.tags,
        metadata: memory.metadata,
      );
    }
  }

  /// 导出记忆到 JSON
  String export() {
    return jsonEncode(_memories.values.map((m) => m.toJson()).toList());
  }

  /// 从 JSON 导入记忆
  void import_(String jsonStr) {
    final List<dynamic> data = jsonDecode(jsonStr);
    for (final item in data) {
      final memory = NeocortexMemory.fromJson(item);
      _memories[memory.id] = memory;
    }
  }
}
