// Plugins Screen

import 'package:flutter/material.dart';

class PluginsScreen extends StatelessWidget {
  const PluginsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Plugins'),
      ),
      body: const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.extension, size: 64, color: Colors.grey),
            SizedBox(height: 16),
            Text('Plugin Management'),
            SizedBox(height: 8),
            Text('Manage Tortoise plugins'),
          ],
        ),
      ),
    );
  }
}
