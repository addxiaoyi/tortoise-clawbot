import 'package:flutter/material.dart';

/// 状态指示器组件
class StatusIndicator extends StatelessWidget {
  final String label;
  final bool isOnline;
  final String? subtitle;

  const StatusIndicator({
    super.key,
    required this.label,
    required this.isOnline,
    this.subtitle,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: isOnline
            ? Colors.green.withOpacity(0.1)
            : Colors.grey.withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: isOnline ? Colors.green.withOpacity(0.3) : Colors.grey.withOpacity(0.3),
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 10,
            height: 10,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: isOnline ? Colors.green : Colors.grey,
              boxShadow: isOnline
                  ? [
                      BoxShadow(
                        color: Colors.green.withOpacity(0.5),
                        blurRadius: 6,
                        spreadRadius: 1,
                      ),
                    ]
                  : null,
            ),
          ),
          const SizedBox(width: 12),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                label,
                style: TextStyle(
                  fontWeight: FontWeight.w600,
                  color: isOnline ? Colors.green.shade700 : Colors.grey.shade700,
                ),
              ),
              if (subtitle != null)
                Text(
                  subtitle!,
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.grey.shade600,
                  ),
                ),
            ],
          ),
        ],
      ),
    );
  }
}

/// 连接状态卡片
class ConnectionStatusCard extends StatelessWidget {
  final String serviceName;
  final bool isConnected;
  final String? lastConnected;
  final VoidCallback? onReconnect;

  const ConnectionStatusCard({
    super.key,
    required this.serviceName,
    required this.isConnected,
    this.lastConnected,
    this.onReconnect,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(
        leading: Icon(
          isConnected ? Icons.check_circle : Icons.error_outline,
          color: isConnected ? Colors.green : Colors.red,
        ),
        title: Text(serviceName),
        subtitle: Text(
          isConnected
              ? '已连接${lastConnected != null ? " - $lastConnected" : ""}'
              : '未连接',
        ),
        trailing: isConnected
            ? null
            : TextButton(
                onPressed: onReconnect,
                child: const Text('重连'),
              ),
      ),
    );
  }
}
