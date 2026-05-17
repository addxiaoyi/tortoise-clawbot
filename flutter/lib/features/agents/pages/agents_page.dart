import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/agents_provider.dart';

class AgentsPage extends ConsumerWidget {
  const AgentsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final agentsState = ref.watch(agentsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('多代理管理'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.read(agentsProvider.notifier).refreshAgents(),
          ),
          IconButton(
            icon: const Icon(Icons.add),
            onPressed: () => _showAddAgentDialog(context, ref),
          ),
        ],
      ),
      body: agentsState.isLoading
          ? const Center(child: CircularProgressIndicator())
          : _buildAgentsList(context, ref, agentsState),
    );
  }

  Widget _buildAgentsList(BuildContext context, WidgetRef ref, AgentsState state) {
    if (state.agents.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.smart_toy_outlined, size: 64, color: Colors.grey[400]),
            const SizedBox(height: 16),
            Text(
              '暂无代理',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                color: Colors.grey[600],
              ),
            ),
            const SizedBox(height: 8),
            TextButton.icon(
              onPressed: () => _showAddAgentDialog(context, ref),
              icon: const Icon(Icons.add),
              label: const Text('创建代理'),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: state.agents.length,
      itemBuilder: (context, index) {
        final agent = state.agents[index];
        return _AgentCard(
          agent: agent,
          onTap: () => _showAgentDetails(context, ref, agent),
          onToggle: (enabled) {
            ref.read(agentsProvider.notifier).toggleAgent(agent.id, enabled);
          },
          onDelete: () => _confirmDeleteAgent(context, ref, agent),
        );
      },
    );
  }

  void _showAddAgentDialog(BuildContext context, WidgetRef ref) {
    showDialog(
      context: context,
      builder: (context) => _AddAgentDialog(
        onAdd: (name, model, instructions) {
          ref.read(agentsProvider.notifier).addAgent(name, model, instructions);
        },
      ),
    );
  }

  void _showAgentDetails(BuildContext context, WidgetRef ref, Agent agent) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) => _AgentDetailsSheet(agent: agent),
    );
  }

  void _confirmDeleteAgent(BuildContext context, WidgetRef ref, Agent agent) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('删除代理'),
        content: Text('确定要删除代理 "${agent.name}" 吗？'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () {
              ref.read(agentsProvider.notifier).deleteAgent(agent.id);
              Navigator.pop(context);
            },
            child: const Text('删除'),
          ),
        ],
      ),
    );
  }
}

class _AgentCard extends StatelessWidget {
  final Agent agent;
  final VoidCallback onTap;
  final Function(bool) onToggle;
  final VoidCallback onDelete;

