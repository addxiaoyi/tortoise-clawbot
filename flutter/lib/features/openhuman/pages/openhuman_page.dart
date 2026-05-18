import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/neocortex/openhuman_provider.dart';
import '../../../core/neocortex/neocortex_memory.dart';
import '../../../core/neocortex/subconscious_loop.dart';
import '../../../core/neocortex/voice_tools.dart';

/// OpenHuman 核心功能页面
class OpenHumanPage extends ConsumerStatefulWidget {
  const OpenHumanPage({super.key});

  @override
  ConsumerState<OpenHumanPage> createState() => _OpenHumanPageState();
}

class _OpenHumanPageState extends ConsumerState<OpenHumanPage>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('OpenHuman Core'),
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(text: 'Neocortex'),
            Tab(text: 'Subconscious'),
            Tab(text: 'Screen'),
            Tab(text: 'Voice'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: const [
          _NeocortexTab(),
          _SubconsciousTab(),
          _ScreenTab(),
          _VoiceTab(),
        ],
      ),
    );
  }
}

/// Neocortex 记忆层级标签页
class _NeocortexTab extends ConsumerWidget {
  const _NeocortexTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final memories = ref.watch(filteredMemoriesProvider);
    final selectedLayer = ref.watch(memoryLayerFilterProvider);

    return Column(
      children: [
        // 层级筛选
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              _FilterChip(
                label: 'All',
                isSelected: selectedLayer == null,
                onTap: () => ref.read(memoryLayerFilterProvider.notifier).state = null,
              ),
              const SizedBox(width: 8),
              _FilterChip(
                label: 'Working',
                isSelected: selectedLayer == NeocortexMemory.layerWorking,
                onTap: () => ref.read(memoryLayerFilterProvider.notifier).state =
                    NeocortexMemory.layerWorking,
              ),
              const SizedBox(width: 8),
              _FilterChip(
                label: 'Short Term',
                isSelected: selectedLayer == NeocortexMemory.layerShortTerm,
                onTap: () => ref.read(memoryLayerFilterProvider.notifier).state =
                    NeocortexMemory.layerShortTerm,
              ),
              const SizedBox(width: 8),
              _FilterChip(
                label: 'Long Term',
                isSelected: selectedLayer == NeocortexMemory.layerLongTerm,
                onTap: () => ref.read(memoryLayerFilterProvider.notifier).state =
                    NeocortexMemory.layerLongTerm,
              ),
            ],
          ),
        ),
        // 记忆列表
        Expanded(
          child: memories.isEmpty
              ? const Center(child: Text('No memories'))
              : ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: memories.length,
                  itemBuilder: (context, index) {
                    final memory = memories[index];
                    return _MemoryCard(memory: memory);
                  },
                ),
        ),
        // 添加记忆按钮
        Padding(
          padding: const EdgeInsets.all(16),
          child: ElevatedButton.icon(
            onPressed: () => _showAddMemoryDialog(context, ref),
            icon: const Icon(Icons.add),
            label: const Text('Add Memory'),
          ),
        ),
      ],
    );
  }

  void _showAddMemoryDialog(BuildContext context, WidgetRef ref) {
    final contentController = TextEditingController();
    String selectedLayer = NeocortexMemory.layerShortTerm;

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Add Memory'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: contentController,
              decoration: const InputDecoration(
                labelText: 'Content',
                border: OutlineInputBorder(),
              ),
              maxLines: 3,
            ),
            const SizedBox(height: 16),
            DropdownButtonFormField<String>(
              value: selectedLayer,
              decoration: const InputDecoration(
                labelText: 'Layer',
                border: OutlineInputBorder(),
              ),
              items: const [
                DropdownMenuItem(
                  value: NeocortexMemory.layerWorking,
                  child: Text('Working'),
                ),
                DropdownMenuItem(
                  value: NeocortexMemory.layerShortTerm,
                  child: Text('Short Term'),
                ),
                DropdownMenuItem(
                  value: NeocortexMemory.layerLongTerm,
                  child: Text('Long Term'),
                ),
              ],
              onChanged: (value) {
                if (value != null) selectedLayer = value;
              },
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () {
              if (contentController.text.isNotEmpty) {
                ref.read(addMemoryProvider.notifier).add(
                      content: contentController.text,
                      layer: selectedLayer,
                    );
                Navigator.pop(context);
              }
            },
            child: const Text('Add'),
          ),
        ],
      ),
    );
  }
}

/// 记忆卡片
class _MemoryCard extends StatelessWidget {
  final NeocortexMemory memory;

