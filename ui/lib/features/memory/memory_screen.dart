// Memory Screen

import 'package:flutter/material.dart';

class MemoryScreen extends StatelessWidget {
  const MemoryScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Memory'),
      ),
      body: const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.memory, size: 64, color: Colors.grey),
            SizedBox(height: 16),
            Text('Memory Management'),
            SizedBox(height: 8),
            Text('View and manage agent memories'),
          ],
        ),
      ),
    );
  }
}
