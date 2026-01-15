// Message Bubble Widget

import 'package:flutter/material.dart';
import 'package:tortoise/core/theme/app_theme.dart';
import 'package:tortoise/features/chat/bloc/chat_bloc.dart';
import 'package:timeago/timeago.dart' as timeago;

class MessageBubble extends StatelessWidget {
  final Message message;
  final VoidCallback? onRetry;
  final VoidCallback? onCopy;
  final VoidCallback? onDelete;

  const MessageBubble({
    super.key,
    required this.message,
    this.onRetry,
    this.onCopy,
    this.onDelete,
  });

  bool get isUser => message.role == MessageRole.user;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    
    return Align(
      alignment: isUser ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        constraints: BoxConstraints(
          maxWidth: MediaQuery.of(context).size.width * 0.75,
        ),
        margin: const EdgeInsets.symmetric(vertical: 4),
        child: GestureDetector(
          onLongPress: () => _showContextMenu(context),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: isUser 
                  ? (theme.brightness == Brightness.dark 
                      ? AppColors.userBubbleDark 
                      : AppColors.userBubble)
                  : (theme.brightness == Brightness.dark 
                      ? AppColors.assistantBubbleDark 
                      : AppColors.assistantBubble),
              borderRadius: BorderRadius.only(
                topLeft: const Radius.circular(20),
                topRight: const Radius.circular(20),
                bottomLeft: Radius.circular(isUser ? 20 : 4),
                bottomRight: Radius.circular(isUser ? 4 : 20),
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (!isUser && message.agentName != null) ...[
                  Text(
                    message.agentName!,
                    style: theme.textTheme.bodySmall?.copyWith(
                      fontWeight: FontWeight.w600,
                      color: AppColors.primary,
                    ),
                  ),
                  const SizedBox(height: 4),
                ],
                SelectableText(
                  message.content,
                  style: TextStyle(
                    color: isUser ? Colors.white : null,
                    height: 1.4,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  timeago.format(message.timestamp),
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: isUser 
                        ? Colors.white70 
                        : theme.textTheme.bodySmall?.color?.withOpacity(0.6),
                    fontSize: 10,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  void _showContextMenu(BuildContext context) {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.copy),
              title: const Text('Copy'),
              onTap: () {
                onCopy?.call();
                Navigator.pop(context);
              },
            ),
            if (message.status == MessageStatus.error && onRetry != null)
              ListTile(
                leading: const Icon(Icons.refresh),
                title: const Text('Retry'),
                onTap: () {
                  onRetry?.call();
                  Navigator.pop(context);
                },
              ),
            if (onDelete != null)
              ListTile(
                leading: const Icon(Icons.delete_outline),
                title: const Text('Delete'),
                onTap: () {
                  onDelete?.call();
                  Navigator.pop(context);
                },
              ),
          ],
        ),
      ),
    );
  }
}
