import 'package:flutter/material.dart';

/// 聊天输入框组件
class ChatInput extends StatefulWidget {
  final Function(String) onSend;
  final bool isLoading;
  final VoidCallback? onStop;
  final String? hintText;

  const ChatInput({
    super.key,
    required this.onSend,
    this.isLoading = false,
    this.onStop,
    this.hintText,
  });

  @override
  State<ChatInput> createState() => _ChatInputState();
}

class _ChatInputState extends State<ChatInput> {
  final _controller = TextEditingController();
  final _focusNode = FocusNode();

  @override
  void dispose() {
    _controller.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  void _handleSend() {
    final text = _controller.text.trim();
    if (text.isEmpty) return;
    
    widget.onSend(text);
    _controller.clear();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Theme.of(context).scaffoldBackgroundColor,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 10,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: SafeArea(
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            // 扩展按钮
            IconButton(
              icon: const Icon(Icons.add_circle_outline),
              onPressed: () {
                // 显示扩展菜单
                _showExpandMenu(context);
              },
            ),
            
            // 输入框
            Expanded(
              child: Container(
                constraints: const BoxConstraints(maxHeight: 150),
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.surfaceContainerHighest,
                  borderRadius: BorderRadius.circular(24),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Expanded(
                      child: TextField(
                        controller: _controller,
                        focusNode: _focusNode,
                        maxLines: null,
                        textInputAction: TextInputAction.newline,
                        decoration: InputDecoration(
                          hintText: widget.hintText ?? '输入消息...',
                          border: InputBorder.none,
                          contentPadding: const EdgeInsets.symmetric(
                            horizontal: 20,
                            vertical: 12,
                          ),
                        ),
                        onSubmitted: (_) => _handleSend(),
                      ),
                    ),
                    
                    // 语音输入按钮
                    IconButton(
                      icon: const Icon(Icons.mic),
                      onPressed: () {
                        // 语音输入
                      },
                    ),
                  ],
                ),
              ),
            ),
            
            const SizedBox(width: 8),
            
            // 发送按钮
            widget.isLoading
                ? IconButton(
                    icon: const Icon(Icons.stop_circle, color: Colors.red),
                    onPressed: widget.onStop,
                  )
                : Container(
                    decoration: BoxDecoration(
                      color: Theme.of(context).colorScheme.primary,
                      shape: BoxShape.circle,
                    ),
                    child: IconButton(
                      icon: const Icon(Icons.send, color: Colors.white),
                      onPressed: _handleSend,
                    ),
                  ),
          ],
        ),
      ),
    );
  }

  void _showExpandMenu(BuildContext context) {
    showModalBottomSheet(
      context: context,
      builder: (context) => Container(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.image),
              title: const Text('发送图片'),
              onTap: () {
                Navigator.pop(context);
                // 选择图片
              },
            ),
            ListTile(
              leading: const Icon(Icons.attach_file),
              title: const Text('发送文件'),
              onTap: () {
                Navigator.pop(context);
                // 选择文件
              },
            ),
            ListTile(
              leading: const Icon(Icons.code),
              title: const Text('插入代码'),
              onTap: () {
                Navigator.pop(context);
                // 插入代码
              },
            ),
          ],
        ),
      ),
    );
  }
}
