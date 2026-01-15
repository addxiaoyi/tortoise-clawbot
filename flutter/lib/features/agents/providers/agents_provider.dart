import 'package:flutter_riverpod/flutter_riverpod.dart';

// Agent role types
enum AgentRole {
  coordinator,
  researcher,
  coder,
  reviewer,
  executor,
  specialist,
}

// Agent model
class Agent {
  final String id;
  final String name;
  final String description;
  final AgentRole role;
  final List<String> skills;
  final int maxConcurrentTasks;
  final int priority;
  final bool isActive;

  const Agent({
    required this.id,
    required this.name,
    required this.description,
    required this.role,
    this.skills = const [],
    this.maxConcurrentTasks = 3,
    this.priority = 50,
    this.isActive = true,
  });

  Agent copyWith({
    String? id,
    String? name,
    String? description,
    AgentRole? role,
    List<String>? skills,
    int? maxConcurrentTasks,
    int? priority,
    bool? isActive,
  }) {
    return Agent(
      id: id ?? this.id,
      name: name ?? this.name,
      description: description ?? this.description,
      role: role ?? this.role,
      skills: skills ?? this.skills,
      maxConcurrentTasks: maxConcurrentTasks ?? this.maxConcurrentTasks,
      priority: priority ?? this.priority,
      isActive: isActive ?? this.isActive,
    );
  }

  factory Agent.fromJson(Map<String, dynamic> json) {
    return Agent(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      description: json['description'] ?? '',
      role: AgentRole.values.firstWhere(
        (e) => e.name == json['role'],
        orElse: () => AgentRole.specialist,
      ),
      skills: List<String>.from(json['skills'] ?? []),
      maxConcurrentTasks: json['max_concurrent_tasks'] ?? 3,
      priority: json['priority'] ?? 50,
      isActive: json['is_active'] ?? true,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'description': description,
      'role': role.name,
      'skills': skills,
      'max_concurrent_tasks': maxConcurrentTasks,
      'priority': priority,
      'is_active': isActive,
    };
  }
}

// Task model
class AgentTask {
  final String id;
  final String description;
  final String status; // pending, running, completed, failed
  final String? assignedAgentId;
  final DateTime createdAt;
  final DateTime? completedAt;
  final dynamic result;

  const AgentTask({
    required this.id,
    required this.description,
    this.status = 'pending',
    this.assignedAgentId,
    required this.createdAt,
    this.completedAt,
    this.result,
  });

  AgentTask copyWith({
    String? id,
    String? description,
    String? status,
    String? assignedAgentId,
    DateTime? createdAt,
    DateTime? completedAt,
    dynamic result,
  }) {
    return AgentTask(
      id: id ?? this.id,
      description: description ?? this.description,
      status: status ?? this.status,
      assignedAgentId: assignedAgentId ?? this.assignedAgentId,
      createdAt: createdAt ?? this.createdAt,
      completedAt: completedAt ?? this.completedAt,
      result: result ?? this.result,
    );
  }
}

// Agents state
class AgentsState {
  final List<Agent> agents;
  final List<AgentTask> tasks;
  final bool isLoading;
  final String? error;

  const AgentsState({
    this.agents = const [],
    this.tasks = const [],
    this.isLoading = false,
    this.error,
  });

  AgentsState copyWith({
    List<Agent>? agents,
    List<AgentTask>? tasks,
    bool? isLoading,
    String? error,
  }) {
    return AgentsState(
      agents: agents ?? this.agents,
      tasks: tasks ?? this.tasks,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }

  List<AgentTask> get pendingTasks =>
      tasks.where((t) => t.status == 'pending').toList();

  List<AgentTask> get runningTasks =>
      tasks.where((t) => t.status == 'running').toList();

  List<AgentTask> get completedTasks =>
      tasks.where((t) => t.status == 'completed').toList();
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
          description: '负责协调多个代理完成任务',
          role: AgentRole.coordinator,
          skills: ['planning', 'delegation', 'monitoring'],
          maxConcurrentTasks: 3,
          priority: 100,
        ),
        const Agent(
          id: 'researcher-1',
          name: '研究员',
          description: '收集和分析信息',
          role: AgentRole.researcher,
          skills: ['web-search', 'analysis', 'summarization'],
          maxConcurrentTasks: 5,
          priority: 50,
        ),
        const Agent(
          id: 'coder-1',
          name: '程序员',
          description: '编写和重构代码',
          role: AgentRole.coder,
          skills: ['code-generation', 'code-review', 'debugging'],
          maxConcurrentTasks: 3,
          priority: 60,
        ),
        const Agent(
          id: 'reviewer-1',
          name: '审查员',
          description: '审查和评估工作质量',
          role: AgentRole.reviewer,
          skills: ['code-review', 'quality', 'security'],
          maxConcurrentTasks: 4,
          priority: 40,
        ),
        const Agent(
          id: 'executor-1',
          name: '执行器',
          description: '执行命令和脚本',
          role: AgentRole.executor,
          skills: ['bash', 'git', 'deployment'],
          maxConcurrentTasks: 2,
          priority: 30,
        ),
      ],
      tasks: [
        AgentTask(
          id: 'task-1',
          description: '分析项目需求文档',
          status: 'completed',
          assignedAgentId: 'researcher-1',
          createdAt: DateTime.now().subtract(const Duration(hours: 2)),
          completedAt: DateTime.now().subtract(const Duration(hours: 1)),
          result: '需求分析完成',
        ),
        AgentTask(
          id: 'task-2',
          description: '实现用户认证模块',
          status: 'running',
          assignedAgentId: 'coder-1',
          createdAt: DateTime.now().subtract(const Duration(minutes: 30)),
        ),
        AgentTask(
          id: 'task-3',
          description: '编写单元测试',
          status: 'pending',
          createdAt: DateTime.now(),
        ),
      ],
    );
  }

  Future<void> createTask(String description) async {
    final task = AgentTask(
      id: 'task-${DateTime.now().millisecondsSinceEpoch}',
      description: description,
      createdAt: DateTime.now(),
    );
    
    state = state.copyWith(
      tasks: [...state.tasks, task],
    );
  }

  Future<void> assignTask(String taskId, String agentId) async {
    state = state.copyWith(
      tasks: state.tasks.map((t) {
        if (t.id == taskId) {
          return t.copyWith(
            assignedAgentId: agentId,
            status: 'running',
          );
        }
        return t;
      }).toList(),
    );
  }

  Future<void> completeTask(String taskId, dynamic result) async {
    state = state.copyWith(
      tasks: state.tasks.map((t) {
        if (t.id == taskId) {
          return t.copyWith(
            status: 'completed',
            completedAt: DateTime.now(),
            result: result,
          );
        }
        return t;
      }).toList(),
    );
  }

  Future<void> toggleAgent(String agentId) async {
    state = state.copyWith(
      agents: state.agents.map((a) {
        if (a.id == agentId) {
          return a.copyWith(isActive: !a.isActive);
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

// Selected agent filter
final selectedAgentRoleProvider = StateProvider<AgentRole?>((ref) => null);

// Filtered agents
final filteredAgentsProvider = Provider<List<Agent>>((ref) {
  final state = ref.watch(agentsProvider);
  final roleFilter = ref.watch(selectedAgentRoleProvider);
  
  var agents = state.agents;
  
  if (roleFilter != null) {
    agents = agents.where((a) => a.role == roleFilter).toList();
  }
  
  return agents;
});
