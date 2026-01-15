import 'package:flutter/material.dart';

class StatsScreen extends StatelessWidget {
  const StatsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Statistics'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () {},
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Overview Cards
          Row(
            children: [
              Expanded(
                child: _StatCard(
                  title: 'Total Messages',
                  value: '1,234',
                  icon: Icons.chat,
                  color: Colors.blue,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _StatCard(
                  title: 'Total Tokens',
                  value: '56K',
                  icon: Icons.toll,
                  color: Colors.green,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              Expanded(
                child: _StatCard(
                  title: 'Tool Calls',
                  value: '89',
                  icon: Icons.build,
                  color: Colors.orange,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _StatCard(
                  title: 'Avg Response',
                  value: '1.2s',
                  icon: Icons.timer,
                  color: Colors.purple,
                ),
              ),
            ],
          ),
          
          const SizedBox(height: 24),
          
          // Usage Chart
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Usage Over Time',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                    ),
                  ),
                  const SizedBox(height: 16),
                  SizedBox(
                    height: 200,
                    child: _UsageChart(),
                  ),
                ],
              ),
            ),
          ),
          
          const SizedBox(height: 16),
          
          // Model Usage
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Model Usage',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                    ),
                  ),
                  const SizedBox(height: 16),
                  _ModelUsageRow(
                    model: 'GPT-4',
                    percentage: 65,
                    color: Colors.green,
                  ),
                  const SizedBox(height: 8),
                  _ModelUsageRow(
                    model: 'Claude',
                    percentage: 25,
                    color: Colors.blue,
                  ),
                  const SizedBox(height: 8),
                  _ModelUsageRow(
                    model: 'Gemini',
                    percentage: 10,
                    color: Colors.orange,
                  ),
                ],
              ),
            ),
          ),
          
          const SizedBox(height: 16),
          
          // Channel Distribution
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Channel Distribution',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                    ),
                  ),
                  const SizedBox(height: 16),
                  _ChannelRow(
                    channel: 'Discord',
                    count: 456,
                    icon: Icons.discord,
                  ),
                  const SizedBox(height: 8),
                  _ChannelRow(
                    channel: 'Telegram',
                    count: 234,
                    icon: Icons.telegram,
                  ),
                  const SizedBox(height: 8),
                  _ChannelRow(
                    channel: 'WhatsApp',
                    count: 123,
                    icon: Icons.chat,
                  ),
                  const SizedBox(height: 8),
                  _ChannelRow(
                    channel: 'Direct',
                    count: 89,
                    icon: Icons.email,
                  ),
                ],
              ),
            ),
          ),
          
          const SizedBox(height: 16),
          
          // Top Tools
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Top Tools',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                    ),
                  ),
                  const SizedBox(height: 16),
                  _ToolRow(
                    tool: 'Web Search',
                    count: 45,
                    icon: Icons.search,
                  ),
                  const SizedBox(height: 8),
                  _ToolRow(
                    tool: 'Calculator',
                    count: 23,
                    icon: Icons.calculate,
                  ),
                  const SizedBox(height: 8),
                  _ToolRow(
                    tool: 'File Manager',
                    count: 12,
                    icon: Icons.folder,
                  ),
                  const SizedBox(height: 8),
                  _ToolRow(
                    tool: 'GitHub',
                    count: 9,
                    icon: Icons.code,
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _StatCard extends StatelessWidget {
  final String title;
  final String value;
  final IconData icon;
  final Color color;

  const _StatCard({
    required this.title,
    required this.value,
    required this.icon,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, color: color, size: 24),
            const SizedBox(height: 8),
            Text(
              value,
              style: const TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.bold,
              ),
            ),
            Text(
              title,
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ],
        ),
      ),
    );
  }
}

class _UsageChart extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return CustomPaint(
      painter: _ChartPainter(),
      size: const Size(double.infinity, 200),
    );
  }
}

class _ChartPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = Theme.of(canvas).colorScheme.primary.withOpacity(0.3)
      ..style = PaintingStyle.fill;

    final linePaint = Paint()
      ..color = Theme.of(canvas).colorScheme.primary
      ..strokeWidth = 2
      ..style = PaintingStyle.stroke;

    final path = Path();
    final linePath = Path();

    path.moveTo(0, size.height);
    linePath.moveTo(0, size.height * 0.7);

    final points = [0.7, 0.5, 0.6, 0.4, 0.8, 0.6, 0.5, 0.7];
    final step = size.width / (points.length - 1);

    for (var i = 0; i < points.length; i++) {
      final x = i * step;
      final y = size.height * (1 - points[i]);
      path.lineTo(x, y);
      linePath.lineTo(x, y);
    }

    path.lineTo(size.width, size.height);
    path.close();

    canvas.drawPath(path, paint);
    canvas.drawPath(linePath, linePaint);

    // Draw dots
    final dotPaint = Paint()
      ..color = Theme.of(canvas).colorScheme.primary
      ..style = PaintingStyle.fill;

    for (var i = 0; i < points.length; i++) {
      final x = i * step;
      final y = size.height * (1 - points[i]);
      canvas.drawCircle(Offset(x, y), 4, dotPaint);
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _ModelUsageRow extends StatelessWidget {
  final String model;
  final int percentage;
  final Color color;

  const _ModelUsageRow({
    required this.model,
    required this.percentage,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        SizedBox(
          width: 80,
          child: Text(model),
        ),
        Expanded(
          child: ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: LinearProgressIndicator(
              value: percentage / 100,
              backgroundColor: Theme.of(context).colorScheme.surfaceContainerHighest,
              valueColor: AlwaysStoppedAnimation<Color>(color),
              minHeight: 8,
            ),
          ),
        ),
        const SizedBox(width: 8),
        Text('$percentage%'),
      ],
    );
  }
}

class _ChannelRow extends StatelessWidget {
  final String channel;
  final int count;
  final IconData icon;

  const _ChannelRow({
    required this.channel,
    required this.count,
    required this.icon,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Icon(icon, size: 20),
        const SizedBox(width: 8),
        Expanded(child: Text(channel)),
        Text(
          count.toString(),
          style: Theme.of(context).textTheme.bodySmall,
        ),
      ],
    );
  }
}

class _ToolRow extends StatelessWidget {
  final String tool;
  final int count;
  final IconData icon;

  const _ToolRow({
    required this.tool,
    required this.count,
    required this.icon,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Icon(icon, size: 20),
        const SizedBox(width: 8),
        Expanded(child: Text(tool)),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
          decoration: BoxDecoration(
            color: Theme.of(context).colorScheme.primaryContainer,
            borderRadius: BorderRadius.circular(12),
          ),
          child: Text(
            count.toString(),
            style: Theme.of(context).textTheme.bodySmall,
          ),
        ),
      ],
    );
  }
}
