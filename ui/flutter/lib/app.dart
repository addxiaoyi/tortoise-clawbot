import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'core/theme/theme.dart';
import 'features/home/presentation/home_screen.dart';
import 'features/settings/presentation/settings_screen.dart';
import 'features/chat/presentation/chat_screen.dart';

class TortoiseApp extends ConsumerWidget {
  const TortoiseApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final themeMode = ref.watch(themeModeProvider);
    
    return MaterialApp(
      title: 'Tortoise',
      debugShowCheckedModeBanner: false,
      theme: TortoiseTheme.lightTheme,
      darkTheme: TortoiseTheme.darkTheme,
      themeMode: themeMode,
      initialRoute: '/',
      routes: {
        '/': (context) => const HomeScreen(),
        '/chat': (context) => const ChatScreen(),
        '/settings': (context) => const SettingsScreen(),
      },
    );
  }
}
