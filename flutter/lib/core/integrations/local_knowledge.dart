import 'dart:convert';
import 'dart:io';

/// Local Knowledge Base - 本地知识库
/// 支持 Obsidian vault 和本地文档索引
class LocalKnowledgeBase {
  final String _basePath;
  final List<Document> _documents = [];
  bool _isIndexed = false;

  LocalKnowledgeBase(this._basePath);

  /// 索引目录
  Future<void> index() async {
    _documents.clear();
    final dir = Directory(_basePath);
    if (!await dir.exists()) return;

    await for (final entity in dir.list(recursive: true)) {
      if (entity is File) {
        final ext = entity.path.split('.').last.toLowerCase();
        if (['md', 'txt', 'pdf', 'doc', 'docx'].contains(ext)) {
          await _indexFile(entity);
        }
      }
    }
    _isIndexed = true;
  }

  Future<void> _indexFile(File file) async {
    try {
      final content = await file.readAsString();
      final stat = await file.stat();
      _documents.add(Document(
        id: file.path,
        title: file.path.split('/').last.split('\\').last,
        content: content,
        path: file.path,
        modifiedAt: stat.modified,
        type: _getType(file.path),
      ));
    } catch (e) {
      // 跳过无法读取的文件
    }
  }

  String _getType(String path) {
    final ext = path.split('.').last.toLowerCase();
    switch (ext) {
      case 'md':
        return 'markdown';
      case 'txt':
        return 'text';
      case 'pdf':
        return 'pdf';
      case 'doc':
      case 'docx':
        return 'word';
      default:
        return 'unknown';
    }
  }

  /// 搜索
  List<SearchResult> search(String query, {int limit = 10}) {
    if (!_isIndexed) return [];
    final results = <SearchResult>[];
    final queryLower = query.toLowerCase();

    for (final doc in _documents) {
      if (doc.content.toLowerCase().contains(queryLower)) {
        final lines = doc.content.split('\n');
        final matches = <String>[];
        for (final line in lines) {
          if (line.toLowerCase().contains(queryLower)) {
            matches.add(line.trim());
            if (matches.length >= 3) break;
          }
        }
        results.add(SearchResult(
          document: doc,
          snippets: matches,
          score: _calculateScore(queryLower, doc.content),
        ));
      }
    }

    results.sort((a, b) => b.score.compareTo(a.score));
    return results.take(limit).toList();
  }

  double _calculateScore(String query, String content) {
    final count = RegExp(query, caseSensitive: false).allMatches(content).length;
    return count.toDouble();
  }

  /// 获取文档
  Document? getDocument(String id) {
    try {
      return _documents.firstWhere((d) => d.id == id);
    } catch (e) {
      return null;
    }
  }

  /// 获取所有文档
  List<Document> getAll() => List.from(_documents);

  /// 获取文档数量
  int get documentCount => _documents.length;

  /// 是否已索引
  bool get isIndexed => _isIndexed;
}

/// 文档
class Document {
  final String id;
  final String title;
  final String content;
  final String path;
  final DateTime modifiedAt;
  final String type;

  Document({
    required this.id,
    required this.title,
    required this.content,
    required this.path,
    required this.modifiedAt,
    required this.type,
  });

  Map<String, dynamic> toJson() => {
    'id': id,
    'title': title,
    'path': path,
    'modifiedAt': modifiedAt.toIso8601String(),
    'type': type,
  };
}

/// 搜索结果
class SearchResult {
  final Document document;
  final List<String> snippets;
  final double score;

  SearchResult({
    required this.document,
    required this.snippets,
    required this.score,
  });
}

/// Memory Graph - 记忆图谱
/// 实体关系图谱
class MemoryGraph {
  final Map<String, MemoryNode> _nodes = {};
  final List<MemoryEdge> _edges = [];

  /// 添加记忆节点
  void addNode(MemoryNode node) {
    _nodes[node.id] = node;
  }

  /// 添加关系边
  void addEdge(MemoryEdge edge) {
    _edges.add(edge);
  }

  /// 获取节点
  MemoryNode? getNode(String id) => _nodes[id];

