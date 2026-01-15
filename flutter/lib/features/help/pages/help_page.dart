import 'package:flutter/material.dart';

/// 帮助页面
class HelpPage extends StatelessWidget {
  const HelpPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('帮助'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // 快速入门
          _buildSection(
            context,
            title: '🚀 快速入门',
            icon: Icons.play_circle_outline,
            children: [
              _buildFAQItem(
                context,
                question: '如何开始使用？',
                answer: '1. 在设置页面配置您的 AI API Key\n2. 选择您偏好的 AI 提供商\n3. 开始创建新会话并开始对话',
              ),
              _buildFAQItem(
                context,
                question: '支持哪些 AI 模型？',
                answer: '目前支持 OpenAI GPT-4、Claude 3、Gemini 等主流模型。您可以在设置中切换不同的模型。',
              ),
            ],
          ),
          
          const SizedBox(height: 20),
          
          // 消息渠道
          _buildSection(
            context,
            title: '💬 消息渠道',
            icon: Icons.cable_outlined,
            children: [
              _buildFAQItem(
                context,
                question: '如何连接 Telegram？',
                answer: '1. 在 Telegram 中创建 Bot（联系 @BotFather）\n2. 复制 Bot Token\n3. 在渠道设置中粘贴 Token\n4. 点击连接按钮',
              ),
              _buildFAQItem(
                context,
                question: '如何连接 Discord？',
                answer: '1. 在 Discord Developer Portal 创建应用\n2. 添加 Bot 并获取 Token\n3. 启用 Message Content Intent\n4. 在渠道设置中配置',
              ),
            ],
          ),
          
          const SizedBox(height: 20),
          
          // 记忆系统
          _buildSection(
            context,
            title: '🧠 记忆系统',
            icon: Icons.psychology_outlined,
            children: [
              _buildFAQItem(
                context,
                question: '记忆是什么？',
                answer: '记忆是 Tortoise 的知识库功能，可以保存重要的信息、会话摘要和用户偏好，让 AI 更好地了解您的需求。',
              ),
              _buildFAQItem(
                context,
                question: '如何添加记忆？',
                answer: '在记忆页面，点击右下角的 + 按钮，您可以添加标题和内容。记忆会自动分类和索引，方便后续查找。',
              ),
            ],
          ),
          
          const SizedBox(height: 20),
          
          // 插件系统
          _buildSection(
            context,
            title: '🔌 插件系统',
            icon: Icons.extension_outlined,
            children: [
              _buildFAQItem(
                context,
                question: '什么是插件？',
                answer: '插件是扩展 Tortoise 功能的组件，可以添加新的 AI 能力、工具和集成。',
              ),
              _buildFAQItem(
                context,
                question: '如何安装插件？',
                answer: '在插件页面，浏览可用插件列表，点击安装即可。插件安装后会自动启用。',
              ),
            ],
          ),
          
          const SizedBox(height: 20),
          
          // 常见问题
          _buildSection(
            context,
            title: '❓ 常见问题',
            icon: Icons.help_outline,
            children: [
              _buildFAQItem(
                context,
                question: 'API 调用失败怎么办？',
                answer: '1. 检查 API Key 是否正确\n2. 确认 API Key 有足够的配额\n3. 检查网络连接\n4. 查看错误日志获取详细信息',
              ),
              _buildFAQItem(
                context,
                question: '数据存储在哪里？',
                answer: '所有数据默认存储在本地设备上。您可以在设置中管理数据存储和导出选项。',
              ),
              _buildFAQItem(
                context,
                question: '如何导出会话记录？',
                answer: '在会话详情页面，点击右上角菜单，选择"导出"选项即可导出为 JSON 或 Markdown 格式。',
              ),
            ],
          ),
          
          const SizedBox(height: 20),
          
          // 联系方式
          Card(
            child: Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(
                        Icons.support_agent,
                        color: Theme.of(context).colorScheme.primary,
                      ),
                      const SizedBox(width: 12),
                      Text(
                        '获取帮助',
                        style: Theme.of(context).textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  const Text('如果您遇到问题或有建议，可以通过以下方式联系我们：'),
                  const SizedBox(height: 12),
                  Wrap(
                    spacing: 12,
                    runSpacing: 12,
                    children: [
                      OutlinedButton.icon(
                        onPressed: () {},
                        icon: const Icon(Icons.bug_report),
                        label: const Text('提交问题'),
                      ),
                      OutlinedButton.icon(
                        onPressed: () {},
                        icon: const Icon(Icons.lightbulb),
                        label: const Text('功能建议'),
                      ),
                      OutlinedButton.icon(
                        onPressed: () {},
                        icon: const Icon(Icons.email),
                        label: const Text('邮件联系'),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          
          const SizedBox(height: 40),
        ],
      ),
    );
  }

  Widget _buildSection(
    BuildContext context, {
    required String title,
    required IconData icon,
    required List<Widget> children,
  }) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, color: Theme.of(context).colorScheme.primary),
                const SizedBox(width: 12),
                Text(
                  title,
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            ...children,
          ],
        ),
      ),
    );
  }

  Widget _buildFAQItem(
    BuildContext context, {
    required String question,
    required String answer,
  }) {
    return ExpansionTile(
      title: Text(
        question,
        style: const TextStyle(fontWeight: FontWeight.w500),
      ),
      childrenPadding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
      children: [
        Text(
          answer,
          style: TextStyle(
            color: Theme.of(context).colorScheme.onSurface.withOpacity(0.7),
            height: 1.5,
          ),
        ),
      ],
    );
  }
}
