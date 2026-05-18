import 'dart:async';
import 'dart:math';

/// Subconscious Self-Learning Loop - 潜意识自学习循环
/// 模拟人类潜意识的学习和适应过程
class SubconsciousLoop {
  /// 学习循环状态
  bool _isRunning = false;
  Timer? _learningTimer;
  final _random = Random();

  /// 用户行为追踪
  final List<UserBehavior> _behaviorLog = [];

  /// 学习到的模式
  final List<LearnedPattern> _patterns = [];

  /// 偏好设置
  final Map<String, dynamic> _preferences = {};

  /// 回调函数
  Function(String event, Map<String, dynamic> data)? onInsight;
  Function(String pattern, double confidence)? onPatternLearned;

  /// 启动学习循环
  void start() {
    if (_isRunning) return;
    _isRunning = true;

    // 每 5 分钟执行一次潜意识处理
    _learningTimer = Timer.periodic(
      const Duration(minutes: 5),
      (_) => _processSubconscious(),
    );
  }

  /// 停止学习循环
  void stop() {
    _isRunning = false;
    _learningTimer?.cancel();
    _learningTimer = null;
  }

  /// 记录用户行为
  void recordBehavior({
    required String action,
    required String context,
    Map<String, dynamic> metadata = const {},
  }) {
    _behaviorLog.add(UserBehavior(
      action: action,
      context: context,
      timestamp: DateTime.now(),
      metadata: metadata,
    ));

    // 保持最近 1000 条记录
    if (_behaviorLog.length > 1000) {
      _behaviorLog.removeAt(0);
    }

    // 实时学习：检查是否形成新模式
    _checkForNewPatterns(action, context, metadata);
  }

  /// 检查是否形成新模式
  void _checkForNewPatterns(String action, String context, Map<String, dynamic> metadata) {
    // 简单模式检测：如果同一个 action+context 出现 3 次以上
    final recentCount = _behaviorLog.where((b) =>
        b.action == action && b.context == context).length;

    if (recentCount >= 3) {
      // 检查是否已存在该模式
      final existingPattern = _patterns.firstWhere(
        (p) => p.action == action && p.context == context,
        orElse: () => LearnedPattern(action: '', context: '', confidence: 0),
      );

      if (existingPattern.action.isEmpty) {
        // 新模式
        final pattern = LearnedPattern(
          action: action,
          context: context,
          confidence: min(0.9, 0.3 + recentCount * 0.1),
          sampleCount: recentCount,
          learnedAt: DateTime.now(),
        );
        _patterns.add(pattern);
        onPatternLearned?.call(action, pattern.confidence);
      } else {
        // 更新置信度
        existingPattern.sampleCount = recentCount;
        existingPattern.confidence = min(0.95, 0.3 + recentCount * 0.1);
      }
    }
  }

  /// 处理潜意识
  Future<void> _processSubconscious() async {
    if (!_isRunning) return;

    // 1. 提取行为趋势
    final trends = _analyzeTrends();

    // 2. 更新偏好
    _updatePreferences(trends);

    // 3. 生成洞察
    final insights = _generateInsights(trends);

    // 4. 触发回调
    for (final insight in insights) {
      onInsight?.call(insight.type, insight.data);
    }

    // 5. 清理旧数据
    _cleanupOldData();
  }

  /// 分析趋势
  Map<String, dynamic> _analyzeTrends() {
    final now = DateTime.now();
    final hourAgo = now.subtract(const Duration(hours: 1));
    final dayAgo = now.subtract(const Duration(days: 1));

    final recentBehaviors = _behaviorLog.where((b) => b.timestamp.isAfter(hourAgo)).toList();
    final todayBehaviors = _behaviorLog.where((b) => b.timestamp.isAfter(dayAgo)).toList();

    // 按动作分组
    final actionCounts = <String, int>{};
    for (final behavior in recentBehaviors) {
      actionCounts[behavior.action] = (actionCounts[behavior.action] ?? 0) + 1;
    }

    // 最常见的动作
    final topActions = actionCounts.entries.toList()
      ..sort((a, b) => b.value.compareTo(a.value));

    return {
      'total_actions_1h': recentBehaviors.length,
      'total_actions_24h': todayBehaviors.length,
      'top_actions': topActions.take(5).map((e) => {'action': e.key, 'count': e.value}).toList(),
      'timestamp': now.toIso8601String(),
    };
  }

