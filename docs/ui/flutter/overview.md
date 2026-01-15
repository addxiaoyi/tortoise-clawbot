# Tortoise Flutter UI 设计

## 概述

Tortoise UI 采用 Flutter 实现，提供跨平台（iOS、Android、macOS、Windows、Linux、Web）的统一用户体验。

## 项目结构

```
ui/
├── lib/
│   ├── main.dart
│   ├── app.dart
│   ├── core/
│   │   ├── theme/
│   │   │   ├── app_theme.dart
│   │   │   ├── colors.dart
│   │   │   └── typography.dart
│   │   ├── routing/
│   │   │   └── app_router.dart
│   │   └── utils/
│   │       └── extensions.dart
│   ├── features/
│   │   ├── chat/
│   │   │   ├── chat_screen.dart
│   │   │   ├── widgets/
│   │   │   │   ├── message_bubble.dart
│   │   │   │   ├── input_bar.dart
│   │   │   │   └── typing_indicator.dart
│   │   │   └── bloc/
│   │   │       └── chat_bloc.dart
│   │   ├── settings/
│   │   │   ├── settings_screen.dart
│   │   │   └── widgets/
│   │   │       ├── model_selector.dart
│   │   │       ├── channel_list.dart
│   │   │       └── skill_manager.dart
│   │   ├── memory/
│   │   │   ├── memory_screen.dart
│   │   │   └── widgets/
│   │   │       └── memory_card.dart
│   │   └── plugins/
│   │       ├── plugins_screen.dart
│   │       └── widgets/
│   │           └── plugin_card.dart
│   ├── shared/
│   │   ├── widgets/
│   │   │   ├── loading_indicator.dart
│   │   │   └── error_view.dart
│   │   └── services/
│   │       ├── api_service.dart
│   │       └── websocket_service.dart
│   └── generated/
│       └── gpt.dart
├── assets/
│   ├── icons/
│   └── images/
├── test/
└── pubspec.yaml
```

## 核心组件

### 主应用入口

```dart
// lib/main.dart

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:tortoise/app.dart';
import 'package:tortoise/core/services/logger.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  
  // 初始化日志
  await Logger.initialize();
  
  // 设置系统 UI
  SystemChrome.setSystemUIOverlayStyle(
    const SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.dark,
    ),
  );
  
  // 启用响应式布局
  await SystemChrome.setPreferredOrientations([
    DeviceOrientation.portraitUp,
    DeviceOrientation.portraitDown,
    DeviceOrientation.landscapeLeft,
    DeviceOrientation.landscapeRight,
  ]);
  
  runApp(const TortoiseApp());
}
```

### 应用主体

```dart
// lib/app.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:tortoise/core/theme/app_theme.dart';
import 'package:tortoise/core/routing/app_router.dart';

class TortoiseApp extends ConsumerWidget {
  const TortoiseApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(appRouterProvider);
    
    return MaterialApp.router(
      title: 'Tortoise',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme,
      darkTheme: AppTheme.darkTheme,
      themeMode: ThemeMode.system,
      routerConfig: router,
    );
  }
}
```

### 主题系统

```dart
// lib/core/theme/app_theme.dart

import 'package:flutter/material.dart';
import 'colors.dart';
import 'typography.dart';

class AppTheme {
  AppTheme._();

  static ThemeData get lightTheme => ThemeData(
    useMaterial3: true,
    brightness: Brightness.light,
    colorScheme: ColorScheme.light(
      primary: AppColors.primary,
      secondary: AppColors.secondary,
      tertiary: AppColors.tertiary,
      surface: AppColors.surface,
      error: AppColors.error,
    ),
    scaffoldBackgroundColor: AppColors.background,
    appBarTheme: const AppBarTheme(
      backgroundColor: AppColors.surface,
      foregroundColor: AppColors.onSurface,
      elevation: 0,
      centerTitle: true,
    ),
    cardTheme: CardTheme(
      color: AppColors.surface,
      elevation: 2,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
      ),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: AppColors.surfaceVariant,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(24),
        borderSide: BorderSide.none,
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(24),
        borderSide: BorderSide.none,
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(24),
        borderSide: const BorderSide(color: AppColors.primary, width: 2),
      ),
      contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
    ),
    elevatedButtonTheme: ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: AppColors.primary,
        foregroundColor: AppColors.onPrimary,
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(24),
        ),
      ),
    ),
    textTheme: AppTypography.textTheme,
  );

  static ThemeData get darkTheme => ThemeData(
    useMaterial3: true,
    brightness: Brightness.dark,
    colorScheme: ColorScheme.dark(
      primary: AppColors.primary,
      secondary: AppColors.secondary,
      tertiary: AppColors.tertiary,
      surface: AppColors.darkSurface,
      error: AppColors.error,
    ),
    scaffoldBackgroundColor: AppColors.darkBackground,
    appBarTheme: const AppBarTheme(
      backgroundColor: AppColors.darkSurface,
      foregroundColor: AppColors.darkOnSurface,
      elevation: 0,
      centerTitle: true,
    ),
    cardTheme: CardTheme(
      color: AppColors.darkSurface,
      elevation: 2,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
      ),
    ),
    textTheme: AppTypography.darkTextTheme,
  );
}
```

