import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'providers/voice_provider.dart';

/// Voice Wake 页面
class VoiceWakePage extends ConsumerStatefulWidget {
  const VoiceWakePage({super.key});

  @override
  ConsumerState<VoiceWakePage> createState() => _VoiceWakePageState();
}

class _VoiceWakePageState extends ConsumerState<VoiceWakePage> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(voiceWakeProvider.notifier).initialize();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(voiceWakeProvider);
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('语音唤醒'),
        elevation: 0,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.read(voiceWakeProvider.notifier).initialize(),
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 状态卡片
            _buildStatusCard(state, theme),
            const SizedBox(height: 24),

            // 灵敏度设置
            _buildSensitivitySection(state, theme),
            const SizedBox(height: 24),

            // 唤醒词列表
            _buildWakeWordsSection(state, theme),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: state.isListening
            ? () => ref.read(voiceWakeProvider.notifier).stopListening()
            : () => ref.read(voiceWakeProvider.notifier).startListening(),
        icon: Icon(state.isListening ? Icons.stop : Icons.mic),
        label: Text(state.isListening ? '停止监听' : '开始监听'),
        backgroundColor: state.isListening ? Colors.red : theme.colorScheme.primary,
      ),
    );
  }

  Widget _buildStatusCard(VoiceWakeState state, ThemeData theme) {
    Color statusColor;
    String statusText;
    IconData statusIcon;

    switch (state.status) {
      case VoiceWakeStatus.idle:
        statusColor = Colors.grey;
        statusText = '未初始化';
        statusIcon = Icons.power_settings_new;
        break;
      case VoiceWakeStatus.initializing:
        statusColor = Colors.orange;
        statusText = '初始化中...';
        statusIcon = Icons.hourglass_empty;
        break;
      case VoiceWakeStatus.ready:
        statusColor = Colors.green;
        statusText = '就绪';
        statusIcon = Icons.check_circle;
        break;
      case VoiceWakeStatus.listening:
        statusColor = Colors.blue;
        statusText = '监听中';
        statusIcon = Icons.mic;
        break;
      case VoiceWakeStatus.error:
        statusColor = Colors.red;
        statusText = '错误: ${state.error ?? "未知错误"}';
        statusIcon = Icons.error;
        break;
    }

    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: statusColor.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(statusIcon, color: statusColor, size: 32),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    '语音唤醒状态',
                    style: TextStyle(fontSize: 12, color: Colors.grey),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    statusText,
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: statusColor,
                    ),
                  ),
                ],
              ),
            ),
            if (state.isListening)
              _buildPulsingDot(),
          ],
        ),
      ),
    );
  }

  Widget _buildPulsingDot() {
    return Container(
      width: 12,
      height: 12,
      decoration: BoxDecoration(
        color: Colors.blue,
        shape: BoxShape.circle,
        boxShadow: [
          BoxShadow(
            color: Colors.blue.withOpacity(0.5),
            blurRadius: 8,
            spreadRadius: 2,
          ),
        ],
      ),
    );
  }

  Widget _buildSensitivitySection(VoiceWakeState state, ThemeData theme) {
    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.tune, color: theme.colorScheme.primary),
                const SizedBox(width: 8),
                const Text(
                  '唤醒灵敏度',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                const Text('低', style: TextStyle(color: Colors.grey)),
                Expanded(
                  child: Slider(
                    value: state.sensitivity,
                    onChanged: (value) {
                      ref.read(voiceWakeProvider.notifier).setSensitivity(value);
                    },
                    divisions: 10,
                    label: '${(state.sensitivity * 100).round()}%',
                  ),
                ),
                const Text('高', style: TextStyle(color: Colors.grey)),
              ],
            ),
            Center(
              child: Text(
                '当前灵敏度: ${(state.sensitivity * 100).round()}%',
                style: TextStyle(color: theme.colorScheme.primary),
              ),
            ),
            const SizedBox(height: 8),
            Text(
              '灵敏度越高，越容易触发唤醒，但也可能增加误触发',
              style: TextStyle(fontSize: 12, color: Colors.grey[600]),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildWakeWordsSection(VoiceWakeState state, ThemeData theme) {
    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    Icon(Icons.record_voice_over, color: theme.colorScheme.primary),
                    const SizedBox(width: 8),
                    const Text(
                      '唤醒词',
                      style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                    ),
                  ],
                ),
                TextButton.icon(
                  onPressed: () => _showAddWakeWordDialog(),
                  icon: const Icon(Icons.add),
                  label: const Text('添加'),
                ),
              ],
            ),
            const SizedBox(height: 8),
            if (state.wakeWords.isEmpty)
              const Center(
                child: Padding(
                  padding: EdgeInsets.all(32),
                  child: Text(
                    '还没有唤醒词\n点击右上角添加',
                    textAlign: TextAlign.center,
                    style: TextStyle(color: Colors.grey),
                  ),
                ),
              )
            else
              ...state.wakeWords.map((wakeWord) => _buildWakeWordTile(wakeWord, state)),
          ],
        ),
      ),
    );
  }

  Widget _buildWakeWordTile(WakeWord wakeWord, VoiceWakeState state) {
    final isSelected = wakeWord.name == state.currentWakeWord;

    return ListTile(
      leading: CircleAvatar(
        backgroundColor: isSelected
            ? Theme.of(context).colorScheme.primary
            : Colors.grey[300],
        child: Icon(
          Icons.volume_up,
          color: isSelected ? Colors.white : Colors.grey[600],
        ),
      ),
      title: Text(
        wakeWord.name,
        style: TextStyle(
          fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
        ),
      ),
      subtitle: Text('灵敏度: ${(wakeWord.sensitivity * 100).round()}%'),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          IconButton(
            icon: const Icon(Icons.play_arrow),
            onPressed: () {
              ref.read(voiceWakeProvider.notifier).testWakeWord(wakeWord.name);
            },
          ),
          PopupMenuButton<String>(
            onSelected: (value) {
              if (value == 'delete') {
                _confirmDeleteWakeWord(wakeWord);
              } else if (value == 'edit') {
                _showEditWakeWordDialog(wakeWord);
              }
            },
            itemBuilder: (context) => [
              const PopupMenuItem(value: 'edit', child: Text('编辑')),
              const PopupMenuItem(value: 'delete', child: Text('删除')),
            ],
          ),
        ],
      ),
      selected: isSelected,
      selectedTileColor: Theme.of(context).colorScheme.primary.withOpacity(0.1),
      onTap: () {
        ref.read(voiceWakeProvider.notifier).startListening(wakeWord: wakeWord.name);
      },
    );
  }

  void _showAddWakeWordDialog() {
    final nameController = TextEditingController();
    double sensitivity = 0.5;

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('添加唤醒词'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: nameController,
              decoration: const InputDecoration(
                labelText: '唤醒词名称',
                hintText: '例如: Hey Tortoise',
              ),
            ),
            const SizedBox(height: 16),
            StatefulBuilder(
              builder: (context, setState) => Column(
                children: [
                  const Text('灵敏度'),
                  Slider(
                    value: sensitivity,
                    onChanged: (value) => setState(() => sensitivity = value),
                    divisions: 10,
                    label: '${(sensitivity * 100).round()}%',
                  ),
                ],
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          ElevatedButton(
            onPressed: () {
              if (nameController.text.isNotEmpty) {
                ref.read(voiceWakeProvider.notifier).addWakeWord(
                  nameController.text,
                  sensitivity: sensitivity,
                );
                Navigator.pop(context);
              }
            },
            child: const Text('添加'),
          ),
        ],
      ),
    );
  }

  void _showEditWakeWordDialog(WakeWord wakeWord) {
    final nameController = TextEditingController(text: wakeWord.name);
    double sensitivity = wakeWord.sensitivity;

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('编辑唤醒词'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: nameController,
              decoration: const InputDecoration(labelText: '唤醒词名称'),
            ),
            const SizedBox(height: 16),
            StatefulBuilder(
              builder: (context, setState) => Column(
                children: [
                  const Text('灵敏度'),
                  Slider(
                    value: sensitivity,
                    onChanged: (value) => setState(() => sensitivity = value),
                    divisions: 10,
                    label: '${(sensitivity * 100).round()}%',
                  ),
                ],
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          ElevatedButton(
            onPressed: () {
              ref.read(voiceWakeProvider.notifier).updateWakeWord(
                wakeWord.id,
                name: nameController.text,
                sensitivity: sensitivity,
              );
              Navigator.pop(context);
            },
            child: const Text('保存'),
          ),
        ],
      ),
    );
  }

  void _confirmDeleteWakeWord(WakeWord wakeWord) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('删除唤醒词'),
        content: Text('确定要删除 "${wakeWord.name}" 吗？'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () {
              ref.read(voiceWakeProvider.notifier).removeWakeWord(wakeWord.id);
              Navigator.pop(context);
            },
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('删除'),
          ),
        ],
      ),
    );
  }
}