  /// 更新偏好
  void _updatePreferences(Map<String, dynamic> trends) {
    final topActions = trends['top_actions'] as List;
    if (topActions.isNotEmpty) {
      _preferences['preferred_action'] = topActions.first['action'];
    }

    final total24h = trends['total_actions_24h'] as int;
    _preferences['activity_level'] = total24h > 50 ? 'high' : (total24h > 20 ? 'medium' : 'low');

    _preferences['last_updated'] = DateTime.now().toIso8601String();
  }

  /// 生成洞察
  List<Insight> _generateInsights(Map<String, dynamic> trends) {
    final insights = <Insight>[];

    // 基于模式的洞察
    for (final pattern in _patterns) {
      if (pattern.confidence > 0.7) {
        insights.add(Insight(
          type: 'pattern_detected',
          data: {
            'action': pattern.action,
            'context': pattern.context,
            'confidence': pattern.confidence,
            'suggestion': '建议在 ${pattern.context} 场景下自动执行 ${pattern.action}',
          },
        ));
      }
    }

    // 基于活跃度的洞察
    final total24h = trends['total_actions_24h'] as int;
    if (total24h > 100) {
      insights.add(Insight(
        type: 'high_activity',
        data: {
          'message': '今天你非常活跃！',
          'action_count': total24h,
        },
      ));
    }

    // 基于时间的洞察
    final hour = DateTime.now().hour;
    if (hour >= 22 || hour < 6) {
      insights.add(Insight(
        type: 'unusual_time',
        data: {'message': '夜深了，注意休息'},
      ));
    }

    return insights;
  }

  /// 清理旧数据
  void _cleanupOldData() {
    final weekAgo = DateTime.now().subtract(const Duration(days: 7));
    _behaviorLog.removeWhere((b) => b.timestamp.isBefore(weekAgo));

    // 清理低置信度模式
    _patterns.removeWhere((p) => p.confidence < 0.3);
  }

  /// 获取当前偏好
  Map<String, dynamic> getPreferences() => Map.from(_preferences);

  /// 获取学习到的模式
  List<LearnedPattern> getPatterns() => List.from(_patterns);

  /// 获取行为日志
  List<UserBehavior> getBehaviorLog() => List.from(_behaviorLog);

  /// 预测下一个动作
  String? predictNextAction(String context) {
    final matchingPatterns = _patterns.where((p) =>
        p.context == context && p.confidence > 0.5).toList();

    if (matchingPatterns.isEmpty) return null;

    matchingPatterns.sort((a, b) => b.confidence.compareTo(a.confidence));
    return matchingPatterns.first.action;
  }
}

/// 用户行为
class UserBehavior {
  final String action;
  final String context;
  final DateTime timestamp;
  final Map<String, dynamic> metadata;

  UserBehavior({
    required this.action,
    required this.context,
    required this.timestamp,
    this.metadata = const {},
  });
}

/// 学习到的模式
class LearnedPattern {
  final String action;
  final String context;
  double confidence;
  int sampleCount;
  final DateTime learnedAt;

  LearnedPattern({
    required this.action,
    required this.context,
    required this.confidence,
    this.sampleCount = 1,
    DateTime? learnedAt,
  }) : learnedAt = learnedAt ?? DateTime.now();
}

/// 洞察
class Insight {
  final String type;
  final Map<String, dynamic> data;

  Insight({required this.type, required this.data});
}
