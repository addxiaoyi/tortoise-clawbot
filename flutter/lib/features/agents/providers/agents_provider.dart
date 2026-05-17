import 'package:flutter_riverpod/flutter_riverpod.dart';

// Agent model
class Agent {
  final String id;
  final String name;
  final String model;
  final String type;
  final String description;
  final String instructions;
  final List<String> capabilities;
  final bool enabled;

  const Agent({
    required this.id,
    required this.name,
    required this.model,
    required this.type,
    this.description = '',
    this.instructions = '',
    this.capabilities = const [],
    this.enabled = true,
  });

  Agent copyWith({
    String? id,
    String? name,
    String? model,
    String? type,
    String? description,
    String? instructions,
    List<String>? capabilities,
    bool? enabled,
  }) {
    return Agent(
      id: id ?? this.id,
      name: name ?? this.name,
      model: model ?? this.model,
      type: type ?? this.type,
      description: description ?? this.description,
      instructions: instructions ?? this.instructions,
      capabilities: capabilities ?? this.capabilities,
      enabled: enabled ?? this.enabled,
    );
  }
}

// Agents state
class AgentsState {
  final List<Agent> agents;
  final bool isLoading;
  final String? error;

  const AgentsState({
    this.agents = const [],
    this.isLoading = false,
    this.error,
  });

  AgentsState copyWith({
    List<Agent>? agents,
    bool? isLoading,
    String? error,
  }) {
    return AgentsState(
      agents: agents ?? this.agents,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

// Agents notifier
class AgentsNotifier extends StateNotifier<AgentsState> {
  AgentsNotifier() : super(const AgentsState()) {
    _loadSampleData();
  }

  void _loadSampleData() {
    state = state.copyWith(
      agents: [
        const Agent(
          id: 'coordinator-1',
          name: '主协调器',
          model: 'claude-3-5-sonnet-20241022',
          type: 'coordinator',
          description: '负责协调多个代理完成任务',
          instructions: '你是一个协调者，负责将复杂任务分解并分配给适当的代理。',
          capabilities: ['任务分解', '代理协调', '进度监控'],
          enabled: true,
        ),
        const Agent(
          id: 'researcher-1',
          name: '研究员',
          model: 'claude-3-5-sonnet-20241022',
          type: 'researcher',
          description: '收集和分析信息',
          instructions: '你是一个研究员，负责搜索和分析各种信息源。',
          capabilities: ['网络搜索', '数据分析', '摘要生成'],
          enabled: true,
        ),
        const Agent(
          id: 'coder-1',
          name: '程序员',
          model: 'claude-3-5-sonnet-20241022',
          type: 'coder',
          description: '编写和重构代码',
          instructions: '你是一个程序员，负责编写高质量的代码。',
          capabilities: ['代码生成', '代码审查', '调试'],
          enabled: true,
        ),
        const Agent(
          id: 'critic-1',
          name: '评论家',
          model: 'claude-3-opus-20240229',
          type: 'critic',
          description: '审查和评估工作质量',
          instructions: '你是一个评论家，负责审查和评估工作质量。',
          capabilities: ['代码审查', '质量评估', '安全检查'],
          enabled: false,
        ),
        const Agent(
          id: 'summarizer-1',
          name: '总结者',
          model: 'claude-3-5-sonnet-20241022',
          type: 'summarizer',
          description: '总结和简化信息',
          instructions: '你是一个总结者，负责将复杂信息简化为易于理解的摘要。',
          capabilities: ['内容总结', '要点提取', '格式整理'],
          enabled: false,
        ),
      ],
    );
  }

  Future<void> refreshAgents() async {
    state = state.copyWith(isLoading: true);
    await Future.delayed(const Duration(milliseconds: 500));
    _loadSampleData();
    state = state.copyWith(isLoading: false);
  }

  Future<void> addAgent(String name, String model, String instructions) async {
    final agent = Agent(
      id: 'agent-${DateTime.now().millisecondsSinceEpoch}',
      name: name,
      model: model,
      type: 'coordinator',
      instructions: instructions,
      capabilities: ['通用能力'],
      enabled: true,
    );
    
    state = state.copyWith(
      agents: [...state.agents, agent],
    );
  }

  Future<void> deleteAgent(String agentId) async {
    state = state.copyWith(
      agents: state.agents.where((a) => a.id != agentId).toList(),
    );
  }

  Future<void> toggleAgent(String agentId, bool enabled) async {
    state = state.copyWith(
      agents: state.agents.map((a) {
        if (a.id == agentId) {
          return a.copyWith(enabled: enabled);
        }
        return a;
      }).toList(),
    );
  }

  Future<void> updateAgent(Agent agent) async {
    state = state.copyWith(
      agents: state.agents.map((a) {
        if (a.id == agent.id) {
          return agent;
        }
        return a;
      }).toList(),
    );
  }
}

// Provider
final agentsProvider = StateNotifierProvider<AgentsNotifier, AgentsState>((ref) {
  return AgentsNotifier();
});

// Selected agent role filter
final selectedAgentTypeProvider = StateProvider<String?>((ref) => null);

// Filtered agents
final filteredAgentsProvider = Provider<List<Agent>>((ref) {
  final state = ref.watch(agentsProvider);
  final typeFilter = ref.watch(selectedAgentTypeProvider);
  
  var agents = state.agents;
  
  if (typeFilter != null) {
    agents = agents.where((a) => a.type == typeFilter).toList();
  }
  
  return agents;
});
