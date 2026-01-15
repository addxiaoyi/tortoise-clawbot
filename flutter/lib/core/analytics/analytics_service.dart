import 'package:flutter/foundation.dart';

/// 分析服务 - 跟踪应用使用情况
class AnalyticsService {
  static AnalyticsService? _instance;
  static AnalyticsService get instance => _instance ??= AnalyticsService._();
  AnalyticsService._();

  bool _isEnabled = false;
  final List<AnalyticsEvent> _events = [];

  bool get isEnabled => _isEnabled;

  /// 初始化
  Future<void> initialize({bool enabled = true}) async {
    _isEnabled = enabled;
    debugPrint('AnalyticsService initialized: $enabled');
  }

  /// 记录事件
  void trackEvent(String name, [Map<String, dynamic>? params]) {
    if (!_isEnabled) return;
    
    final event = AnalyticsEvent(
      name: name,
      params: params ?? {},
      timestamp: DateTime.now(),
    );
    _events.add(event);
    debugPrint('Analytics: $name $params');
  }

  /// 记录页面浏览
  void trackScreenView(String screenName) {
    trackEvent('screen_view', {'screen_name': screenName});
  }

  /// 记录错误
  void trackError(String error, [String? stackTrace]) {
    trackEvent('error', {
      'error': error,
      if (stackTrace != null) 'stack_trace': stackTrace,
    });
  }

  /// 记录 AI 请求
  void trackAIRequest(String provider, String model, int tokens) {
    trackEvent('ai_request', {
      'provider': provider,
      'model': model,
      'tokens': tokens,
    });
  }

  /// 记录用户操作
  void trackAction(String action, String category, [Map<String, dynamic>? params]) {
    trackEvent('action', {
      'action': action,
      'category': category,
      ...?params,
    });
  }

  /// 获取事件列表
  List<AnalyticsEvent> getEvents() => List.unmodifiable(_events);

  /// 清空事件
  void clearEvents() => _events.clear();

  /// 启用/禁用
  void setEnabled(bool enabled) {
    _isEnabled = enabled;
  }
}

/// 分析事件
class AnalyticsEvent {
  final String name;
  final Map<String, dynamic> params;
  final DateTime timestamp;

  AnalyticsEvent({
    required this.name,
    required this.params,
    required this.timestamp,
  });

  Map<String, dynamic> toJson() => {
    'name': name,
    'params': params,
    'timestamp': timestamp.toIso8601String(),
  };
}
