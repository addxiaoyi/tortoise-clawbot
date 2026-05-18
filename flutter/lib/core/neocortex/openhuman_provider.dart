import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'neocortex_memory.dart';
import 'subconscious_loop.dart';
import 'screen_intelligence.dart';
import 'voice_tools.dart';

/// OpenHuman Core Provider
/// 整合所有 OpenHuman 核心功能
class OpenHumanCore {
  final NeocortexManager neocortex;
  final SubconsciousLoop subconscious;
  final ScreenIntelligence screenIntelligence;
  final VoiceTools voiceTools;

  OpenHumanCore({
    required this.neocortex,
    required this.subconscious,
    required this.screenIntelligence,
    required this.voiceTools,
  });

  /// 初始化所有模块
  Future<void> initialize() async {
    await screenIntelligence.initialize();
    await voiceTools.initialize();
    subconscious.start();
  }

  /// 销毁所有模块
  void dispose() {
    subconscious.stop();
    voiceTools.dispose();
  }
}

/// OpenHuman 核心 Provider
final openHumanCoreProvider = Provider<OpenHumanCore>((ref) {
  final core = OpenHumanCore(
    neocortex: NeocortexManager(),
    subconscious: SubconsciousLoop(),
    screenIntelligence: ScreenIntelligence(),
    voiceTools: VoiceTools(),
  );

  ref.onDispose(() {
    core.dispose();
  });

  return core;
});

/// Neocortex 记忆 Provider
final neocortexProvider = Provider<NeocortexManager>((ref) {
  final core = ref.watch(openHumanCoreProvider);
  return core.neocortex;
});

/// Subconscious 循环 Provider
final subconsciousProvider = Provider<SubconsciousLoop>((ref) {
  final core = ref.watch(openHumanCoreProvider);
  return core.subconscious;
});

/// 屏幕智能 Provider
final screenIntelligenceProvider = Provider<ScreenIntelligence>((ref) {
  final core = ref.watch(openHumanCoreProvider);
  return core.screenIntelligence;
});

/// 语音工具 Provider
final voiceToolsProvider = Provider<VoiceTools>((ref) {
  final core = ref.watch(openHumanCoreProvider);
  return core.voiceTools;
});

/// 记忆层级状态
final memoryLayerFilterProvider = StateProvider<String?>((ref) => null);

/// 筛选后的记忆列表
final filteredMemoriesProvider = Provider<List<NeocortexMemory>>((ref) {
  final neocortex = ref.watch(neocortexProvider);
  final filter = ref.watch(memoryLayerFilterProvider);

  if (filter == null) {
    return neocortex.getAll();
  }
  return neocortex.getByLayer(filter);
});

/// 学习到的模式列表
final learnedPatternsProvider = Provider<List<LearnedPattern>>((ref) {
  final subconscious = ref.watch(subconsciousProvider);
  return subconscious.getPatterns();
});

/// 用户偏好
final userPreferencesProvider = Provider<Map<String, dynamic>>((ref) {
  final subconscious = ref.watch(subconsciousProvider);
  return subconscious.getPreferences();
});

/// 语音状态
final voiceStatusProvider = StateProvider<VoiceStatus>((ref) => VoiceStatus.idle);

/// 记忆搜索
final memorySearchQueryProvider = StateProvider<String>((ref) => '');

/// 搜索结果
final memorySearchResultsProvider = Provider<List<NeocortexMemory>>((ref) {
  final neocortex = ref.watch(neocortexProvider);
  final query = ref.watch(memorySearchQueryProvider);

  if (query.isEmpty) {
    return [];
  }
  return neocortex.search(query);
});

/// 添加记忆动作
class AddMemoryNotifier extends StateNotifier<AsyncValue<void>> {
  final NeocortexManager _neocortex;

  AddMemoryNotifier(this._neocortex) : super(const AsyncValue.data(null));

  Future<void> add({
    required String content,
    required String layer,
    double importance = 0.5,
    List<String> tags = const [],
  }) async {
    state = const AsyncValue.loading();
    try {
      _neocortex.addMemory(
        content: content,
        layer: layer,
        importance: importance,
        tags: tags,
      );
      state = const AsyncValue.data(null);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }
}

final addMemoryProvider = StateNotifierProvider<AddMemoryNotifier, AsyncValue<void>>((ref) {
  final neocortex = ref.watch(neocortexProvider);
  return AddMemoryNotifier(neocortex);
});

/// 记忆整合动作
class ConsolidateMemoriesNotifier extends StateNotifier<AsyncValue<void>> {
  final NeocortexManager _neocortex;

  ConsolidateMemoriesNotifier(this._neocortex) : super(const AsyncValue.data(null));

  Future<void> consolidate() async {
    state = const AsyncValue.loading();
    try {
      await _neocortex.consolidate();
      state = const AsyncValue.data(null);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }
}

final consolidateMemoriesProvider = StateNotifierProvider<ConsolidateMemoriesNotifier, AsyncValue<void>>((ref) {
  final neocortex = ref.watch(neocortexProvider);
  return ConsolidateMemoriesNotifier(neocortex);
});

/// 行为记录动作
class RecordBehaviorNotifier extends StateNotifier<void> {
  final SubconsciousLoop _subconscious;

  RecordBehaviorNotifier(this._subconscious) : super(null);

  void record({
    required String action,
    required String context,
    Map<String, dynamic> metadata = const {},
  }) {
    _subconscious.recordBehavior(
      action: action,
      context: context,
      metadata: metadata,
    );
  }
}

final recordBehaviorProvider = StateNotifierProvider<RecordBehaviorNotifier, void>((ref) {
  final subconscious = ref.watch(subconsciousProvider);
  return RecordBehaviorNotifier(subconscious);
});
