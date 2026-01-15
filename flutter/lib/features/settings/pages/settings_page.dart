import 'package:flutter/material.dart';

class SettingsPage extends StatelessWidget {
  const SettingsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('设置'),
      ),
      body: ListView(
        children: [
          const ListTile(
            title: Text('AI 设置'),
            subtitle: Text('配置 AI 提供商'),
          ),
          const ListTile(
            title: Text('渠道设置'),
            subtitle: Text('配置消息渠道'),
          ),
          const ListTile(
            title: Text('主题'),
            subtitle: Text('深色/浅色模式'),
          ),
        ],
      ),
    );
  }
}
