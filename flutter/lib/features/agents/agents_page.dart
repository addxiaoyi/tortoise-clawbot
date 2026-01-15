import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'providers/agents_provider.dart';

class AgentsPage extends ConsumerStatefulWidget {
  const AgentsPage({super.key});

  @override
  ConsumerState<AgentsPage> createState() => _AgentsPageState();
}

class _AgentsPageState extends ConsumerState<AgentsPage>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(agentsProvider);
    final filteredAgents = ref.watch(filteredAgentsProvider);
    final roleFilter = ref.watch(selectedAgentRoleProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('多代理'),
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(text: '代理', icon: Icon(Icons.smart_toy)),
            Tab(text: '任务', icon: Icon(Icons.assignment)),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          // Agents Tab
          Column(
            children: [
              // Role Filter
              SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    FilterChip(
                      label: const Text('全部'),
                      selected: roleFilter == null,
                      onSelected: (_) {
                        ref.read(selectedAgentRoleProvider.notifier).state = null;
                      },
                    ),
                    const SizedBox(width: 8),
                    ...AgentRole.values.map((role) {
                      return Padding(
                        padding: const EdgeInsets.only(right: 8),
                        child: FilterChip(
                          label: Text(_getRoleName(role)),
                          selected: roleFilter == role,
                          onSelected: (_) {
                            ref.read(selectedAgentRoleProvider.notifier).state = role;
                          },
                        ),
                      );
                    }),
                  ],
                ),
              ),
              // Agents List
              Expanded(
                child: state.isLoading
                    ? const Center(child: CircularProgressIndicator())
                    : filteredAgents.isEmpty
                        ? _buildEmptyAgents()
                        : ListView.builder(
                            padding: const EdgeInsets.all(16),
                            itemCount: filteredAgents.length,
                            itemBuilder: (context, index) {
                              final agent = filteredAgents[index];
                              return _AgentCard(
                                agent: agent,
                                onToggle: () {
                                  ref
                                      .read(agentsProvider.notifier)
                                      .toggleAgent(agent.id);
                                },
                              );
                            },
                          ),
              ),
            ],
          ),
          // Tasks Tab
          Column(
            children: [
              // Task Stats
              Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                  children: [
                    _StatCard(
                      icon: Icons.pending,
                      label: '待处理',
                      count: state.pendingTasks.length,
                      color: Colors.orange,
                    ),
                    _StatCard(
                      icon: Icons.play_circle,
                      label: '进行中',
                      count: state.runningTasks.length,
                      color: Colors.blue,
                    ),
                    _StatCard(
                      icon: Icons.check_circle,
                      label: '已完成',
                      count: state.completedTasks.length,
                      color: Colors.green,
                    ),
                  ],
                ),
              ),
              // Tasks List
              Expanded(
                child: state.tasks.isEmpty
                    ? _buildEmptyTasks()
                    : ListView.builder(
                        padding: const EdgeInsets.all(16),
                        itemCount: state.tasks.length,
                        itemBuilder: (context, index) {
                          final task = state.tasks[index];
                          return _TaskCard(
                            task: task,
                            agents: state.agents,
                            onAssign: (agentId) {
                              ref
                                  .read(agentsProvider.notifier)
                                  .assignTask(task.id, agentId);
                            },
                            onComplete: (result) {
                              ref
                                  .read(agentsProvider.notifier)
                                  .completeTask(task.id, result);
                            },
                          );
                        },
                      ),
              ),
            ],
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showCreateTaskDialog(context),
        child: const Icon(Icons.add),
      ),
    );
  }

  Widget _buildEmptyAgents() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.smart_toy_outlined,
            size: 80,
            color: Colors.grey[400],
          ),
          const SizedBox(height: 16),
          Text(
            '暂无代理',
            style: TextStyle(
              fontSize: 18,
              color: Colors.grey[600],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyTasks() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.assignment_outlined,
            size: 80,
            color: Colors.grey[400],
          ),
          const SizedBox(height: 16),
          Text(
            '暂无任务',
            style: TextStyle(
              fontSize: 18,
              color: Colors.grey[600],
            ),
          ),
          const SizedBox(height: 8),
          Text(
            '点击 + 创建新任务',
            style: TextStyle(
              color: Colors.grey[500],
            ),
          ),
        ],
      ),
    );
  }

  String _getRoleName(AgentRole role) {
    switch (role) {
      case AgentRole.coordinator:
        return '协调器';
      case AgentRole.researcher:
        return '研究员';
      case AgentRole.coder:
        return '程序员';
      case AgentRole.reviewer:
        return '审查员';
      case AgentRole.executor:
        return '执行器';
      case AgentRole.specialist:
        return '专家';
    }
  }

  Future<void> _showCreateTaskDialog(BuildContext context) async {
    final descController = TextEditingController();

    await showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('创建任务'),
        content: TextField(
          controller: descController,
          decoration: const InputDecoration(
            labelText: '任务描述',
            hintText: '输入任务描述...',
          ),
          maxLines: 3,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          ElevatedButton(
            onPressed: () {
              if (descController.text.isNotEmpty) {
                ref
                    .read(agentsProvider.notifier)
                    .createTask(descController.text);
                Navigator.pop(context);
              }
            },
            child: const Text('创建'),
          ),
        ],
      ),
    );
  }
}

