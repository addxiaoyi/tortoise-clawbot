// Chat Screen

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:tortoise/features/chat/bloc/chat_bloc.dart';
import 'package:tortoise/features/chat/widgets/message_bubble.dart';
import 'package:tortoise/features/chat/widgets/input_bar.dart';
import 'package:tortoise/features/chat/widgets/typing_indicator.dart';
import 'package:tortoise/shared/widgets/loading_indicator.dart';

/// Chat Screen
class ChatScreen extends ConsumerStatefulWidget {
  const ChatScreen({super.key});

  @override
  ConsumerState<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends ConsumerState<ChatScreen> {
  final ScrollController _scrollController = ScrollController();
  final TextEditingController _inputController = TextEditingController();
  final FocusNode _inputFocus = FocusNode();

  @override
  void dispose() {
    _scrollController.dispose();
    _inputController.dispose();
    _inputFocus.dispose();
    super.dispose();
  }

  void _scrollToBottom() {
    if (_scrollController.hasClients) {
      _scrollController.animateTo(
        _scrollController.position.maxScrollExtent,
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeOut,
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final chatState = ref.watch(chatBlocProvider);
    
    return Scaffold(
      appBar: AppBar(
        title: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.primaryContainer,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(
                Icons.smart_toy_outlined,
                color: Theme.of(context).colorScheme.primary,
              ),
            ),
            const SizedBox(width: 12),
            const Text('Tortoise'),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.more_vert),
            onPressed: () => _showChatOptions(context),
          ),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: chatState.isLoading && chatState.messages.isEmpty
                ? const Center(child: LoadingIndicator())
                : ListView.builder(
                    controller: _scrollController,
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                    itemCount: chatState.messages.length + (chatState.isTyping ? 1 : 0),
                    itemBuilder: (context, index) {
                      if (chatState.isTyping && index == chatState.messages.length) {
                        return const Padding(
                          padding: EdgeInsets.only(top: 8),
                          child: TypingIndicator(),
                        );
                      }
                      final message = chatState.messages[index];
                      return MessageBubble(
                        message: message,
                        onRetry: () => ref.read(chatBlocProvider.notifier).retry(message.id),
                        onCopy: () => _copyMessage(message.content),
                        onDelete: () => ref.read(chatBlocProvider.notifier).deleteMessage(message.id),
                      );
                    },
                  ),
          ),
          InputBar(
            controller: _inputController,
            focusNode: _inputFocus,
            onSend: (text) {
              ref.read(chatBlocProvider.notifier).sendMessage(text);
              _inputController.clear();
              _scrollToBottom();
            },
            onAttach: _showAttachmentOptions,
            onMic: () => _startVoiceInput(),
            enabled: !chatState.isLoading,
          ),
        ],
      ),
    );
  }

  void _showChatOptions(BuildContext context) {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.person_outline),
              title: const Text('Switch Agent'),
              onTap: () => Navigator.pop(context),
            ),
            ListTile(
              leading: const Icon(Icons.history),
              title: const Text('Chat History'),
              onTap: () => Navigator.pop(context),
            ),
            ListTile(
              leading: const Icon(Icons.delete_outline),
              title: const Text('Clear Chat'),
              onTap: () {
                ref.read(chatBlocProvider.notifier).clearChat();
                Navigator.pop(context);
              },
            ),
          ],
        ),
      ),
    );
  }

  void _showAttachmentOptions() {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.image_outlined),
              title: const Text('Image'),
              onTap: () => Navigator.pop(context),
            ),
            ListTile(
              leading: const Icon(Icons.insert_drive_file_outlined),
              title: const Text('File'),
              onTap: () => Navigator.pop(context),
            ),
            ListTile(
              leading: const Icon(Icons.code),
              title: const Text('Code Block'),
              onTap: () => Navigator.pop(context),
            ),
          ],
        ),
      ),
    );
  }

  void _copyMessage(String content) {
    // TODO: Implement copy
  }

  void _startVoiceInput() {
    // TODO: Implement voice input
  }
}
