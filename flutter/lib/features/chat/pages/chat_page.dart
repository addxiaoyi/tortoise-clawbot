import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/chat_models.dart';
import '../providers/chat_provider.dart';

class ChatPage extends ConsumerStatefulWidget {
  const ChatPage({super.key});

  @override
  ConsumerState<ChatPage> createState() => _ChatPageState();
}

class _ChatPageState extends ConsumerState<ChatPage> {
  final _textController = TextEditingController();
  final _scrollController = ScrollController();
  bool _isLoading = false;

  @override
  void dispose() {
    _textController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final sessions = ref.watch(sessionsProvider);
    final activeId = ref.watch(activeSessionIdProvider) ?? 'default';
    
    return Scaffold(
      appBar: AppBar(
        title: Text(_getSessionTitle(sessions, activeId)),
        actions: [
          IconButton(
            icon: const Icon(Icons.add),
            onPressed: _createNewSession,
          ),
        ],
      ),
      body: Column(
        children: [
          Expanded(child: _buildMessagesList(sessions, activeId)),
          _buildInputArea(),
        ],
      ),
    );
  }

  String _getSessionTitle(List<ChatSession> sessions, String activeId) {
    final session = sessions.where((s) => s.id == activeId).firstOrNull;
    return session?.title ?? '聊天';
  }

  Widget _buildMessagesList(List<ChatSession> sessions, String activeId) {
    final session = sessions.where((s) => s.id == activeId).firstOrNull;
    final messages = session?.messages ?? [];
    
    if (messages.isEmpty) {
      return const Center(
        child: Text('开始对话吧！'),
      );
    }

    return ListView.builder(
      controller: _scrollController,
      padding: const EdgeInsets.all(16),
      itemCount: messages.length,
      itemBuilder: (context, index) {
        final message = messages[index];
        return _buildMessageBubble(message);
      },
    );
  }

  Widget _buildMessageBubble(ChatMessage message) {
    final isUser = message.role == 'user';
    
    return Align(
      alignment: isUser ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.symmetric(vertical: 4),
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: isUser ? Colors.blue : Colors.grey[300],
          borderRadius: BorderRadius.circular(12),
        ),
        child: Text(
          message.content,
          style: TextStyle(color: isUser ? Colors.white : Colors.black),
        ),
      ),
    );
  }

  Widget _buildInputArea() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Theme.of(context).cardColor,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 4,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: Row(
        children: [
          Expanded(
            child: TextField(
              controller: _textController,
              decoration: const InputDecoration(
                hintText: '输入消息...',
                border: OutlineInputBorder(),
              ),
              onSubmitted: (_) => _sendMessage(),
            ),
          ),
          const SizedBox(width: 8),
          IconButton(
            icon: _isLoading 
              ? const SizedBox(
                  width: 24,
                  height: 24,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Icon(Icons.send),
            onPressed: _isLoading ? null : _sendMessage,
          ),
        ],
      ),
    );
  }

  void _createNewSession() {
    ref.read(sessionsProvider.notifier).createSession(title: '新会话');
  }

  Future<void> _sendMessage() async {
    final content = _textController.text.trim();
    if (content.isEmpty) return;

    setState(() => _isLoading = true);
    _textController.clear();

    try {
      final activeId = ref.read(activeSessionIdProvider) ?? 'default';
      final sessions = ref.read(sessionsProvider);
      final session = sessions.where((s) => s.id == activeId).firstOrNull;
      final model = session?.model;

      // 添加用户消息
      ref.read(sessionsProvider.notifier).updateSession(
        ChatSession(
          id: activeId,
          title: session?.title ?? '新会话',
          messages: [
            ...(session?.messages ?? []),
            ChatMessage(
              id: DateTime.now().millisecondsSinceEpoch.toString(),
              sessionId: activeId,
              role: 'user',
              content: content,
              createdAt: DateTime.now(),
              model: model,
            ),
          ],
          createdAt: session?.createdAt ?? DateTime.now(),
          updatedAt: DateTime.now(),
          model: model,
        ),
      );

      // TODO: 调用 AI 服务
      await Future.delayed(const Duration(seconds: 1));
      
      // 添加 AI 响应
      final updatedSession = ref.read(sessionsProvider)
          .where((s) => s.id == activeId).firstOrNull;
      
      if (updatedSession != null) {
        ref.read(sessionsProvider.notifier).updateSession(
          ChatSession(
            id: activeId,
            title: updatedSession.title,
            messages: [
              ...updatedSession.messages,
              ChatMessage(
                id: DateTime.now().millisecondsSinceEpoch.toString(),
                sessionId: activeId,
                role: 'assistant',
                content: '这是一个示例响应',
                createdAt: DateTime.now(),
                model: model,
              ),
            ],
            createdAt: updatedSession.createdAt,
            updatedAt: DateTime.now(),
            model: model,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }
}
