/// 应用统计服务
class StatsService {
  static StatsService? _instance;
  static StatsService get instance => _instance ??= StatsService._();
  StatsService._();

  int _totalMessages = 0;
  int _totalSessions = 0;
  int _totalTokens = 0;
  Duration _totalConversationTime = Duration.zero;
  DateTime? _lastSessionAt;

  /// 获取统计摘要
  StatsSummary get summary => StatsSummary(
    totalMessages: _totalMessages,
    totalSessions: _totalSessions,
    totalTokens: _totalTokens,
    totalConversationTime: _totalConversationTime,
    lastSessionAt: _lastSessionAt,
  );

  /// 记录消息
  void recordMessage({required bool isFromUser, int? tokens}) {
    _totalMessages++;
    if (tokens != null) {
      _totalTokens += tokens;
    }
  }

  /// 记录会话开始
  void recordSessionStart() {
    _totalSessions++;
    _lastSessionAt = DateTime.now();
  }

  /// 记录会话时长
  void recordConversationTime(Duration duration) {
    _totalConversationTime += duration;
  }

  /// 重置统计
  void reset() {
    _totalMessages = 0;
    _totalSessions = 0;
    _totalTokens = 0;
    _totalConversationTime = Duration.zero;
    _lastSessionAt = null;
  }
}

/// 统计摘要
class StatsSummary {
  final int totalMessages;
  final int totalSessions;
  final int totalTokens;
  final Duration totalConversationTime;
  final DateTime? lastSessionAt;

  StatsSummary({
    required this.totalMessages,
    required this.totalSessions,
    required this.totalTokens,
    required this.totalConversationTime,
    this.lastSessionAt,
  });

  /// 格式化会话时长
  String get formattedConversationTime {
    final hours = totalConversationTime.inHours;
    final minutes = totalConversationTime.inMinutes % 60;
    if (hours > 0) {
      return '${hours}h ${minutes}m';
    }
    return '${minutes}m';
  }

  /// 估算成本 (假设 $0.002/1K tokens)
  double get estimatedCost => totalTokens * 0.002 / 1000;
}
