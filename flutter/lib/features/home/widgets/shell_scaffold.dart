import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class ShellScaffold extends StatelessWidget {
  final Widget child;
  
  const ShellScaffold({super.key, required this.child});
  
  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final isWide = constraints.maxWidth >= 800;
        
        if (isWide) {
          return _WideLayout(child: child);
        } else {
          return _NarrowLayout(child: child);
        }
      },
    );
  }
}

class _WideLayout extends StatelessWidget {
  final Widget child;
  
  const _WideLayout({required this.child});
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Row(
        children: [
          const _NavigationRail(),
          const VerticalDivider(width: 1),
          Expanded(child: child),
        ],
      ),
    );
  }
}

class _NarrowLayout extends StatelessWidget {
  final Widget child;
  
  const _NarrowLayout({required this.child});
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: child,
      bottomNavigationBar: const _BottomNav(),
    );
  }
}

class _NavigationRail extends StatelessWidget {
  const _NavigationRail();
  
  @override
  Widget build(BuildContext context) {
    final currentPath = GoRouterState.of(context).uri.path;
    final selectedIndex = _getSelectedIndex(currentPath);
    
    return NavigationRail(
      extended: MediaQuery.of(context).size.width >= 1200,
      selectedIndex: selectedIndex,
      onDestinationSelected: (index) => _onDestinationSelected(context, index),
      leading: Padding(
        padding: const EdgeInsets.symmetric(vertical: 16),
        child: Column(
          children: [
            Container(
              width: 48,
              height: 48,
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.primary,
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Icon(
                Icons.smart_toy_outlined,
                color: Colors.white,
                size: 28,
              ),
            ),
            const SizedBox(height: 8),
            if (MediaQuery.of(context).size.width >= 1200)
              Text(
                'Tortoise',
                style: Theme.of(context).textTheme.titleSmall?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
          ],
        ),
      ),
      destinations: const [
        NavigationRailDestination(
          icon: Icon(Icons.home_outlined),
          selectedIcon: Icon(Icons.home),
          label: Text('首页'),
        ),
        NavigationRailDestination(
          icon: Icon(Icons.chat_outlined),
          selectedIcon: Icon(Icons.chat),
          label: Text('对话'),
        ),
        NavigationRailDestination(
          icon: Icon(Icons.cable_outlined),
          selectedIcon: Icon(Icons.cable),
          label: Text('渠道'),
        ),
        NavigationRailDestination(
          icon: Icon(Icons.psychology_outlined),
          selectedIcon: Icon(Icons.psychology),
          label: Text('记忆'),
        ),
        NavigationRailDestination(
          icon: Icon(Icons.extension_outlined),
          selectedIcon: Icon(Icons.extension),
          label: Text('插件'),
        ),
        NavigationRailDestination(
          icon: Icon(Icons.store_outlined),
          selectedIcon: Icon(Icons.store),
          label: Text('市场'),
        ),
        NavigationRailDestination(
          icon: Icon(Icons.mic_outlined),
          selectedIcon: Icon(Icons.mic),
          label: Text('语音'),
        ),
        NavigationRailDestination(
          icon: Icon(Icons.smart_toy_outlined),
          selectedIcon: Icon(Icons.smart_toy),
          label: Text('代理'),
        ),
        NavigationRailDestination(
          icon: Icon(Icons.settings_outlined),
          selectedIcon: Icon(Icons.settings),
          label: Text('设置'),
        ),
      ],
    );
  }
  
  int _getSelectedIndex(String path) {
    if (path == '/') return 0;
    if (path.startsWith('/chat')) return 1;
    if (path.startsWith('/channels')) return 2;
    if (path.startsWith('/memory')) return 3;
    if (path.startsWith('/plugins')) return 4;
    if (path.startsWith('/marketplace')) return 5;
    if (path.startsWith('/voice')) return 6;
    if (path.startsWith('/agents')) return 7;
    if (path.startsWith('/settings')) return 8;
    return 0;
  }
  
  void _onDestinationSelected(BuildContext context, int index) {
    switch (index) {
      case 0:
        context.go('/');
        break;
      case 1:
        context.go('/chat');
        break;
      case 2:
        context.go('/channels');
        break;
      case 3:
        context.go('/memory');
        break;
      case 4:
        context.go('/plugins');
        break;
      case 5:
        context.go('/marketplace');
        break;
      case 6:
        context.go('/voice');
        break;
      case 7:
        context.go('/agents');
        break;
      case 8:
        context.go('/settings');
        break;
    }
  }
}

class _BottomNav extends StatelessWidget {
  const _BottomNav();
  
  @override
  Widget build(BuildContext context) {
    final currentPath = GoRouterState.of(context).uri.path;
    final selectedIndex = _getSelectedIndex(currentPath);
    
    return NavigationBar(
      selectedIndex: selectedIndex,
      onDestinationSelected: (index) => _onDestinationSelected(context, index),
      destinations: const [
        NavigationDestination(
          icon: Icon(Icons.home_outlined),
          selectedIcon: Icon(Icons.home),
          label: '首页',
        ),
        NavigationDestination(
          icon: Icon(Icons.chat_outlined),
          selectedIcon: Icon(Icons.chat),
          label: '对话',
        ),
        NavigationDestination(
          icon: Icon(Icons.cable_outlined),
          selectedIcon: Icon(Icons.cable),
          label: '渠道',
        ),
        NavigationDestination(
          icon: Icon(Icons.psychology_outlined),
          selectedIcon: Icon(Icons.psychology),
          label: '记忆',
        ),
        NavigationDestination(
          icon: Icon(Icons.extension_outlined),
          selectedIcon: Icon(Icons.extension),
          label: '插件',
        ),
        NavigationDestination(
          icon: Icon(Icons.store_outlined),
          selectedIcon: Icon(Icons.store),
          label: '市场',
        ),
      ],
    );
  }
  
  int _getSelectedIndex(String path) {
    if (path == '/') return 0;
    if (path.startsWith('/chat')) return 1;
    if (path.startsWith('/channels')) return 2;
    if (path.startsWith('/memory')) return 3;
    if (path.startsWith('/plugins')) return 4;
    if (path.startsWith('/marketplace')) return 5;
    if (path.startsWith('/voice')) return 6;
    if (path.startsWith('/agents')) return 7;
    if (path.startsWith('/settings')) return 8;
    return 0;
  }
  
  void _onDestinationSelected(BuildContext context, int index) {
    switch (index) {
      case 0:
        context.go('/');
        break;
      case 1:
        context.go('/chat');
        break;
      case 2:
        context.go('/channels');
        break;
      case 3:
        context.go('/memory');
        break;
      case 4:
        context.go('/plugins');
        break;
      case 5:
        context.go('/marketplace');
        break;
    }
  }
}
