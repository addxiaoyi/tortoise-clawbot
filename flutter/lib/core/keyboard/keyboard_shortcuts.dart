import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

/// 快捷键预定义
class KeyboardShortcuts {
  static const newChatShortcut = SingleActivator(LogicalKeyboardKey.keyN, control: true);
  static const searchShortcut = SingleActivator(LogicalKeyboardKey.keyF, control: true);
  static const settingsShortcut = SingleActivator(LogicalKeyboardKey.comma, control: true);
  static const sendMessageShortcut = SingleActivator(LogicalKeyboardKey.enter, control: true);
  static const toggleThemeShortcut = SingleActivator(LogicalKeyboardKey.keyD, control: true);
}

/// 快捷键 Intent
class CallbackIntent extends Intent {
  final VoidCallback callback;
  const CallbackIntent(this.callback);
}

/// 快捷键 Action
class CallbackAction extends Action<CallbackIntent> {
  CallbackAction();

  @override
  Object? invoke(CallbackIntent intent) {
    intent.callback();
    return null;
  }
}

/// 快捷键包装组件
class KeyboardShortcutsWrapper extends StatelessWidget {
  final Widget child;
  final Map<ShortcutActivator, VoidCallback> shortcuts;

  const KeyboardShortcutsWrapper({
    super.key,
    required this.child,
    required this.shortcuts,
  });

  @override
  Widget build(BuildContext context) {
    final intentMap = shortcuts.map(
      (key, value) => MapEntry(key, CallbackIntent(value)),
    );

    return Shortcuts(
      shortcuts: intentMap,
      child: Actions(
        actions: {
          CallbackIntent: CallbackAction(),
        },
        child: Focus(
          autofocus: true,
          child: child,
        ),
      ),
    );
  }
}

/// 快捷键提示小部件
class ShortcutHint extends StatelessWidget {
  final String label;
  final Widget child;

  const ShortcutHint({
    super.key,
    required this.label,
    required this.child,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        child,
        const SizedBox(width: 8),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
          decoration: BoxDecoration(
            color: Theme.of(context).colorScheme.surfaceContainerHighest,
            borderRadius: BorderRadius.circular(4),
          ),
          child: Text(
            label,
            style: Theme.of(context).textTheme.labelSmall,
          ),
        ),
      ],
    );
  }
}