  const _MemoryCard({required this.memory});

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
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: _getLayerColor(memory.layer),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    memory.layer,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 12,
                    ),
                  ),
                ),
                const Spacer(),
                Row(
                  children: List.generate(
                    5,
                    (index) => Icon(
                      Icons.star,
                      size: 16,
                      color: index < (memory.importance * 5).round()
                          ? Colors.amber
                          : Colors.grey[300],
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Text(
              memory.content,
              style: const TextStyle(fontSize: 14),
            ),
            const SizedBox(height: 8),
            Text(
              'Access: ${memory.accessCount}',
              style: TextStyle(
                fontSize: 12,
                color: Colors.grey[600],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Color _getLayerColor(String layer) {
    switch (layer) {
      case NeocortexMemory.layerWorking:
        return Colors.orange;
      case NeocortexMemory.layerShortTerm:
        return Colors.blue;
      case NeocortexMemory.layerLongTerm:
        return Colors.green;
      default:
        return Colors.grey;
    }
  }
}

/// Subconscious 自学习循环标签页
class _SubconsciousTab extends ConsumerWidget {
  const _SubconsciousTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final patterns = ref.watch(learnedPatternsProvider);
    final preferences = ref.watch(userPreferencesProvider);

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 偏好卡片
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'User Preferences',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 16),
                  _PreferenceItem(
                    label: 'Activity Level',
                    value: preferences['activity_level'] ?? 'unknown',
                  ),
                  _PreferenceItem(
                    label: 'Preferred Action',
                    value: preferences['preferred_action'] ?? 'none',
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),
          // 学习到的模式
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Learned Patterns',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 16),
                  if (patterns.isEmpty)
                    const Text('No patterns learned yet')
                  else
                    ...patterns.map((p) => _PatternItem(pattern: p)),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _PreferenceItem extends StatelessWidget {
  final String label;
  final String value;

  const _PreferenceItem({required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label),
          Text(
            value,
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
        ],
      ),
    );
  }
}

class _PatternItem extends StatelessWidget {
  final LearnedPattern pattern;

  const _PatternItem({required this.pattern});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            '${pattern.action} in ${pattern.context}',
            style: const TextStyle(fontWeight: FontWeight.w500),
          ),
          const SizedBox(height: 4),
          LinearProgressIndicator(
            value: pattern.confidence,
            backgroundColor: Colors.grey[200],
          ),
          const SizedBox(height: 4),
          Text(
            'Confidence: ${(pattern.confidence * 100).toStringAsFixed(0)}%',
            style: TextStyle(
              fontSize: 12,
              color: Colors.grey[600],
            ),
          ),
        ],
      ),
    );
  }
}

/// 屏幕智能标签页
class _ScreenTab extends ConsumerWidget {
  const _ScreenTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.screen_search_desktop,
            size: 64,
            color: Colors.grey[400],
          ),
          const SizedBox(height: 16),
          const Text(
            'Screen Intelligence',
            style: TextStyle(
              fontSize: 20,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Capture and analyze screen content',
            style: TextStyle(color: Colors.grey[600]),
          ),
          const SizedBox(height: 24),
          ElevatedButton.icon(
            onPressed: () {
              // TODO: 触发屏幕捕获
            },
            icon: const Icon(Icons.camera_alt),
            label: const Text('Capture Screen'),
          ),
        ],
      ),
    );
  }
}

/// 语音标签页
class _VoiceTab extends ConsumerWidget {
  const _VoiceTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final voiceStatus = ref.watch(voiceStatusProvider);

    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          // 状态指示器
          Container(
            width: 120,
            height: 120,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: _getStatusColor(voiceStatus),
            ),
            child: Icon(
              _getStatusIcon(voiceStatus),
              size: 48,
              color: Colors.white,
            ),
          ),
          const SizedBox(height: 24),
          Text(
            _getStatusText(voiceStatus),
            style: const TextStyle(
              fontSize: 20,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 32),
          // 控制按钮
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              ElevatedButton.icon(
                onPressed: () {
                  ref.read(voiceToolsProvider).startListening();
                },
                icon: const Icon(Icons.mic),
                label: const Text('Listen'),
              ),
              const SizedBox(width: 16),
              OutlinedButton.icon(
                onPressed: () {
                  ref.read(voiceToolsProvider).stopListening();
                },
                icon: const Icon(Icons.stop),
                label: const Text('Stop'),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Color _getStatusColor(VoiceStatus status) {
    switch (status) {
      case VoiceStatus.idle:
        return Colors.grey;
      case VoiceStatus.listening:
        return Colors.blue;
      case VoiceStatus.speaking:
        return Colors.green;
    }
  }

  IconData _getStatusIcon(VoiceStatus status) {
    switch (status) {
      case VoiceStatus.idle:
        return Icons.mic_off;
      case VoiceStatus.listening:
        return Icons.mic;
      case VoiceStatus.speaking:
        return Icons.volume_up;
    }
  }

  String _getStatusText(VoiceStatus status) {
    switch (status) {
      case VoiceStatus.idle:
        return 'Idle';
      case VoiceStatus.listening:
        return 'Listening...';
      case VoiceStatus.speaking:
        return 'Speaking...';
    }
  }
}

/// 筛选标签
class _FilterChip extends StatelessWidget {
  final String label;
  final bool isSelected;
  final VoidCallback onTap;

  const _FilterChip({
    required this.label,
    required this.isSelected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        decoration: BoxDecoration(
          color: isSelected
              ? Theme.of(context).primaryColor
              : Colors.grey[200],
          borderRadius: BorderRadius.circular(20),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: isSelected ? Colors.white : Colors.black87,
            fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
          ),
        ),
      ),
    );
  }
}
