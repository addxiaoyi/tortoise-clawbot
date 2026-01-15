import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/memory_models.dart';

// Memory types
enum MemoryType {
  fact,
  preference,
  interest,
  work,
  personal,
}

// Memory entry model
class MemoryEntry {
  final String id;
  final String content;
  final MemoryType type;
  final double importance;
  final int accessCount;
  final DateTime? lastAccessed;
  final DateTime createdAt;
  final Map<String, dynamic> metadata;

  const MemoryEntry({
    required this.id,
    required this.content,
    required this.type,
    this.importance = 0.5,
    this.accessCount = 0,
    this.lastAccessed,
    required this.createdAt,
    this.metadata = const {},
  });

  MemoryEntry copyWith({
    String? id,
    String? content,
    MemoryType? type,
    double? importance,
    int? accessCount,
    DateTime? lastAccessed,
    DateTime? createdAt,
    Map<String, dynamic>? metadata,
  }) {
    return MemoryEntry(
      id: id ?? this.id,
      content: content ?? this.content,
      type: type ?? this.type,
      importance: importance ?? this.importance,
      accessCount: accessCount ?? this.accessCount,
      lastAccessed: lastAccessed ?? this.lastAccessed,
      createdAt: createdAt ?? this.createdAt,
      metadata: metadata ?? this.metadata,
    );
  }

  factory MemoryEntry.fromJson(Map<String, dynamic> json) {
    return MemoryEntry(
      id: json['id'] ?? '',
      content: json['content'] ?? '',
      type: MemoryType.values.firstWhere(
        (e) => e.name == json['type'],
        orElse: () => MemoryType.fact,
      ),
      importance: (json['importance'] ?? 0.5).toDouble(),
      accessCount: json['access_count'] ?? 0,
      lastAccessed: json['last_accessed'] != null
          ? DateTime.parse(json['last_accessed'])
          : null,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
      metadata: json['metadata'] ?? {},
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'content': content,
      'type': type.name,
      'importance': importance,
      'access_count': accessCount,
      'last_accessed': lastAccessed?.toIso8601String(),
      'created_at': createdAt.toIso8601String(),
      'metadata': metadata,
    };
  }
}

// Memory state
class MemoryState {
  final List<MemoryEntry> entries;
  final bool isLoading;
  final String? error;
  final String searchQuery;

  const MemoryState({
    this.entries = const [],
    this.isLoading = false,
    this.error,
    this.searchQuery = '',
  });

  MemoryState copyWith({
    List<MemoryEntry>? entries,
    bool? isLoading,
    String? error,
    String? searchQuery,
  }) {
    return MemoryState(
      entries: entries ?? this.entries,
      isLoading: isLoading ?? this.isLoading,
      error: error,
      searchQuery: searchQuery ?? this.searchQuery,
    );
  }

  List<MemoryEntry> get filteredEntries {
    if (searchQuery.isEmpty) return entries;
    final query = searchQuery.toLowerCase();
    return entries.where((e) =>
      e.content.toLowerCase().contains(query)
    ).toList();
  }

  Map<MemoryType, List<MemoryEntry>> get groupedByType {
    final grouped = <MemoryType, List<MemoryEntry>>{};
    for (final type in MemoryType.values) {
      grouped[type] = entries.where((e) => e.type == type).toList();
    }
    return grouped;
  }
}

// Memory notifier
class MemoryNotifier extends StateNotifier<MemoryState> {
  MemoryNotifier() : super(const MemoryState()) {
    _loadSampleData();
  }

  void _loadSampleData() {
    state = state.copyWith(
      entries: [
        MemoryEntry(
          id: '1',
          content: '用户喜欢在晚上工作',
          type: MemoryType.preference,
          importance: 0.8,
          accessCount: 5,
          createdAt: DateTime.now().subtract(const Duration(days: 7)),
        ),
        MemoryEntry(
          id: '2',
          content: '用户对人工智能和机器学习感兴趣',
          type: MemoryType.interest,
          importance: 0.9,
          accessCount: 12,
          createdAt: DateTime.now().subtract(const Duration(days: 14)),
        ),
        MemoryEntry(
          id: '3',
          content: '用户是一名全栈开发工程师',
          type: MemoryType.work,
          importance: 0.7,
          accessCount: 3,
          createdAt: DateTime.now().subtract(const Duration(days: 30)),
        ),
        MemoryEntry(
          id: '4',
          content: 'Tortoise 是一个高性能的 AI 代理框架',
          type: MemoryType.fact,
          importance: 1.0,
          accessCount: 20,
          createdAt: DateTime.now().subtract(const Duration(days: 60)),
        ),
      ],
    );
  }

  Future<void> search(String query) async {
    state = state.copyWith(searchQuery: query, isLoading: true);
    
    // Simulate API call
    await Future.delayed(const Duration(milliseconds: 500));
    
    state = state.copyWith(isLoading: false);
  }

  Future<void> addEntry(MemoryEntry entry) async {
    state = state.copyWith(isLoading: true);
    
    // Simulate API call
    await Future.delayed(const Duration(milliseconds: 300));
    
    state = state.copyWith(
      entries: [...state.entries, entry],
      isLoading: false,
    );
  }

  Future<void> deleteEntry(String id) async {
    state = state.copyWith(
      entries: state.entries.where((e) => e.id != id).toList(),
    );
  }

  Future<void> updateEntry(MemoryEntry entry) async {
    state = state.copyWith(
      entries: state.entries.map((e) => 
        e.id == entry.id ? entry : e
      ).toList(),
    );
  }

  Future<void> refresh() async {
    state = state.copyWith(isLoading: true);
    await Future.delayed(const Duration(seconds: 1));
    _loadSampleData();
    state = state.copyWith(isLoading: false);
  }
}

// Provider
final memoryProvider = StateNotifierProvider<MemoryNotifier, MemoryState>((ref) {
  return MemoryNotifier();
});

// Selected memory type filter
final memoryTypeFilterProvider = StateProvider<MemoryType?>((ref) => null);

// Filtered memories
final filteredMemoriesProvider = Provider<List<MemoryEntry>>((ref) {
  final state = ref.watch(memoryProvider);
  final typeFilter = ref.watch(memoryTypeFilterProvider);
  
  var entries = state.filteredEntries;
  
  if (typeFilter != null) {
    entries = entries.where((e) => e.type == typeFilter).toList();
  }
  
  return entries;
});
