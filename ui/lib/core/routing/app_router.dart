// Tortoise App Router

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:tortoise/features/chat/chat_screen.dart';
import 'package:tortoise/features/memory/memory_screen.dart';
import 'package:tortoise/features/plugins/plugins_screen.dart';
import 'package:tortoise/features/settings/settings_screen.dart';
import 'package:tortoise/shared/widgets/shell_scaffold.dart';

/// App Router Provider
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