  const _AgentCard({
    required this.agent,
    required this.onTap,
    required this.onToggle,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    width: 48,
                    height: 48,
                    decoration: BoxDecoration(
                      color: _getAgentColor(agent.type).withOpacity(0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Icon(
                      _getAgentIcon(agent.type),
                      color: _getAgentColor(agent.type),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          agent.name,
                          style: Theme.of(context).textTheme.titleMedium,
                        ),
                        const SizedBox(height: 4),
                        Text(
                          agent.model,
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: Colors.grey[600],
                          ),
                        ),
                      ],
                    ),
                  ),
                  Switch(
                    value: agent.enabled,
                    onChanged: onToggle,
                  ),
                  IconButton(
                    icon: const Icon(Icons.delete_outline),
                    onPressed: onDelete,
                    color: Colors.red,
                  ),
                ],
              ),
              if (agent.description.isNotEmpty) ...[
                const SizedBox(height: 12),
                Text(
                  agent.description,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[700],
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
              const SizedBox(height: 12),
              Wrap(
                spacing: 8,
                children: agent.capabilities.map((cap) {
                  return Chip(
                    label: Text(cap, style: const TextStyle(fontSize: 12)),
                    padding: EdgeInsets.zero,
                    materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                  );
                }).toList(),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Color _getAgentColor(String type) {
    switch (type) {
      case 'coordinator':
        return Colors.purple;
      case 'researcher':
        return Colors.blue;
      case 'coder':
        return Colors.green;
      case 'critic':
        return Colors.orange;
      case 'summarizer':
        return Colors.teal;
      default:
        return Colors.grey;
    }
  }

  IconData _getAgentIcon(String type) {
    switch (type) {
      case 'coordinator':
        return Icons.hub;
      case 'researcher':
        return Icons.search;
      case 'coder':
        return Icons.code;
      case 'critic':
        return Icons.rate_review;
      case 'summarizer':
        return Icons.summarize;
      default:
        return Icons.smart_toy;
    }
  }
}

class _AddAgentDialog extends StatefulWidget {
  final Function(String name, String model, String instructions) onAdd;

  const _AddAgentDialog({required this.onAdd});

  @override
  State<_AddAgentDialog> createState() => _AddAgentDialogState();
}

class _AddAgentDialogState extends State<_AddAgentDialog> {
  final _nameController = TextEditingController();
  final _instructionsController = TextEditingController();
  String _selectedModel = 'claude-3-5-sonnet-20241022';
  String _selectedType = 'coordinator';

  final _models = [
    'claude-3-5-sonnet-20241022',
    'claude-3-opus-20240229',
    'gpt-4o',
    'gpt-4-turbo',
  ];

  final _types = [
    ('coordinator', '协调者'),
    ('researcher', '研究者'),
    ('coder', '编码员'),
    ('critic', '评论家'),
    ('summarizer', '总结者'),
  ];

  @override
  void dispose() {
    _nameController.dispose();
    _instructionsController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('创建代理'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: _nameController,
              decoration: const InputDecoration(
                labelText: '代理名称',
                hintText: '输入代理名称',
              ),
            ),
            const SizedBox(height: 16),
            DropdownButtonFormField<String>(
              value: _selectedType,
              decoration: const InputDecoration(
                labelText: '代理类型',
              ),
              items: _types.map((t) {
                return DropdownMenuItem(
                  value: t.$1,
                  child: Text(t.$2),
                );
              }).toList(),
              onChanged: (value) {
                if (value != null) setState(() => _selectedType = value);
              },
            ),
            const SizedBox(height: 16),
            DropdownButtonFormField<String>(
              value: _selectedModel,
              decoration: const InputDecoration(
                labelText: 'AI 模型',
              ),
              items: _models.map((m) {
                return DropdownMenuItem(value: m, child: Text(m));
              }).toList(),
              onChanged: (value) {
                if (value != null) setState(() => _selectedModel = value);
              },
            ),
            const SizedBox(height: 16),
            TextField(
              controller: _instructionsController,
              decoration: const InputDecoration(
                labelText: '指令',
                hintText: '描述代理的行为和职责',
              ),
              maxLines: 3,
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('取消'),
        ),
        ElevatedButton(
          onPressed: () {
            if (_nameController.text.isNotEmpty) {
              widget.onAdd(
                _nameController.text,
                _selectedModel,
                _instructionsController.text,
              );
              Navigator.pop(context);
            }
          },
          child: const Text('创建'),
        ),
      ],
    );
  }
}

class _AgentDetailsSheet extends StatelessWidget {
  final Agent agent;

  const _AgentDetailsSheet({required this.agent});

  @override
  Widget build(BuildContext context) {
    return DraggableScrollableSheet(
      initialChildSize: 0.6,
      minChildSize: 0.3,
      maxChildSize: 0.9,
      expand: false,
      builder: (context, scrollController) {
        return SingleChildScrollView(
          controller: scrollController,
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Center(
                child: Container(
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(
                    color: Colors.grey[300],
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
              ),
              const SizedBox(height: 24),
              Row(
                children: [
                  Container(
                    width: 64,
                    height: 64,
                    decoration: BoxDecoration(
                      color: Colors.blue.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: const Icon(Icons.smart_toy, size: 32),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          agent.name,
                          style: Theme.of(context).textTheme.headlineSmall,
                        ),
                        Text(
                          agent.model,
                          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: Colors.grey[600],
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 24),
              const Divider(),
              const SizedBox(height: 16),
              _buildStatRow(context, '状态', agent.enabled ? '运行中' : '已停止'),
              _buildStatRow(context, '类型', agent.type),
              _buildStatRow(context, '能力', agent.capabilities.length.toString()),
              const SizedBox(height: 24),
              if (agent.description.isNotEmpty) ...[
                Text(
                  '描述',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 8),
                Text(agent.description),
                const SizedBox(height: 24),
              ],
              if (agent.instructions.isNotEmpty) ...[
                Text(
                  '指令',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 8),
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: Colors.grey[100],
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(agent.instructions),
                ),
              ],
            ],
          ),
        );
      },
    );
  }

  Widget _buildStatRow(BuildContext context, String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            label,
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
              color: Colors.grey[600],
            ),
          ),
          Text(
            value,
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }
}
