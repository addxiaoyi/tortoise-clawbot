// Input Bar Widget

import 'package:flutter/material.dart';

class InputBar extends StatelessWidget {
  final TextEditingController controller;
  final FocusNode focusNode;
  final Function(String) onSend;
  final VoidCallback? onAttach;
  final VoidCallback? onMic;
  final bool enabled;

  const InputBar({
    super.key,
    required this.controller,
    required this.focusNode,
    required this.onSend,
    this.onAttach,
    this.onMic,
    this.enabled = true,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: SafeArea(
        child: Row(
          children: [
            if (onAttach != null)
              IconButton(
                icon: const Icon(Icons.attach_file),
                onPressed: enabled ? onAttach : null,
              ),
            Expanded(
              child: TextField(
                controller: controller,
                focusNode: focusNode,
                enabled: enabled,
                decoration: const InputDecoration(
                  hintText: 'Type a message...',
                  border: InputBorder.none,
                  contentPadding: EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                ),
                textInputAction: TextInputAction.send,
                onSubmitted: (value) {
                  if (value.isNotEmpty) {
                    onSend(value);
                  }
                },
              ),
            ),
            if (onMic != null)
              IconButton(
                icon: const Icon(Icons.mic),
                onPressed: enabled ? onMic : null,
              ),
            IconButton(
              icon: Icon(
                Icons.send,
                color: enabled 
                    ? Theme.of(context).colorScheme.primary 
                    : Colors.grey,
              ),
              onPressed: enabled
                  ? () {
                      if (controller.text.isNotEmpty) {
                        onSend(controller.text);
                      }
                    }
                  : null,
            ),
          ],
        ),
      ),
    );
  }
}