class _AgentCard extends StatelessWidget {
  final Agent agent;
  final VoidCallback onToggle;

  const _AgentCard({
    required this.agent,
    required this.onToggle,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                CircleAvatar(
                  backgroundColor:
                      agent.isActive ? Colors.green : Colors.grey[300],
                  child: Icon(
                    Icons.smart_toy,
                    color: agent.isActive ? Colors.white : Colors.grey,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        agent.name,
                        style: const TextStyle(
                          fontWeight: FontWeight.bold,
                          fontSize: 16,
                        ),
                      ),
                      Text(
                        _getRoleName(agent.role),
                        style: TextStyle(
                          color: Colors.grey[600],
                          fontSize: 12,
                        ),
                      ),
                    ],
                  ),
                ),
                Switch(
                  value: agent.isActive,
                  onChanged: (_) => onToggle(),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Text(
              agent.description,
              style: TextStyle(color: Colors.grey[700]),
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 4,
              runSpacing: 4,
              children: agent.skills.map((skill) {
                return Chip(
                  label: Text(skill),
                  labelStyle: const TextStyle(fontSize: 10),
                  padding: EdgeInsets.zero,
                  materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                );
              }).toList(),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Icon(Icons.speed, size: 14, color: Colors.grey[500]),
                const SizedBox(width: 4),
                Text(
                  '优先级: ${agent.priority}',
                  style: TextStyle(
                    color: Colors.grey[500],
                    fontSize: 12,
                  ),
                ),
                const SizedBox(width: 16),
                Icon(Icons.work, size: 14, color: Colors.grey[500]),
                const SizedBox(width: 4),
                Text(
                  '并发: ${agent.maxConcurrentTasks}',
                  style: TextStyle(
                    color: Colors.grey[500],
                    fontSize: 12,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  String _getRoleName(AgentRole role) {
    switch (role) {
      case AgentRole.coordinator:
        return '协调器';
      case AgentRole.researcher:
        return '研究员';
      case AgentRole.coder:
        return '程序员';
      case AgentRole.reviewer:
        return '审查员';
      case AgentRole.executor:
        return '执行器';
      case AgentRole.specialist:
        return '专家';
    }
  }
}

class _TaskCard extends StatelessWidget {
  final AgentTask task;
  final List<Agent> agents;
  final Function(String) onAssign;
  final Function(String) onComplete;

  const _TaskCard({
    required this.task,
    required this.agents,
    required this.onAssign,
    required this.onComplete,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                _getStatusIcon(),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    task.description,
                    style: const TextStyle(fontSize: 16),
                  ),
                ),
                _getStatusBadge(),
              ],
            ),
            const SizedBox(height: 12),
            if (task.assignedAgentId != null)
              Row(
                children: [
                  const Icon(Icons.person, size: 16, color: Colors.grey),
                  const SizedBox(width: 4),
                  Text(
                    '分配给: ${_getAgentName(task.assignedAgentId!)}',
                    style: TextStyle(color: Colors.grey[600], fontSize: 12),
                  ),
                ],
              ),
            const SizedBox(height: 8),
            Row(
              children: [
                Icon(Icons.access_time, size: 14, color: Colors.grey[500]),
                const SizedBox(width: 4),
                Text(
                  _formatDate(task.createdAt),
                  style: TextStyle(
                    color: Colors.grey[500],
                    fontSize: 12,
                  ),
                ),
                const Spacer(),
                if (task.status == 'pending')
                  TextButton.icon(
                    onPressed: () => _showAssignDialog(context),
                    icon: const Icon(Icons.person_add, size: 16),
                    label: const Text('分配'),
                  ),
                if (task.status == 'running')
                  TextButton.icon(
                    onPressed: () => onComplete('任务完成'),
                    icon: const Icon(Icons.check, size: 16),
                    label: const Text('完成'),
                  ),
              ],
            ),
            if (task.result != null) ...[
              const Divider(),
              Text(
                '结果: ${task.result}',
                style: TextStyle(color: Colors.grey[700]),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _getStatusIcon() {
    switch (task.status) {
      case 'pending':
        return const Icon(Icons.pending, color: Colors.orange);
      case 'running':
        return const Icon(Icons.play_circle, color: Colors.blue);
      case 'completed':
        return const Icon(Icons.check_circle, color: Colors.green);
      case 'failed':
        return const Icon(Icons.error, color: Colors.red);
      default:
        return const Icon(Icons.help, color: Colors.grey);
    }
  }

  Widget _getStatusBadge() {
    Color color;
    String label;
    switch (task.status) {
      case 'pending':
        color = Colors.orange;
        label = '待处理';
        break;
      case 'running':
        color = Colors.blue;
        label = '进行中';
        break;
      case 'completed':
        color = Colors.green;
        label = '已完成';
        break;
      case 'failed':
        color = Colors.red;
        label = '失败';
        break;
      default:
        color = Colors.grey;
        label = '未知';
    }
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: color,
          fontSize: 12,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }

  String _getAgentName(String id) {
    return agents.firstWhere(
      (a) => a.id == id,
      orElse: () => const Agent(
        id: '',
        name: '未知',
        description: '',
        role: AgentRole.specialist,
      ),
    ).name;
  }

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final diff = now.difference(date);

    if (diff.inDays > 0) {
      return '${diff.inDays} 天前';
    } else if (diff.inHours > 0) {
      return '${diff.inHours} 小时前';
    } else if (diff.inMinutes > 0) {
      return '${diff.inMinutes} 分钟前';
    } else {
      return '刚刚';
    }
  }

  Future<void> _showAssignDialog(BuildContext context) async {
    String? selectedAgentId;

    await showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('分配任务'),
        content: DropdownButtonFormField<String>(
          decoration: const InputDecoration(
            labelText: '选择代理',
          ),
          items: agents.where((a) => a.isActive).map((agent) {
            return DropdownMenuItem(
              value: agent.id,
              child: Text(agent.name),
            );
          }).toList(),
          onChanged: (value) {
            selectedAgentId = value;
          },
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          ElevatedButton(
            onPressed: () {
              if (selectedAgentId != null) {
                onAssign(selectedAgentId!);
                Navigator.pop(context);
              }
            },
            child: const Text('分配'),
          ),
        ],
      ),
    );
  }
}

class _StatCard extends StatelessWidget {
  final IconData icon;
  final String label;
  final int count;
  final Color color;

  const _StatCard({
    required this.icon,
    required this.label,
    required this.count,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            Icon(icon, color: color, size: 32),
            const SizedBox(height: 8),
            Text(
              '$count',
              style: TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.bold,
                color: color,
              ),
            ),
            Text(
              label,
              style: TextStyle(
                color: Colors.grey[600],
                fontSize: 12,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