  /// 获取节点的所有关系
  List<MemoryEdge> getEdges(String nodeId) {
    return _edges.where((e) => e.from == nodeId || e.to == nodeId).toList();
  }

  /// 获取关联节点
  List<MemoryNode> getConnectedNodes(String nodeId) {
    final connected = <String>{};
    for (final edge in getEdges(nodeId)) {
      if (edge.from == nodeId) {
        connected.add(edge.to);
      } else {
        connected.add(edge.from);
      }
    }
    return connected.map((id) => _nodes[id]).whereType<MemoryNode>().toList();
  }

  /// 获取所有节点
  List<MemoryNode> getAllNodes() => _nodes.values.toList();

  /// 获取所有边
  List<MemoryEdge> getAllEdges() => List.from(_edges);

  /// 搜索节点
  List<MemoryNode> searchNodes(String query) {
    final q = query.toLowerCase();
    return _nodes.values.where((n) =>
        n.label.toLowerCase().contains(q) ||
        n.content.toLowerCase().contains(q)).toList();
  }

  /// 导出为 JSON
  String export() {
    return jsonEncode({
      'nodes': _nodes.values.map((n) => n.toJson()).toList(),
      'edges': _edges.map((e) => e.toJson()).toList(),
    });
  }

  /// 从 JSON 导入
  void import_(String jsonStr) {
    final data = jsonDecode(jsonStr);
    for (final n in data['nodes']) {
      _nodes[n['id']] = MemoryNode.fromJson(n);
    }
    for (final e in data['edges']) {
      _edges.add(MemoryEdge.fromJson(e));
    }
  }
}

/// 记忆节点
class MemoryNode {
  final String id;
  final String label;
  final String content;
  final String type;
  final double importance;
  final DateTime createdAt;
  final Map<String, dynamic> metadata;

  MemoryNode({
    required this.id,
    required this.label,
    required this.content,
    required this.type,
    this.importance = 0.5,
    DateTime? createdAt,
    this.metadata = const {},
  }) : createdAt = createdAt ?? DateTime.now();

  Map<String, dynamic> toJson() => {
    'id': id,
    'label': label,
    'content': content,
    'type': type,
    'importance': importance,
    'createdAt': createdAt.toIso8601String(),
    'metadata': metadata,
  };

  factory MemoryNode.fromJson(Map<String, dynamic> json) => MemoryNode(
    id: json['id'],
    label: json['label'],
    content: json['content'],
    type: json['type'],
    importance: (json['importance'] as num?)?.toDouble() ?? 0.5,
    createdAt: DateTime.parse(json['createdAt']),
    metadata: Map<String, dynamic>.from(json['metadata'] ?? {}),
  );
}

/// 记忆边
class MemoryEdge {
  final String from;
  final String to;
  final String relation;
  final double weight;

  MemoryEdge({
    required this.from,
    required this.to,
    required this.relation,
    this.weight = 1.0,
  });

  Map<String, dynamic> toJson() => {
    'from': from,
    'to': to,
    'relation': relation,
    'weight': weight,
  };

  factory MemoryEdge.fromJson(Map<String, dynamic> json) => MemoryEdge(
    from: json['from'],
    to: json['to'],
    relation: json['relation'],
    weight: (json['weight'] as num?)?.toDouble() ?? 1.0,
  );
}

/// Memory Tree - 记忆树
/// 分层记忆结构可视化
class MemoryTree {
  final String id;
  final String label;
  final List<MemoryTree> children = [];
  final Map<String, dynamic> data;
  bool isExpanded = true;

  MemoryTree({
    required this.id,
    required this.label,
    this.data = const {},
    this.children = const [],
  });

  /// 添加子节点
  void addChild(MemoryTree child) {
    children.add(child);
  }

  /// 移除子节点
  void removeChild(String childId) {
    children.removeWhere((c) => c.id == childId);
  }

  /// 查找节点
  MemoryTree? find(String nodeId) {
    if (id == nodeId) return this;
    for (final child in children) {
      final found = child.find(nodeId);
      if (found != null) return found;
    }
    return null;
  }

  /// 展平为列表
  List<MemoryTree> flatten() {
    final result = <MemoryTree>[this];
    for (final child in children) {
      result.addAll(child.flatten());
    }
    return result;
  }
}
