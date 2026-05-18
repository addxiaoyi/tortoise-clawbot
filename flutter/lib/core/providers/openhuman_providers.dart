import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../integrations/oauth_integrations.dart';
import '../integrations/local_knowledge.dart';
import '../desktop/desktop_manager.dart';
import '../coding/coding_tools.dart';
import '../automation/workflow_engine.dart';
import '../file_manager/file_ops.dart';

/// OpenHuman Providers 整合
/// 所有 OpenHuman 功能模块的 Riverpod Provider

// ============ OAuth 集成 ============

final oAuthManagerProvider = Provider<OAuthIntegrationManager>((ref) {
  return OAuthIntegrationManager();
});

final connectedProvidersProvider = Provider<List<OAuthProvider>>((ref) {
  final manager = ref.watch(oAuthManagerProvider);
  return manager.getConnected();
});

final providersByCategoryProvider = Provider.family<List<OAuthProvider>, String>((ref, category) {
  final manager = ref.watch(oAuthManagerProvider);
  return manager.getByCategory(category);
});

// ============ 本地知识库 ============

final knowledgeBaseProvider = StateProvider<LocalKnowledgeBase?>((ref) => null);

final knowledgeSearchProvider = StateProvider<String>((ref) => '');

final knowledgeSearchResultsProvider = Provider<List<SearchResult>>((ref) {
  final kb = ref.watch(knowledgeBaseProvider);
  final query = ref.watch(knowledgeSearchProvider);
  if (kb == null || query.isEmpty) return [];
  return kb.search(query);
});

// ============ 记忆图谱 ============

final memoryGraphProvider = StateNotifierProvider<MemoryGraphNotifier, MemoryGraph>((ref) {
  return MemoryGraphNotifier();
});

class MemoryGraphNotifier extends StateNotifier<MemoryGraph> {
  MemoryGraphNotifier() : super(MemoryGraph(nodes: [], edges: []));

  void addNode(MemoryNode node) {
    state = state.copyWith(
      nodes: [...state.nodes, node],
    );
  }

  void addEdge(MemoryEdge edge) {
    state = state.copyWith(
      edges: [...state.edges, edge],
    );
  }

  void removeNode(String id) {
    state = state.copyWith(
      nodes: state.nodes.where((n) => n.id != id).toList(),
      edges: state.edges.where((e) => e.from != id && e.to != id).toList(),
    );
  }

  List<MemoryNode> findPath(String from, String to) {
    // BFS 路径查找
    return [];
  }
}

// ============ 桌面管理 ============

final desktopManagerProvider = Provider<DesktopManager>((ref) {
  return DesktopManager();
});

final systemTrayProvider = Provider<SystemTrayManager>((ref) {
  return SystemTrayManager();
});

final hotkeyManagerProvider = Provider<GlobalHotkeyManager>((ref) {
  return GlobalHotkeyManager();
});

// ============ 编程工具 ============

final codingToolsProvider = Provider<CodingTools>((ref) {
  return CodingTools();
});

final codeExecutorProvider = Provider<CodeExecutor>((ref) {
  return CodeExecutor();
});

final gitAssistantProvider = Provider<GitAssistant>((ref) {
  return GitAssistant();
});

// ============ 自动化引擎 ============

final automationEngineProvider = Provider<AutomationEngine>((ref) {
  return AutomationEngine();
});

final taskSchedulerProvider = Provider<TaskScheduler>((ref) {
  return TaskScheduler();
});

// ============ 文件管理 ============

final fileManagerProvider = Provider<FileManager>((ref) {
  return FileManager();
});

// ============ 功能开关 ============

final featuresEnabledProvider = StateProvider<Map<String, bool>>((ref) => {
  'oauth': true,
  'knowledge': true,
  'coding': true,
  'automation': true,
  'voice': true,
  'screen': true,
});

// ============ OpenHuman 设置 ============

class OpenHumanSettings {
  final bool memoryEnabled;
  final bool learningEnabled;
  final bool voiceEnabled;
  final String wakeWord;
  final String language;

  OpenHumanSettings({
    this.memoryEnabled = true,
    this.learningEnabled = true,
    this.voiceEnabled = true,
    this.wakeWord = 'Hey Tortoise',
    this.language = 'zh-CN',
  });

  OpenHumanSettings copyWith({
    bool? memoryEnabled,
    bool? learningEnabled,
    bool? voiceEnabled,
    String? wakeWord,
    String? language,
  }) {
    return OpenHumanSettings(
      memoryEnabled: memoryEnabled ?? this.memoryEnabled,
      learningEnabled: learningEnabled ?? this.learningEnabled,
      voiceEnabled: voiceEnabled ?? this.voiceEnabled,
      wakeWord: wakeWord ?? this.wakeWord,
      language: language ?? this.language,
    );
  }
}

final openHumanSettingsProvider = StateNotifierProvider<OpenHumanSettingsNotifier, OpenHumanSettings>((ref) {
  return OpenHumanSettingsNotifier();
});

class OpenHumanSettingsNotifier extends StateNotifier<OpenHumanSettings> {
  OpenHumanSettingsNotifier() : super(OpenHumanSettings());

  void toggleMemory(bool value) {
    state = state.copyWith(memoryEnabled: value);
  }

  void toggleLearning(bool value) {
    state = state.copyWith(learningEnabled: value);
  }

  void toggleVoice(bool value) {
    state = state.copyWith(voiceEnabled: value);
  }

  void setWakeWord(String word) {
    state = state.copyWith(wakeWord: word);
  }

  void setLanguage(String lang) {
    state = state.copyWith(language: lang);
  }
}