### 颜色系统

```dart
// lib/core/theme/colors.dart

import 'package:flutter/material.dart';

class AppColors {
  AppColors._();

  // 主色调
  static const Color primary = Color(0xFF6366F1);      // Indigo
  static const Color secondary = Color(0xFF8B5CF6);    // Violet
  static const Color tertiary = Color(0xFF14B8A6);     // Teal

  // 功能色
  static const Color success = Color(0xFF22C55E);
  static const Color warning = Color(0xFFF59E0B);
  static const Color error = Color(0xFFEF4444);
  static const Color info = Color(0xFF3B82F6);

  // 浅色主题
  static const Color background = Color(0xFFF8FAFC);
  static const Color surface = Color(0xFFFFFFFF);
  static const Color surfaceVariant = Color(0xFFF1F5F9);
  static const Color onSurface = Color(0xFF1E293B);
  static const Color onSurfaceVariant = Color(0xFF64748B);

  // 深色主题
  static const Color darkBackground = Color(0xFF0F172A);
  static const Color darkSurface = Color(0xFF1E293B);
  static const Color darkSurfaceVariant = Color(0xFF334155);
  static const Color darkOnSurface = Color(0xFFF8FAFC);
  static const Color darkOnSurfaceVariant = Color(0xFF94A3B8);

  // 消息气泡颜色
  static const Color userBubble = Color(0xFF6366F1);
  static const Color userBubbleDark = Color(0xFF4F46E5);
  static const Color assistantBubble = Color(0xFFE2E8F0);
  static const Color assistantBubbleDark = Color(0xFF334155);

  // 代理身份颜色
  static const Color orchestratorColor = Color(0xFF8B5CF6);
  static const Color specialistColor = Color(0xFF14B8A6);
  static const Color researcherColor = Color(0xFFF59E0B);
  static const Color coderColor = Color(0xFF22C55E);
  static const Color criticColor = Color(0xFFEF4444);
}
```

### 路由系统

```dart
// lib/core/routing/app_router.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:tortoise/features/chat/chat_screen.dart';
import 'package:tortoise/features/settings/settings_screen.dart';
import 'package:tortoise/features/memory/memory_screen.dart';
import 'package:tortoise/features/plugins/plugins_screen.dart';
import 'package:tortoise/shared/widgets/shell_scaffold.dart';

final appRouterProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/chat',
    routes: [
      ShellRoute(
        builder: (context, state, child) => ShellScaffold(child: child),
        routes: [
          GoRoute(
            path: '/chat',
            name: 'chat',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: ChatScreen(),
            ),
          ),
          GoRoute(
            path: '/memory',
            name: 'memory',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: MemoryScreen(),
            ),
          ),
          GoRoute(
            path: '/plugins',
            name: 'plugins',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: PluginsScreen(),
            ),
          ),
          GoRoute(
            path: '/settings',
            name: 'settings',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: SettingsScreen(),
            ),
          ),
        ],
      ),
    ],
  );
});
```

### 聊天界面

```dart
// lib/features/chat/chat_screen.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:tortoise/features/chat/bloc/chat_bloc.dart';
import 'package:tortoise/features/chat/widgets/message_bubble.dart';
import 'package:tortoise/features/chat/widgets/input_bar.dart';
import 'package:tortoise/features/chat/widgets/typing_indicator.dart';
import 'package:tortoise/shared/widgets/loading_indicator.dart';

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
```

### 消息气泡

