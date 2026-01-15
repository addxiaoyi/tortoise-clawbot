import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/theme/theme.dart';
import '../../core/di/providers.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  @override
  Widget build(BuildContext context) {
    final sessions = ref.watch(sessionsProvider);
    final connectionState = ref.watch(connectionStateProvider);
    
    return Scaffold(
      body: Row(
        children: [
          // Sidebar
          _buildSidebar(context, connectionState),
          // Main content
          Expanded(
            child: _buildMainContent(context, sessions),
          ),
        ],
      ),
    );
  }
  
  Widget _buildSidebar(BuildContext context, ConnectionState state) {
    final theme = Theme.of(context);
    
    return Container(
      width: 280,
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        border: Border(
          right: BorderSide(
            color: theme.dividerColor.withOpacity(0.1),
          ),
        ),
      ),
      child: Column(
        children: [
          // Header
          Padding(
            padding: const EdgeInsets.all(20),
            child: Row(
              children: [
                Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    color: TortoiseTheme.primary.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: const Icon(
                    Icons.smart_toy_outlined,
                    color: TortoiseTheme.primary,
                  ),
                ),
                const SizedBox(width: 12),
                Text(
                  'Tortoise',
                  style: theme.textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ),
          
          // New chat button
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: ElevatedButton.icon(
              onPressed: _createNewSession,
              icon: const Icon(Icons.add),
              label: const Text('New Chat'),
              style: ElevatedButton.styleFrom(
                backgroundColor: TortoiseTheme.primary,
                foregroundColor: Colors.white,
                minimumSize: const Size(double.infinity, 48),
              ),
            ),
          ),
          
          const SizedBox(height: 20),
          
          // Session list
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.symmetric(horizontal: 12),
              itemCount: sessions.length,
              itemBuilder: (context, index) {
                final session = sessions[index];
                return _buildSessionTile(context, session);
              },
            ),
          ),
          
          // Bottom actions
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                IconButton(
                  onPressed: () => Navigator.pushNamed(context, '/settings'),
                  icon: const Icon(Icons.settings_outlined),
                ),
                const Spacer(),
                _buildConnectionIndicator(state),
              ],
            ),
          ),
        ],
      ),
    );
  }
  
  Widget _buildSessionTile(BuildContext context, Session session) {
    final theme = Theme.of(context);
    final activeSession = ref.watch(activeSessionProvider);
    final isActive = activeSession == session.id;
    
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: ListTile(
        leading: Icon(
          Icons.chat_bubble_outline,
          color: isActive ? TortoiseTheme.primary : null,
        ),
        title: Text(
          session.name,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
        subtitle: Text(
          '${session.messageCount} messages',
          style: theme.textTheme.bodySmall,
        ),
        selected: isActive,
        selectedTileColor: TortoiseTheme.primary.withOpacity(0.1),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(10),
        ),
        onTap: () {
          ref.read(activeSessionProvider.notifier).state = session.id;
          Navigator.pushNamed(context, '/chat');
        },
      ),
    );
  }
  
  Widget _buildConnectionIndicator(ConnectionState state) {
    Color color;
    String text;
    IconData icon;
    
    switch (state) {
      case ConnectionState.connected:
        color = Colors.green;
        text = 'Connected';
        icon = Icons.check_circle;
        break;
      case ConnectionState.connecting:
        color = Colors.orange;
        text = 'Connecting';
        icon = Icons.sync;
        break;
      case ConnectionState.disconnected:
        color = Colors.grey;
        text = 'Offline';
        icon = Icons.cloud_off;
        break;
      case ConnectionState.error:
        color = Colors.red;
        text = 'Error';
        icon = Icons.error;
        break;
    }
    
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 16, color: color),
        const SizedBox(width: 6),
        Text(
          text,
          style: TextStyle(color: color, fontSize: 12),
        ),
      ],
    );
  }
  
  Widget _buildMainContent(BuildContext context, List<Session> sessions) {
    final theme = Theme.of(context);
    
    if (sessions.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.smart_toy_outlined,
              size: 80,
              color: theme.colorScheme.primary.withOpacity(0.3),
            ),
            const SizedBox(height: 24),
            Text(
              'Welcome to Tortoise',
              style: theme.textTheme.headlineMedium?.copyWith(
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 12),
            Text(
              'Start a new conversation or select an existing session',
              style: theme.textTheme.bodyLarge?.copyWith(
                color: theme.textTheme.bodySmall?.color,
              ),
            ),
            const SizedBox(height: 32),
            ElevatedButton.icon(
              onPressed: _createNewSession,
              icon: const Icon(Icons.add),
              label: const Text('New Chat'),
            ),
          ],
        ),
      );
    }
    
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.arrow_back, size: 32),
          const SizedBox(height: 16),
          Text(
            'Select a session',
            style: theme.textTheme.titleMedium,
          ),
        ],
      ),
    );
  }
  
  void _createNewSession() {
    final session = Session(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      name: 'New Chat ${DateTime.now().millisecondsSinceEpoch}',
      createdAt: DateTime.now(),
    );
    
    ref.read(sessionsProvider.notifier).addSession(session);
    ref.read(activeSessionProvider.notifier).state = session.id;
    Navigator.pushNamed(context, '/chat');
  }
}
