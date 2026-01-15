import 'package:flutter/material.dart';

/// 关于页面
class AboutPage extends StatelessWidget {
  const AboutPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('关于'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            const SizedBox(height: 20),
            
            // Logo
            Container(
              width: 120,
              height: 120,
              decoration: BoxDecoration(
                gradient: const LinearGradient(
                  colors: [Color(0xFF0D9488), Color(0xFF0F766E)],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
                borderRadius: BorderRadius.circular(30),
                boxShadow: [
                  BoxShadow(
                    color: const Color(0xFF0D9488).withOpacity(0.3),
                    blurRadius: 20,
                    offset: const Offset(0, 10),
                  ),
                ],
              ),
              child: const Icon(
                Icons.smart_toy,
                size: 60,
                color: Colors.white,
              ),
            ),
            
            const SizedBox(height: 24),
            
            // 应用名称
            Text(
              'Tortoise',
              style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            
            const SizedBox(height: 8),
            
            // 版本
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.primaryContainer,
                borderRadius: BorderRadius.circular(20),
              ),
              child: Text(
                'Version 0.1.0',
                style: TextStyle(
                  color: Theme.of(context).colorScheme.onPrimaryContainer,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
            
            const SizedBox(height: 24),
            
            // 描述
            Text(
              '下一代 AI 代理框架',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    color: Theme.of(context)
                        .colorScheme
                        .onSurface
                        .withOpacity(0.7),
                  ),
            ),
            
            const SizedBox(height: 40),
            
            // 功能特性
            _buildFeatureCard(
              context,
              icon: Icons.psychology,
              title: '多模型支持',
              description: '支持 OpenAI、Claude、Gemini 等多种 AI 模型',
            ),
            _buildFeatureCard(
              context,
              icon: Icons.cable,
              title: '多渠道接入',
              description: '支持 Telegram、Discord、Slack 等多种消息渠道',
            ),
            _buildFeatureCard(
              context,
              icon: Icons.extension,
              title: '插件系统',
              description: '丰富的插件生态，无限扩展功能',
            ),
            _buildFeatureCard(
              context,
              icon: Icons.storage,
              title: '本地优先',
              description: '数据本地存储，保护隐私安全',
            ),
            
            const SizedBox(height: 40),
            
            // 技术栈
            Card(
              child: Padding(
                padding: const EdgeInsets.all(20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      '🛠 技术栈',
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                    ),
                    const SizedBox(height: 16),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        _buildTechChip('Flutter'),
                        _buildTechChip('Dart'),
                        _buildTechChip('Riverpod'),
                        _buildTechChip('Go Router'),
                        _buildTechChip('Hive'),
                        _buildTechChip('Dio'),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            
            const SizedBox(height: 20),
            
            // 链接
            Card(
              child: Column(
                children: [
                  ListTile(
                    leading: const Icon(Icons.code),
                    title: const Text('源代码'),
                    trailing: const Icon(Icons.open_in_new, size: 18),
                    onTap: () {},
                  ),
                  const Divider(height: 1),
                  ListTile(
                    leading: const Icon(Icons.description),
                    title: const Text('使用文档'),
                    trailing: const Icon(Icons.open_in_new, size: 18),
                    onTap: () {},
                  ),
                  const Divider(height: 1),
                  ListTile(
                    leading: const Icon(Icons.bug_report),
                    title: const Text('问题反馈'),
                    trailing: const Icon(Icons.open_in_new, size: 18),
                    onTap: () {},
                  ),
                ],
              ),
            ),
            
            const SizedBox(height: 20),
            
            // 许可证
            Card(
              child: Padding(
                padding: const EdgeInsets.all(20),
                child: Column(
                  children: [
                    const Icon(Icons.copyright, size: 40),
                    const SizedBox(height: 12),
                    Text(
                      'Apache 2.0 License',
                      style: Theme.of(context).textTheme.titleSmall,
                    ),
                    const SizedBox(height: 8),
                    Text(
                      '© 2024 Tortoise Team\nAll rights reserved.',
                      textAlign: TextAlign.center,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: Theme.of(context)
                                .colorScheme
                                .onSurface
                                .withOpacity(0.6),
                          ),
                    ),
                  ],
                ),
              ),
            ),
            
            const SizedBox(height: 40),
          ],
        ),
      ),
    );
  }

  Widget _buildFeatureCard(
    BuildContext context, {
    required IconData icon,
    required String title,
    required String description,
  }) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: ListTile(
        leading: Container(
          padding: const EdgeInsets.all(10),
          decoration: BoxDecoration(
            color: Theme.of(context).colorScheme.primaryContainer,
            borderRadius: BorderRadius.circular(12),
          ),
          child: Icon(
            icon,
            color: Theme.of(context).colorScheme.primary,
          ),
        ),
        title: Text(title),
        subtitle: Text(description),
      ),
    );
  }

  Widget _buildTechChip(String label) {
    return Chip(
      label: Text(label),
      padding: const EdgeInsets.symmetric(horizontal: 8),
    );
  }
}