```dart
// lib/features/chat/widgets/message_bubble.dart

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:tortoise/core/theme/colors.dart';
import 'package:tortoise/core/models/message.dart';
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
        child: Column(
          crossAxisAlignment: isUser ? CrossAxisAlignment.end : CrossAxisAlignment.start,
          children: [
            if (!isUser) ...[
              _buildAgentHeader(context),
              const SizedBox(height: 4),
            ],
            GestureDetector(
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
                    if (message.isStreaming)
                      _buildStreamingContent(context)
                    else
                      _buildContent(context),
                    const SizedBox(height: 4),
                    Text(
                      timeago.format(DateTime.fromMillisecondsSinceEpoch(message.timestamp)),
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
          ],
        ),
      ),
    );
  }

  Widget _buildAgentHeader(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          padding: const EdgeInsets.all(4),
          decoration: BoxDecoration(
            color: _getAgentColor().withOpacity(0.2),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(
            _getAgentIcon(),
            size: 14,
            color: _getAgentColor(),
          ),
        ),
        const SizedBox(width: 8),
        Text(
          message.agentName ?? 'Tortoise',
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
            fontWeight: FontWeight.w600,
            color: _getAgentColor(),
          ),
        ),
      ],
    );
  }

  Widget _buildContent(BuildContext context) {
    return SelectableText(
      message.content,
      style: TextStyle(
        color: isUser ? Colors.white : null,
        height: 1.4,
      ),
    );
  }

  Widget _buildStreamingContent(BuildContext context) {
    return _buildContent(context);
  }

  Color _getAgentColor() {
    switch (message.agentRole) {
      case AgentRole.orchestrator:
        return AppColors.orchestratorColor;
      case AgentRole.specialist:
        return AppColors.specialistColor;
      case AgentRole.researcher:
        return AppColors.researcherColor;
      case AgentRole.coder:
        return AppColors.coderColor;
      case AgentRole.critic:
        return AppColors.criticColor;
      default:
        return AppColors.primary;
    }
  }

  IconData _getAgentIcon() {
    switch (message.agentRole) {
      case AgentRole.orchestrator:
        return Icons.hub;
      case AgentRole.specialist:
        return Icons.psychology;
      case AgentRole.researcher:
        return Icons.science;
      case AgentRole.coder:
        return Icons.code;
      case AgentRole.critic:
        return Icons.rate_review;
      default:
        return Icons.smart_toy;
    }
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
                Clipboard.setData(ClipboardData(text: message.content));
                Navigator.pop(context);
              },
            ),
            if (message.status == MessageStatus.error && onRetry != null)
              ListTile(
                leading: const Icon(Icons.refresh),
                title: const Text('Retry'),
                onTap: () {
                  onRetry!();
                  Navigator.pop(context);
                },
              ),
            if (onDelete != null)
              ListTile(
                leading: const Icon(Icons.delete_outline),
                title: const Text('Delete'),
                onTap: () {
                  onDelete!();
                  Navigator.pop(context);
                },
              ),
          ],
        ),
      ),
    );
  }
}
```

### Chat BLoC

```dart
// lib/features/chat/bloc/chat_bloc.dart

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:tortoise/core/models/message.dart';
import 'package:tortoise/core/services/api_service.dart';

enum ChatStatus { idle, loading, error }

class ChatState {
  final List<Message> messages;
  final bool isLoading;
  final bool isTyping;
  final String? error;

  const ChatState({
    this.messages = const [],
    this.isLoading = false,
    this.isTyping = false,
    this.error,
  });

  ChatState copyWith({
    List<Message>? messages,
    bool? isLoading,
    bool? isTyping,
    String? error,
  }) {
    return ChatState(
      messages: messages ?? this.messages,
      isLoading: isLoading ?? this.isLoading,
      isTyping: isTyping ?? this.isTyping,
      error: error,
    );
  }
}

class ChatBloc extends StateNotifier<ChatState> {
  final ApiService _api;

  ChatBloc(this._api) : super(const ChatState());

  Future<void> sendMessage(String content) async {
    if (content.trim().isEmpty) return;

    final userMessage = Message(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      role: MessageRole.user,
      content: content,
      timestamp: DateTime.now().millisecondsSinceEpoch,
      status: MessageStatus.sent,
    );

    state = state.copyWith(
      messages: [...state.messages, userMessage],
      isLoading: true,
    );

    try {
      final response = await _api.sendMessage(
        messages: [...state.messages, userMessage],
        onChunk: (chunk) {
          _updateStreamingMessage(chunk);
        },
      );

      final assistantMessage = Message(
        id: response.id,
        role: MessageRole.assistant,
        content: response.content,
        agentName: response.agentName,
        agentRole: response.agentRole,
        timestamp: DateTime.now().millisecondsSinceEpoch,
        status: MessageStatus.sent,
      );

      state = state.copyWith(
        messages: [...state.messages, assistantMessage],
        isLoading: false,
        isTyping: false,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        isTyping: false,
        error: e.toString(),
      );
    }
  }

  void _updateStreamingMessage(String chunk) {
    final messages = List<Message>.from(state.messages);
    if (messages.isNotEmpty && messages.last.role == MessageRole.assistant) {
      final lastMessage = messages.removeLast();
      messages.add(Message(
        ...lastMessage,
        content: lastMessage.content + chunk,
        isStreaming: true,
      ));
      state = state.copyWith(messages: messages, isTyping: true);
    }
  }

  Future<void> retry(String messageId) async {
    final index = state.messages.indexWhere((m) => m.id == messageId);
    if (index == -1) return;

    final message = state.messages[index];
    state = state.copyWith(
      messages: state.messages.sublist(0, index),
    );
    await sendMessage(message.content);
  }

  void deleteMessage(String messageId) {
    state = state.copyWith(
      messages: state.messages.where((m) => m.id != messageId).toList(),
    );
  }

  void clearChat() {
    state = const ChatState();
  }
}

final chatBlocProvider = StateNotifierProvider<ChatBloc, ChatState>((ref) {
  final api = ref.watch(apiServiceProvider);
  return ChatBloc(api);
});
```
