import 'package:flutter/material.dart';

class PluginsScreen extends StatelessWidget {
  const PluginsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Plugins'),
        actions: [
          IconButton(
            icon: const Icon(Icons.store),
            onPressed: () {},
          ),
        ],
      ),
      body: ListView(
        children: [
          const _SectionHeader(title: 'Installed'),
          _PluginTile(
            name: 'Web Search',
            description: 'Search the web for information',
            version: '1.0.0',
            enabled: true,
          ),
          _PluginTile(
            name: 'Calculator',
            description: 'Perform mathematical calculations',
            version: '1.0.0',
            enabled: true,
          ),
          _PluginTile(
            name: 'File Manager',
            description: 'Read and write files',
            version: '1.0.0',
            enabled: false,
          ),
          const Divider(),
          const _SectionHeader(title: 'Available'),
          _PluginTile(
            name: 'GitHub Integration',
            description: 'Interact with GitHub repositories',
            version: '1.0.0',
            enabled: false,
            installable: true,
          ),
          _PluginTile(
            name: 'Calendar',
            description: 'Manage your calendar events',
            version: '1.0.0',
            enabled: false,
            installable: true,
          ),
        ],
      ),
    );
  }
}

class _PluginTile extends StatelessWidget {
  final String name;
  final String description;
  final String version;
  final bool enabled;
  final bool installable;

  const _PluginTile({
    required this.name,
    required this.description,
    required this.version,
    required this.enabled,
    this.installable = false,
  });

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.primaryContainer,
          borderRadius: BorderRadius.circular(8),
        ),
        child: Icon(
          Icons.extension,
          color: Theme.of(context).colorScheme.primary,
        ),
      ),
      title: Text(name),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(description),
          Text(
            'v$version',
            style: Theme.of(context).textTheme.bodySmall,
          ),
        ],
      ),
      trailing: installable
          ? FilledButton.tonal(
              onPressed: () {},
              child: const Text('Install'),
            )
          : Switch(
              value: enabled,
              onChanged: (value) {},
            ),
      isThreeLine: true,
      onTap: () {},
    );
  }
}

class _SectionHeader extends StatelessWidget {
  final String title;

  const _SectionHeader({required this.title});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
      child: Text(
        title,
        style: Theme.of(context).textTheme.titleSmall?.copyWith(
          color: Theme.of(context).colorScheme.primary,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }
}
