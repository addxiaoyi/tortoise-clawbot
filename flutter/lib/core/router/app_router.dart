import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../features/chat/pages/chat_page.dart';
import '../../features/chat/pages/sessions_page.dart';
import '../../features/settings/pages/settings_page.dart';
import '../../features/channels/pages/channels_page.dart';
import '../../features/memory/pages/memory_page.dart';
import '../../features/plugins/pages/plugins_page.dart';
import '../../features/home/pages/home_page.dart';
import '../../features/home/widgets/shell_scaffold.dart';
import '../../features/voice/pages/voice_wake_page.dart';
import '../../features/marketplace/pages/marketplace_page.dart';
import '../../features/agents/pages/agents_page.dart';

final routerProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/',
    routes: [
      ShellRoute(
        builder: (context, state, child) {
          return ShellScaffold(child: child);
        },
        routes: [
          GoRoute(
            path: '/',
            name: 'home',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: HomePage(),
            ),
          ),
          GoRoute(
            path: '/chat',
            name: 'chat',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: ChatPage(),
            ),
            routes: [
              GoRoute(
                path: 'sessions',
                name: 'sessions',
                pageBuilder: (context, state) => const NoTransitionPage(
                  child: SessionsPage(),
                ),
              ),
            ],
          ),
          GoRoute(
            path: '/channels',
            name: 'channels',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: ChannelsPage(),
            ),
          ),
          GoRoute(
            path: '/memory',
            name: 'memory',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: MemoryPage(),
            ),
          ),
          GoRoute(
            path: '/plugins',
            name: 'plugins',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: PluginsPage(),
            ),
          ),
          GoRoute(
            path: '/marketplace',
            name: 'marketplace',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: MarketplacePage(),
            ),
          ),
          GoRoute(
            path: '/voice',
            name: 'voice',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: VoiceWakePage(),
            ),
          ),
          GoRoute(
            path: '/agents',
            name: 'agents',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: AgentsPage(),
            ),
          ),
          GoRoute(
            path: '/settings',
            name: 'settings',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: SettingsPage(),
            ),
          ),
        ],
      ),
    ],
  );
});
