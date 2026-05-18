import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class SettingsPage extends ConsumerWidget {
  const SettingsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text('Settings', style: TextStyle(fontWeight: FontWeight.w600)),
        elevation: 0,
      ),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: const [
          _Section(title: 'Account', children: [
            _Tile(icon: Icons.person, title: 'Profile', subtitle: 'Manage your account'),
            _Tile(icon: Icons.key, title: 'API Keys', subtitle: 'Configure AI providers'),
          ]),
          SizedBox(height: 24),
          _Section(title: 'Preferences', children: [
            _Tile(icon: Icons.palette, title: 'Theme', subtitle: 'Light / Dark / System'),
            _Tile(icon: Icons.language, title: 'Language', subtitle: 'English'),
            _Tile(icon: Icons.notifications, title: 'Notifications', subtitle: 'Push & Email'),
          ]),
          SizedBox(height: 24),
          _Section(title: 'Privacy', children: [
            _Tile(icon: Icons.security, title: 'Security', subtitle: 'Password & 2FA'),
            _Tile(icon: Icons.privacy_tip, title: 'Privacy', subtitle: 'Data & Analytics'),
          ]),
          SizedBox(height: 24),
          _Section(title: 'About', children: [
            _Tile(icon: Icons.info, title: 'About', subtitle: 'Version 0.1.0'),
            _Tile(icon: Icons.help, title: 'Help', subtitle: 'Documentation & Support'),
          ]),
        ],
      ),
    );
  }
}

class _Section extends StatelessWidget {
  final String title;
  final List<Widget> children;
  const _Section({required this.title, required this.children});

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(title, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Color(0xFF667EEA))),
        const SizedBox(height: 12),
        Container(
          decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(16)),
          child: Column(children: children),
        ),
      ],
    );
  }
}

class _Tile extends StatelessWidget {
  final IconData icon;
  final String title;
  final String subtitle;
  const _Tile({required this.icon, required this.title, required this.subtitle});

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(color: const Color(0xFF667EEA).withValues(alpha: 0.1), borderRadius: BorderRadius.circular(8)),
        child: Icon(icon, color: const Color(0xFF667EEA), size: 20),
      ),
      title: Text(title, style: const TextStyle(fontWeight: FontWeight.w500)),
      subtitle: Text(subtitle, style: const TextStyle(fontSize: 12, color: Color(0xFF64748B))),
      trailing: const Icon(Icons.chevron_right, color: Color(0xFF94A3B8)),
      onTap: () {},
    );
  }
}
